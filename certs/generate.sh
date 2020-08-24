#!/bin/bash

rm -f *.key *.csr *.crt *.pem *.srl

SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=ca"

# Generate CA Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout ca.key \
        -subj $SUBJ 

# Generate Req
openssl req \
        -key ca.key \
        -new -out ca.csr \
        -subj $SUBJ

# Generate self signed x509
openssl x509 \
        -signkey ca.key \
        -in ca.csr \
        -req \
        -days 365 -out ca.crt 

TARGET_NAME="target.com"
SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=$TARGET_NAME"
SAN="subjectAltName = DNS:$TARGET_NAME"

# Generate Target Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout target.key \
        -subj $SUBJ

# Generate Req
openssl req \
        -key target.key \
        -new -out target.csr \
        -subj $SUBJ 

# Generate x509 with signed CA
openssl x509 \
        -req \
        -in target.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -CAcreateserial \
        -out target.crt \
        -extensions v3_ca \
        -extfile <(printf "[v3_ca]\nbasicConstraints = CA:FALSE\nkeyUsage = digitalSignature, keyEncipherment\n$SAN")

CLIENT_NAME="client.com"
SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=$CLIENT_NAME"
SAN="subjectAltName = DNS:$CLIENT_NAME"

# Generate Client Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout client.key \
        -subj $SUBJ 

# Generate Req
openssl req \
        -key client.key \
        -new -out client.csr \
        -subj $SUBJ 

# Generate x509 with signed CA
openssl x509 \
        -req \
        -in client.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -out client.crt \
        -extensions v3_ca \
        -extfile <(printf "[v3_ca]\nbasicConstraints = CA:FALSE\nkeyUsage = digitalSignature, keyEncipherment\n$SAN")

echo ""
echo " == Validate Target"
openssl verify -verbose -CAfile ca.crt target.crt
echo ""
echo " == Validate Client"
openssl verify -verbose -CAfile ca.crt client.crt

