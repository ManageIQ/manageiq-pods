#!/bin/bash
#
# ManageIQ bare-metal / VM installer for EL10 (CentOS Stream 10, Rocky 10, RHEL 10)
#
# Usage: sudo bash install.sh [--nightly] [--rpm-dir DIR]
#
#   --nightly      Install from the nightly repository instead of the stable release.
#   --rpm-dir DIR  Install ManageIQ RPMs from a local directory of RPMs (e.g. /tmp/rpms).
#                  A local yum repo is created from DIR and takes precedence over the
#                  upstream ManageIQ repos for manageiq-* packages. Useful for testing
#                  custom-built RPMs.
#
# This script installs ManageIQ from the Uhlmann RPM repository.
# It is the shell-script equivalent of the manageiq-appliance container image build.

set -euo pipefail

info()  { echo "==> $*"; }
warn()  { echo "WARN: $*" >&2; }
die()   { echo "ERROR: $*" >&2; exit 1; }

[[ $EUID -eq 0 ]] || die "This script must be run as root (sudo bash $0)"

NIGHTLY=false
RPM_DIR=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --nightly) NIGHTLY=true ;;
    --rpm-dir)
      [[ $# -ge 2 ]] || die "--rpm-dir requires a directory argument"
      RPM_DIR="$2"
      shift
      ;;
    --rpm-dir=*) RPM_DIR="${1#--rpm-dir=}" ;;
    *) die "Unknown argument: $1" ;;
  esac
  shift
done

if [[ -n "${RPM_DIR}" ]]; then
  [[ -d "${RPM_DIR}" ]] || die "RPM directory not found: ${RPM_DIR}"
  RPM_DIR="$(cd "${RPM_DIR}" && pwd)"
  compgen -G "${RPM_DIR}/*.rpm" >/dev/null || die "No .rpm files found in ${RPM_DIR}"
fi

info "Detecting OS..."
if [[ -f /etc/os-release ]]; then
  # shellcheck source=/dev/null
  source /etc/os-release
  OS_ID="${ID:-unknown}"
  OS_VERSION_ID="${VERSION_ID:-unknown}"
  OS_MAJOR="${OS_VERSION_ID%%.*}"
  info "Detected: ${PRETTY_NAME:-$OS_ID $OS_VERSION_ID}"
else
  OS_ID="unknown"
  OS_MAJOR="unknown"
fi

if [[ "$OS_MAJOR" != "10" ]]; then
  warn "This script targets EL10 (CentOS Stream 10, Rocky Linux 10, RHEL 10)."
  warn "Detected major version: ${OS_MAJOR}. Proceeding anyway, but things may break."
fi

RPM_PREFIX="${RPM_PREFIX:-manageiq}"
EPEL_RELEASE_RPM="https://dl.fedoraproject.org/pub/epel/epel-release-latest-10.noarch.rpm"
MIQ_RELEASE_RPM="https://rpm.manageiq.org/release/21-uhlmann/el10/noarch/manageiq-release-21.0-1.el10.noarch.rpm"
MIQ_NIGHTLY_REPO="manageiq-21-uhlmann-nightly"

if [[ "${NIGHTLY}" == "true" ]]; then
  info "Build type: nightly"
else
  info "Build type: release"
fi

info "Configuring dnf defaults..."
dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save

info "Updating system packages..."
dnf -y update

info "Enabling EPEL and CRB..."
if [[ "${OS_ID}" == "rhel" ]]; then
  # epel-release is not in the RHEL repos, so install it from the URL
  dnf -y install "${EPEL_RELEASE_RPM}"
else
  dnf -y install epel-release
fi
/usr/bin/crb enable

info "Installing ManageIQ release repository..."
dnf -y install "${MIQ_RELEASE_RPM}"

if [[ "${NIGHTLY}" == "true" ]]; then
  info "Enabling ManageIQ nightly repository..."
  dnf config-manager --enable "${MIQ_NIGHTLY_REPO}"
fi

# Same as the miq_prepare_local_yum_repo build tool. Must come after the release
# repo RPM is installed so the manageiq-* repos exist for the exclude.
if [[ -n "${RPM_DIR}" ]]; then
  info "Creating local yum repository from ${RPM_DIR}..."
  dnf -y install createrepo_c
  rm -rf "${RPM_DIR}/repodata"
  createrepo_c "${RPM_DIR}"

  cat > /etc/yum.repos.d/local_rpm.repo << EOF
[local-rpm]
name=Local yum repo
baseurl=file://${RPM_DIR}
enabled=1
gpgcheck=0
EOF

  dnf config-manager --setopt='manageiq-*.exclude=manageiq-*' --save
fi

# Pin qpid-proton to the ManageIQ repo; EPEL ships a newer version that is
# incompatible. See https://github.com/ManageIQ/manageiq/blob/b5bf7bee569991d3f1a182a4792fae8cb6265863/Gemfile#L220
info "Pinning qpid-proton to ManageIQ repo..."
dnf config-manager --setopt=epel.exclude=*qpid-proton* --save

info "Installing appliance packages..."
dnf -y install \
  firewalld \
  fuse-libs \
  python3-devel \
  "${RPM_PREFIX}-appliance"

info "Configuring firewall..."
systemctl enable --now firewalld

MIQ_SERVICES=(http https ssh postgresql dhcpv6-client cockpit)

CURRENT_DEFAULT_ZONE=$(firewall-cmd --get-default-zone 2>/dev/null || true)

if [[ -z "${CURRENT_DEFAULT_ZONE}" || "${CURRENT_DEFAULT_ZONE}" == "public" ]]; then
  info "No meaningful default zone found (got: '${CURRENT_DEFAULT_ZONE}'). Creating 'manageiq' zone..."
  if ! firewall-cmd --permanent --info-zone=manageiq &>/dev/null; then
    firewall-cmd --permanent --new-zone=manageiq
  fi
  firewall-cmd --reload
  firewall-cmd --set-default-zone=manageiq
  FIREWALL_ZONE="manageiq"
else
  info "Using existing default firewall zone: '${CURRENT_DEFAULT_ZONE}'"
  FIREWALL_ZONE="${CURRENT_DEFAULT_ZONE}"
fi

for svc in "${MIQ_SERVICES[@]}"; do
  if ! firewall-cmd --permanent --zone="${FIREWALL_ZONE}" --query-service="${svc}" &>/dev/null; then
    firewall-cmd --permanent --zone="${FIREWALL_ZONE}" --add-service="${svc}"
  fi
done

firewall-cmd --reload

info "Configuring systemd services..."

# Disable ipmi if the unit exists
if systemctl list-unit-files ipmi.service &>/dev/null; then
  systemctl disable ipmi
fi

# Disable ctrl-alt-del reboot target
ln -sf /dev/null /etc/systemd/system/ctrl-alt-del.target

# The memcached RPM ships with OPTIONS="-l 127.0.0.1,::1" by default.
# If IPv6 is disabled on this host, binding ::1 will fail (exit code 71 / OSERR).
# Detect IPv6 availability and drop it from the bind list if necessary.
MEMCACHED_SYSCONFIG="/etc/sysconfig/memcached"
if ! ip -6 addr show lo 2>/dev/null | grep -q "::1"; then
  info "IPv6 not available -- removing ::1 from memcached bind address..."
  sed -i 's/OPTIONS="-l 127.0.0.1,::1"/OPTIONS="-l 127.0.0.1"/' "${MEMCACHED_SYSCONFIG}"
fi

# Enable ManageIQ services (cloud-ds-check is intentionally omitted -- cloud-init
# has already completed on a deployed VM)
systemctl enable cockpit.socket evminit memcached miqtop miqvmstat

info "Configuring Cockpit..."
# Allow root to log in via Cockpit
if [[ -f /etc/cockpit/disallowed-users ]]; then
  sed -i 's/root//g' /etc/cockpit/disallowed-users
fi

info "Configuring users..."
usermod -a -G wheel manageiq

mkdir -p /root/.ssh
chmod 700 /root/.ssh
su manageiq -c "mkdir -p -m 700 ~/.ssh"

info "Running ManageIQ setup..."
# /etc/default/evm is provided by the manageiq-appliance RPM and exports
# APPLIANCE_SOURCE_DIRECTORY (and other env vars) needed by manageiq-setup.sh
[[ -f /etc/default/evm ]] || die "/etc/default/evm not found -- was the ${RPM_PREFIX}-appliance RPM installed correctly?"
# shellcheck source=/dev/null
source /etc/default/evm
"${APPLIANCE_SOURCE_DIRECTORY}/manageiq-setup.sh"

info "Installation complete."
info "Run the following to load the aliases into your current session:"
info "  source /etc/profile.d/evm.sh && cd /var/www/miq/vmdb"
info "Then run 'appliance_console' (or 'ap') to configure the appliance."
