#!/bin/bash

rm -f *.key *.csr *.crt *.pem *.srl

SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=ca"
SAN="subjectAltName = DNS:ca"

# Generate CA Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout ca.key \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate Req
openssl req \
        -key ca.key \
        -new -out ca.csr \
        -subj $SUBJ \
        -reqexts SAN \
        -config <(cat /etc/ssl/openssl.cnf \
                <(printf "\n[SAN]\n$SAN"))

# Generate self signed x509
openssl x509 \
        -signkey ca.key \
        -in ca.csr \
        -req \
        -days 365 -out ca.crt 

CN="target.com"
SUBJ="/C=NZ/ST=Test/L=Test/O=Test/OU=Test/CN=$CN"
SAN="subjectAltName=DNS:target.com"

# Generate Target Private Key
openssl req \
        -newkey rsa:2048 \
        -nodes \
        -keyout target.key \
        -subj $SUBJ \
        -extensions EXT \
        -config <( \
                printf "[dn]\nCN=$CN\n[req]\ndistinguished_name = dn\n[EXT]\n%s\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth" "$SAN")

# Generate Req
openssl req \
        -key target.key \
        -new -out target.csr \
        -subj $SUBJ \
        -extensions EXT \
        -config <( \
                printf "[dn]\nCN=$CN\n[req]\ndistinguished_name = dn\n[EXT]\n%s\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth" "$SAN")

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

