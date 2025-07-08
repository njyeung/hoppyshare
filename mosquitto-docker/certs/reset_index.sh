#!/bin/bash

set -e

echo "Resetting OpenSSL index, serial, newcerts, and CRL..."

# Define working directory (current folder)
CERT_DIR="$(dirname "$0")"
cd "$CERT_DIR"

# Files to reset
rm -f index.txt index.txt.attr index.txt.old index.txt.attr.old
rm -f ca-crl.pem
rm -rf newcerts
mkdir -p newcerts

# Recreate fresh files
touch index.txt
echo 1000 > serial

# Regenerate empty CRL
if [ -f openssl.cnf ]; then
  openssl ca -gencrl -out ca-crl.pem -config openssl.cnf
  echo "ca-crl.pem regenerated"
else
  echo "⚠️  openssl.cnf not found — skipping CRL regeneration"
fi

echo "Reset complete:"
echo "- index.txt recreated"
echo "- serial reset to 1000"
echo "- newcerts/ cleared"
echo "- ca-crl.pem generated"

