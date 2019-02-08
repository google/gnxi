# gNMI CLI

A simple Python script that performs interactions with gNMI Targets.

## Dependencies

```
pip install -r requirements.txt
```
You may also need to pip install setuptools.

## Usage Examples
gNMI GetRequests. Substitute where applicable.
```
python py_gnmicli.py -m get -t <gnmi_target_addr> -p <port> -x <xpath> -u <user> -w <password> [Optional] -rcert <CA certificate>
```
gNMI GetRequest for an Access Point target, who's hostname is "ap-1", with an xpath of root for access-points model:
```
python py_gnmicli.py -m get -t example.net -x /access-points/access-point[hostname=ap-1]/ -u admin -w admin -p 8080 -o openconfig.example.com
```
gNMI GetRequest for an Access Point target, who's hostname is "ap-1", with an xpath of the config container of Radio with ID 0:
```
python py_gnmicli.py -m get -t example.net -x /access-points/access-point[hostname=ap-1]/radios/radio[id=0]/config -u admin -w admin -p 8080 -o openconfig.example.com
```
gNMI SetRequest Replace for an Access Point target, who's hostname is "ap-1", with an xpath of the channel config leaf of Radio with ID 0 (This would assign channel 165 to this Radio):
```
python py_gnmicli.py -t example.net -p 443 -m set-replace -x /access-points/access-point[hostname=test-ap1]/radios/radio[id=0]/config/channel -o openconfig.example.com -user admin -pass admi -rcert ca.cert.pem -val 165
```
The above SetRequest Replace would output the following to stdout:
```
Performing SetRequest Replace, encoding=JSON_IETF  to  openconfig.example.com with the following gNMI Path
 -------------------------
 elem {
  name: "access-points"
}
elem {
  name: "access-point"
  key {
    key: "hostname"
    value: "test-ap1"
  }
}
elem {
  name: "radios"
}
elem {
  name: "radio"
  key {
    key: "id"
    value: "0"
  }
}
elem {
  name: "config"
}
elem {
  name: "channel"
}

The SetRequest response is below
-------------------------
 response {
  path {
    elem {
      name: "access-points"
    }
    elem {
      name: "access-point"
      key {
        key: "hostname"
        value: "test-ap1"
      }
    }
    elem {
      name: "radios"
    }
    elem {
      name: "radio"
      key {
        key: "id"
        value: "0"
      }
    }
    elem {
      name: "config"
    }
    elem {
      name: "channel"
    }
  }
  op: REPLACE
}
```

Note on certificates. If no root certificate is supplied (option: -rcert), it will be automatically downloaded from the Target at runtime. If the gNMI Target is utilizing a self-signed certificate it may also be required to supply the hostname utilized in the certificate (option: --host_override)

For example:
```
python py_gnmicli.py -t target1.example.com -p 443 -m get -x /access-points/access-point[hostname=test-ap1]/radios/radio[id=0]/config -o openconfig.mojonetworks.com -user admin -pass admin -rcert ca.cert.pem
```

### Notable Options
* The default output of a GetRequest is to dump the value as JSON. This can be changed with the -f flag.
* Pay special attention when utilizing a JSON file as the val when performing SetRequests. It MUST be preceded by an '@'; else it is assumed that you are providing a leaf value directly.
* The host_override (-o) option is most likely needed, if the Target is utilizing a self-signed certificate (unless the root CA is trusted on the host machine).
* Use the debug flag (-d) when troubleshooting/reporting gRPC errors.
