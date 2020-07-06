# Generating Certificates

This script is used to generate the required certs for use with both the client and the mock target.

Simply run the script:
```
./generate.sh
```
Then pass the generated certs into the client and server:

## Client Example
```
gnoi_reset \
-target_addr localhost:9399 \
-rollback_os \
-zero_fill \
-ca /path/to/ca.crt \
-cert /path/to/client.crt
-key /path/to/client.key
```

## Target Example
```
gnoi_target \
-zero_fill_unsupported \
-reset_unsupported \
-ca /path/to/ca.crt \
-cert /path/to/client.crt \
-key /path/to/client.key 
```