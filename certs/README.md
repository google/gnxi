# Generating Certificates

This script is used to generate the required certs for use with both the client and the mock target.

Simply run the script:
```
./generate.sh
```
Then pass the generated certs into the client and server:

## Client example
```
gnoi_cert \
-target_addr localhost:9399 \
-rollback_os \
-zero_fill \
```