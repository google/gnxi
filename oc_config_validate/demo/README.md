# Demo

Run the script `demo/run_demo.sh` to see a quick demonstration of `oc_config_validate`.

 1. It starts an instance of [gnmmi_target](../gnmi_target) with a sample configuration.
 1. It runs `oc_config_validate` with a sample [test file](demo/tests.yaml).
 1. It writes the results to [a JSON file](demo/results.json).

Run `demo/run_demo.sh -h` for additional options to use.

> It uses the local TLS certificates from [gnxi/certs/](../certs/).

> `gnmmi_target` needs `golang>=1.17` runtime, install it first.
 
