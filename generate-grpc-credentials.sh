#!/usr/bin/env bash

if [[ -e "credentials" ]]; then
  echo "Error: 'credentials' already exists"
  exit 1
fi

mkdir -p "credentials"

pushd "credentials" >/dev/null || exit 1
echo " :: Generating CA"
openssl genpkey -algorithm ED25519 -out ca.key
openssl req -new -x509 -key ca.key -sha512 -subj "/C=US/ST=WA/O=Jared Allard" -days 3650 -out ca.crt
echo " :: Generating Server Certificate"
openssl genpkey -algorithm ED25519 -out service.key
openssl req -new -key service.key -out service.csr
openssl x509 -req -in service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out service.pem -days 3650 -sha512 -extensions req_ext
popd >/dev/null || exit 
