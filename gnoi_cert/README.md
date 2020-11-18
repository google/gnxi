# gNOI Certificate Management Client

A simple shell binary that performs Certificate Management client operations
against a gNOI Target.

## Certificates

Only the Root certificate and private key are required for this client. The
client will:

* generate a client certificate for establishing the connection to the Target
* sign target signing requests for installing or rotating certificates on the Target

The client certificates can also be provided to establish the connection to the target and will be used instead.

## gNOI Certificate Management operations

*   `-op provision` installs a Certificate and CA Bundle on a Target that is in
    bootstrapping mode, accepting encrypted TLS connections;
*   `-op install` installs a certificate and CA Bundle on a Target where it was
    already provisioned. Connections are authenticated using TLS;
*   `-op rotate` rotates a certificate on a provisioned Target;
*   `-op revoke` revokes a certificate on a provisioned Target;
*   `-op get` gets all installed certificate on a provisioned Target;
*   `-op check` check if a provisioned target can generate CSRs;

## Install

```
go get github.com/google/gnxi/gnoi_cert
go install github.com/google/gnxi/gnoi_cert
```

## Run

```
./gnoi_cert \
  -target_addr localhost:9339 \
  -target_name target.com \
  -ca_key ca.key \
  -ca ca.crt \
  -op provision \
  -cert_id provision_cert
```
