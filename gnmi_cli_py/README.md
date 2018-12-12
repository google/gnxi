# gNMI CLI

A simple Python script that performs for interacting with gNMI Targets.

## Dependencies

```
sudo pip install --no-binary=protobuf -I grpcio-tools==1.15.0
```

## Run

```
python py_gnmicli.py -m get -t <gnmi_target_addr> -p <port> -x <xpath> -u <user> -w <password> [Optional] -rcert <CA certificate>

Example:
python py_gnmicli.py -m get -t example.net -x /access-points/access-point[hostname=ap-1]/ -u admin -w admin -p 8080 -o openconfig.example.com
```

Note on certificates. If no root certificate is supplied (option: -rcert), one will be automatically downloaded from the Target at runtime. If the gNMI Target is utilizing a self-signed certificate it may also be required to supply the hostname utilized in the certificate (option: --host_override)
