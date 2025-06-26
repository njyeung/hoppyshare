#!/bin/bash
set -e

generate_cert () {
  CN=$1
  openssl genrsa -out "$CN.key" 2048
  openssl req -new -key "$CN.key" -out "$CN.csr" -subj "/CN=$CN"
  openssl x509 -req -in "$CN.csr" -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out "$CN.crt" -days 365 -sha256 \
    -extfile openssl.cnf -extensions v3_req
}

generate_cert lambda
generate_cert watchdog
generate_cert server

