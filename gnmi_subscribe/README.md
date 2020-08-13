# gNMI Subscribe Client

A simple shell binary that performs gNMI Subscribe client operations against a gNMI target.

## gNMI Subscribe Operations

There are 3 subscription modes that can be used:

* **ONCE** is enabled using `-once`. In this mode, the target creates the relevant update messages, transmits them, and then closes the RPC.

* **POLL** is enabled using `-poll`. In this mode, the target generates and transmits updates on demand.

* **STREAM** is enabled when neither of the flags above are set. In this mode, data can be streamed:
	* Periodically, using `-sample_interval`. 
		* `-heartbeat_interval` can be used, which specifies an interval, during which the target is forced to generate a telemetry update.
	* On change, using `-stream_on_change`
		* `-suppress_redundant` can be used, which means that updates are only generated for those individual leaf nodes in the subscription that have changed.
        * `-heartbeat_interval` can be used, which specifies an interval, during which the target is forced to generate a telemetry update.
	* If neither flag is set then the target determines the best subscription type.
 



## Install

```
go get github.com/google/gnxi/gnmi_subscribe
go install github.com/google/gnxi/gnmi_subscribe
```

## Run 
```
./gnmi_subscribe \
    -target_addr localhost:9399 \
    -target_name target.com \
    -ca ca.crt \
    -key client.key \
    -cert client.crt \
    -sample_interval 10 \
    -encoding JSON
```
