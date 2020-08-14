#!/bin/bash

rm -f *.key *.csr *.crt *.pem *.srl

CN="ca"
SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=$CN"
SAN="subjectAltName=DNS:$CN"

# Generate Req
openssl req -x509 \
        -out ca.crt \
        -keyout ca.key \
        -newkey rsa:2048 -nodes -sha256 \
        -nodes \
        -subj $SUBJ \
        -extensions EXT \
        -config <( \
                printf "[dn]\nCN=$CN\n[req]\ndistinguished_name = dn\n[EXT]\n%s\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth" "$SAN")


SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=target.com"
SAN="subjectAltName = DNS:target.com"

# Generate Target Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout target.key \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate Req
openssl req \
        -key target.key \
        -new -out target.csr \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate x509 with signed CA
openssl x509 \
        -req \
        -in target.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -CAcreateserial \
        -out target.crt

SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=client.com"
SAN="subjectAltName = DNS:client.com"

# Generate Client Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout client.key \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate Req
openssl req \
        -key client.key \
        -new -out client.csr \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate x509 with signed CA
openssl x509 \
        -req \
        -in client.csr \
        -CA ca.crt \
        -CAkey ca.key \
        -out client.crt

echo ""
echo " == Validate Target"
openssl verify -verbose -CAfile ca.crt target.crt
echo ""
echo " == Validate Client"
openssl verify -verbose -CAfile ca.crt client.crt

