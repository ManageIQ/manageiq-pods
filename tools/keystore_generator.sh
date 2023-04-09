#!/usr/bin/env bash

set -e

echo "Enter CA cert path:"
read CA_CERT_PATH

echo "Enter CA key path:"
read CA_KEY_PATH

echo "Set Keystore password:"
read KEYSTORE_PASS

if [ ! -e "$CA_CERT_PATH" ] || [ ! -e "$CA_KEY_PATH" ]; then
  echo "CA does not exist, please provide the corrrect paths in CA_CERT_PATH and CA_KEY_PATH"
  exit 1
fi

if [ -z "$KEYSTORE_PASS" ]; then
  echo "Please provide a keystore password in KEYSTORE_PASS"
  exit 1
fi

# Generate truststore containing CA
keytool -keystore ./kafka.truststore.jks \
  -alias CARoot -import -file $CA_CERT_PATH \
  -noprompt -dname "CN=kafka" -keypass $KEYSTORE_PASS -storepass $KEYSTORE_PASS

# Generate keystore
keytool -keystore ./kafka.keystore.jks \
  -alias kafka -validity 365 -genkey -keyalg RSA \
  -noprompt -dname "CN=kafka" -keypass $KEYSTORE_PASS -storepass $KEYSTORE_PASS

# Create certificate signing request to keystore
keytool -keystore ./kafka.keystore.jks -alias kafka \
  -certreq -file cert-sign-req -keypass $KEYSTORE_PASS -storepass $KEYSTORE_PASS

# Sign keystore's certificate using CA key
openssl x509 -req -CA $CA_CERT_PATH -CAkey $CA_KEY_PATH \
  -in ./cert-sign-req -out cert-sign \
  -days 365 -CAcreateserial

# Import CA into keystore
keytool -keystore ./kafka.keystore.jks -alias CARoot \
  -import -file $CA_CERT_PATH -keypass $KEYSTORE_PASS -storepass $KEYSTORE_PASS -noprompt

# Import signed certificate back into keystore
keytool -keystore ./kafka.keystore.jks -alias kafka -import \
  -file ./cert-sign -keypass $KEYSTORE_PASS -storepass $KEYSTORE_PASS

echo "Truststore and keystore have been successfully created"
