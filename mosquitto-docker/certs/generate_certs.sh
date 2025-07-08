#!/bin/bash
set -e

generate_cert () {
  CN=$1

  echo "Generating cert for CN=$CN"

  openssl genrsa -out "$CN.key" 2048
  openssl req -new -key "$CN.key" -out "$CN.csr" -subj "/CN=$CN"

  openssl ca -batch \
    -config openssl.cnf \
    -in "$CN.csr" \
    -out "$CN.crt" \
    -keyfile ca.key \
    -cert ca.crt \
    -extensions v3_req
}

generate_cert lambda
generate_cert watchdog
generate_cert server
