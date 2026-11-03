#!/usr/bin/env bash
#
# setup_kafka_fips.sh - Configure Apache Kafka on RHEL 10 in FIPS Mode (SASL/PLAIN over SASL_SSL)
#
set -euo pipefail

# Ensure script is run as root
if [[ $EUID -ne 0 ]]; then
    echo "ERROR: This script must be run as root." >&2
    exit 1
fi

# Configuration Variables (override via environment if desired)
KAFKA_HOST="${KAFKA_HOST:-$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="src") {print $(i+1); exit}}' || hostname -I | awk '{print $1}')}"
KAFKA_HOSTNAME="${KAFKA_HOSTNAME:-$(hostname -f)}"
KAFKA_PORT="${KAFKA_PORT:-9092}"

# Passwords must be supplied by the caller (no insecure defaults)
[[ -n "${STOREPASS:-}"     ]] || { echo "ERROR: STOREPASS must be set"     >&2; exit 1; }
[[ -n "${ADMIN_USER:-}"    ]] || { echo "ERROR: ADMIN_USER must be set"    >&2; exit 1; }
[[ -n "${ADMIN_PASS:-}"    ]] || { echo "ERROR: ADMIN_PASS must be set"    >&2; exit 1; }

KEYSTORE_DIR="/etc/kafka/keystore"
CONFIG_DIR="/etc/kafka/config"

echo "=== Starting Kafka FIPS Setup ==="
echo "Host IP:   ${KAFKA_HOST}"
echo "Hostname:  ${KAFKA_HOSTNAME}"
echo "Port:      ${KAFKA_PORT}"

# 1. Configure Java Security Providers override for Kafka (without modifying system java.security)
echo "==> Configuring Kafka Java security properties in ${CONFIG_DIR}/kafka.security..."
mkdir -p "${CONFIG_DIR}"
cat << 'EOF' > "${CONFIG_DIR}/kafka.security"
fips.provider.8=SunSASL
fips.provider.9=JdkSASL
fips.provider.10=SunJGSS
EOF
chown kafka:kafka "${CONFIG_DIR}/kafka.security"
chmod 644 "${CONFIG_DIR}/kafka.security"

echo "==> Configuring systemd service override for kafka.service..."
mkdir -p /etc/systemd/system/kafka.service.d
cat << 'EOF' > /etc/systemd/system/kafka.service.d/override.conf
[Service]
Environment="KAFKA_OPTS=-Djava.security.auth.login.config=/etc/kafka/config/kafka_server_jaas.conf -Djava.security.properties=/etc/kafka/config/kafka.security"
EOF
systemctl daemon-reload

# 2. Create Keystore directory and generate PKCS12 Keystore & Truststore
echo "==> Generating FIPS-compliant PKCS12 keystore and truststore..."
mkdir -p "${KEYSTORE_DIR}"
rm -f "${KEYSTORE_DIR}"/*

SAN_EXT="SAN=DNS:${KAFKA_HOSTNAME},DNS:localhost,IP:${KAFKA_HOST},IP:127.0.0.1"

# Generate Keypair in PKCS12 keystore
keytool -genkeypair \
    -alias kafka-broker \
    -keyalg RSA \
    -keysize 2048 \
    -sigalg SHA256withRSA \
    -storetype PKCS12 \
    -keystore "${KEYSTORE_DIR}/server.keystore.p12" \
    -storepass "${STOREPASS}" \
    -keypass "${STOREPASS}" \
    -dname "CN=${KAFKA_HOST}, OU=Kafka, O=ManageIQ, C=US" \
    -ext "${SAN_EXT}" \
    -validity 365

# Export public cert in PEM format
keytool -exportcert \
    -rfc \
    -alias kafka-broker \
    -keystore "${KEYSTORE_DIR}/server.keystore.p12" \
    -storepass "${STOREPASS}" \
    -file "${KEYSTORE_DIR}/server-cert.pem"

# Import public cert into PKCS12 truststore
keytool -importcert \
    -alias kafka-broker \
    -file "${KEYSTORE_DIR}/server-cert.pem" \
    -keystore "${KEYSTORE_DIR}/server.truststore.p12" \
    -storetype PKCS12 \
    -storepass "${STOREPASS}" \
    -noprompt

# Set secure permissions
chown -R kafka:kafka "${KEYSTORE_DIR}"
chmod 755 "${KEYSTORE_DIR}"
chmod 644 "${KEYSTORE_DIR}"/*.pem
chmod 600 "${KEYSTORE_DIR}"/*.p12

# 3. Create Server JAAS configuration
echo "==> Configuring JAAS in ${CONFIG_DIR}/kafka_server_jaas.conf..."
cat << EOF > "${CONFIG_DIR}/kafka_server_jaas.conf"
KafkaServer {
    org.apache.kafka.common.security.plain.PlainLoginModule required
    username="${ADMIN_USER}"
    password="${ADMIN_PASS}"
    user_${ADMIN_USER}="${ADMIN_PASS}";
};

KafkaClient {
    org.apache.kafka.common.security.plain.PlainLoginModule required
    username="${ADMIN_USER}"
    password="${ADMIN_PASS}";
};
EOF

chown kafka:kafka "${CONFIG_DIR}/kafka_server_jaas.conf"
chmod 600 "${CONFIG_DIR}/kafka_server_jaas.conf"

# 4. Configure /etc/kafka/config/server.properties
echo "==> Configuring ${CONFIG_DIR}/server.properties..."
cat << EOF > "${CONFIG_DIR}/server.properties"
broker.id=0
num.network.threads=3
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
log.dirs=/var/lib/kafka/logs
num.partitions=1
num.recovery.threads.per.data.dir=1
offsets.topic.replication.factor=1
transaction.state.log.replication.factor=1
transaction.state.log.min.isr=1
log.retention.hours=168
log.segment.bytes=1073741824
log.retention.check.interval.ms=300000
zookeeper.connect=localhost:2181
zookeeper.connection.timeout.ms=18000
group.initial.rebalance.delay.ms=0

# Listeners
listeners=SASL_SSL://${KAFKA_HOST}:${KAFKA_PORT}
advertised.listeners=SASL_SSL://${KAFKA_HOST}:${KAFKA_PORT}
listener.security.protocol.map=SASL_SSL:SASL_SSL

# SASL Configuration
sasl.enabled.mechanisms=PLAIN
sasl.mechanism.inter.broker.protocol=PLAIN
security.inter.broker.protocol=SASL_SSL

# TLS / FIPS Keystore & Truststore
ssl.keystore.location=${KEYSTORE_DIR}/server.keystore.p12
ssl.keystore.password=${STOREPASS}
ssl.key.password=${STOREPASS}
ssl.keystore.type=PKCS12

ssl.truststore.location=${KEYSTORE_DIR}/server.truststore.p12
ssl.truststore.password=${STOREPASS}
ssl.truststore.type=PKCS12

ssl.enabled.protocols=TLSv1.2,TLSv1.3
ssl.protocol=TLSv1.3
ssl.client.auth=none
EOF

chown kafka:kafka "${CONFIG_DIR}/server.properties"

# 5. Configure /etc/kafka/config/client.properties
echo "==> Configuring ${CONFIG_DIR}/client.properties..."
cat << EOF > "${CONFIG_DIR}/client.properties"
bootstrap.servers=${KAFKA_HOST}:${KAFKA_PORT}
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required \\
    username="${ADMIN_USER}" \\
    password="${ADMIN_PASS}";

# TLS / Truststore Configuration
ssl.truststore.location=${KEYSTORE_DIR}/server.truststore.p12
ssl.truststore.password=${STOREPASS}
ssl.truststore.type=PKCS12
ssl.enabled.protocols=TLSv1.2,TLSv1.3
EOF

chown kafka:kafka "${CONFIG_DIR}/client.properties"
chmod 644 "${CONFIG_DIR}/client.properties"

# 6. Update manageiq-messaging-ready for FIPS
if [[ -f /usr/bin/manageiq-messaging-ready ]]; then
    echo "==> Updating /usr/bin/manageiq-messaging-ready for FIPS..."
    sed -i \
        -e "/KAFKA_OPTS=/!s|\`/opt/kafka/bin/kafka-topics\.sh|\`KAFKA_OPTS='-Djava.security.properties=${CONFIG_DIR}/kafka.security' /opt/kafka/bin/kafka-topics.sh|" \
        -e "s|/opt/kafka/config/client\.properties|${CONFIG_DIR}/client.properties|" \
        /usr/bin/manageiq-messaging-ready
fi

# 7. Configure ManageIQ messaging.yml
if [[ -d /var/www/miq/vmdb ]]; then
    echo "==> Configuring /var/www/miq/vmdb/config/messaging.yml..."
    mkdir -p /var/www/miq/vmdb/config
    cat << EOF > /var/www/miq/vmdb/config/messaging.yml
production:
  host: ${KAFKA_HOST}
  port: ${KAFKA_PORT}
  protocol: Kafka
  encoding: json
  username: ${ADMIN_USER}
  password: ${ADMIN_PASS}
  ssl: true
  ca_file: ${KEYSTORE_DIR}/server-cert.pem
EOF
fi

# 8. Configure Firewall
if systemctl is-active --quiet firewalld; then
    echo "==> Configuring firewalld for port ${KAFKA_PORT}/tcp..."
    firewall-cmd --permanent --add-port="${KAFKA_PORT}/tcp" >/dev/null 2>&1 || true
    firewall-cmd --reload >/dev/null 2>&1 || true
fi

# 9. Start Services
echo "==> Starting ZooKeeper and Kafka services..."
systemctl enable --now zookeeper
systemctl restart zookeeper
sleep 2

systemctl enable --now kafka
systemctl restart kafka
sleep 3

# 10. Check Status
if ! systemctl is-active --quiet kafka; then
    echo "ERROR: Kafka service failed to start. Check 'journalctl -u kafka -xe' for details." >&2
    exit 1
fi
