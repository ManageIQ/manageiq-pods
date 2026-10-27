#!/bin/bash
#
# ManageIQ bare-metal / VM installer for EL10 (CentOS Stream 10, Rocky 10, RHEL 10)
#
# Usage: sudo bash install.sh
#
# This script installs ManageIQ from the Uhlmann nightly RPM repository.
# It is the shell-script equivalent of the manageiq-appliance container image build.

set -euo pipefail

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

info()  { echo "==> $*"; }
warn()  { echo "WARN: $*" >&2; }
die()   { echo "ERROR: $*" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Preflight checks
# ---------------------------------------------------------------------------

[[ $EUID -eq 0 ]] || die "This script must be run as root (sudo bash $0)"

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

# ---------------------------------------------------------------------------
# Variables
# ---------------------------------------------------------------------------

RPM_PREFIX="${RPM_PREFIX:-manageiq}"
MIQ_RELEASE_RPM="https://rpm.manageiq.org/release/21-uhlmann/el10/noarch/manageiq-release-21.0-1.el10.noarch.rpm"
MIQ_NIGHTLY_REPO="manageiq-21-uhlmann-nightly"

# ---------------------------------------------------------------------------
# Step 1: Configure dnf defaults
# ---------------------------------------------------------------------------

info "Configuring dnf defaults..."
dnf config-manager --setopt=tsflags=nodocs --setopt=install_weak_deps=False --save

# ---------------------------------------------------------------------------
# Step 2: Update system
# ---------------------------------------------------------------------------

info "Updating system packages..."
dnf -y update

# ---------------------------------------------------------------------------
# Step 3: Enable EPEL and CRB
# ---------------------------------------------------------------------------

info "Enabling EPEL and CRB..."
dnf -y install epel-release
/usr/bin/crb enable

# ---------------------------------------------------------------------------
# Step 4: Install ManageIQ release repo and enable nightly
# ---------------------------------------------------------------------------

info "Installing ManageIQ release repository..."
dnf -y install "${MIQ_RELEASE_RPM}"

info "Enabling ManageIQ nightly repository..."
dnf config-manager --enable "${MIQ_NIGHTLY_REPO}"

# Pin qpid-proton to the ManageIQ repo; EPEL ships a newer version that is
# incompatible. See https://github.com/ManageIQ/manageiq/blob/b5bf7bee569991d3f1a182a4792fae8cb6265863/Gemfile#L220
info "Pinning qpid-proton to ManageIQ repo..."
dnf config-manager --setopt=epel.exclude=*qpid-proton* --save

# ---------------------------------------------------------------------------
# Step 5: Install ManageIQ packages
# ---------------------------------------------------------------------------

info "Installing ManageIQ appliance packages..."
dnf -y install \
  firewalld \
  fuse-libs \
  python3-devel \
  "${RPM_PREFIX}-appliance"

# ---------------------------------------------------------------------------
# Step 6: Configure firewall
# ---------------------------------------------------------------------------

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

# ---------------------------------------------------------------------------
# Step 7: Configure systemd services
# ---------------------------------------------------------------------------

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

# ---------------------------------------------------------------------------
# Step 8: Cockpit configuration
# ---------------------------------------------------------------------------

info "Configuring Cockpit..."
# Allow root to log in via Cockpit
if [[ -f /etc/cockpit/disallowed-users ]]; then
  sed -i 's/root//g' /etc/cockpit/disallowed-users
fi

# ---------------------------------------------------------------------------
# Step 9: User and SSH setup
# ---------------------------------------------------------------------------

info "Configuring users..."
usermod -a -G wheel manageiq

mkdir -p /root/.ssh
chmod 700 /root/.ssh
su manageiq -c "mkdir -p -m 700 ~/.ssh"

# ---------------------------------------------------------------------------
# Step 10: Run ManageIQ setup
# ---------------------------------------------------------------------------

info "Running ManageIQ setup..."
# /etc/default/evm is provided by the manageiq-appliance RPM and exports
# APPLIANCE_SOURCE_DIRECTORY (and other env vars) needed by manageiq-setup.sh
[[ -f /etc/default/evm ]] || die "/etc/default/evm not found -- was the ${RPM_PREFIX}-appliance RPM installed correctly?"
# shellcheck source=/dev/null
source /etc/default/evm
"${APPLIANCE_SOURCE_DIRECTORY}/manageiq-setup.sh"

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------

info "ManageIQ installation complete."
info "Run the following to load ManageIQ aliases into your current session:"
info "  source /etc/profile.d/evm.sh && cd /var/www/miq/vmdb"
info "Then run 'appliance_console' (or 'ap') to configure the appliance."
