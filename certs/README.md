# Generating Certificates

This script is used to generate the required certs for use with both the client and the mock target.

 * CA.crt -> Certificate Authority certificate (a self-signed certificate that signs all the others).
 * CA.key -> CA Private key (should be kept very secret).
 * client.crt/target.crt -> A client or target certificate that was signed by the CA.
 * client.key/target.key -> The private key corresponding to the client or target.

Simply run the script:
```
./generate.sh
```
Then pass the generated certs into the client and server:

## Client Example
```
gnoi_reset \
 <...>
-ca /path/to/ca.crt \
-cert /path/to/client.crt
-key /path/to/client.key
```

## Target Example
```
gnoi_target \
 <...>
-ca /path/to/ca.crt \
-cert /path/to/target.crt \
-key /path/to/target.key 
```
