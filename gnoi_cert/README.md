# gNOI Certificate Management Client

A simple shell binary that performs Certificate Management client operations
against a gNOI Target.

## Certificates

Only the Root certificate and private key are required for this client. The
client will generate a client certificate and sign target signing requests
(CSRs) with it. However this is not recommended for operations where an external
signing authority is recommended.

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
gnoi_cert \
  -target_addr localhost:10161 \
  -target_name hostname.com \
  -key ca.key \
  -ca ca.crt \
  -op provision \
  -cert_id provision_cert \
  -alsologtostderr
```
