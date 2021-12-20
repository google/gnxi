## oc_config_validate

This tools allows Network Operators and Administrators test their OpenConfig-based
configurations of devices, using gNMI.

### Use case

As a Network Operator or Administrator, I need to ensure the configuration of 
devices using OpenConfig models and gNMI protocol works as intended, meaning:

 *  The device replies, without errors and with a valid OpenConfig model,
    to gNMI Get requests, for different xpaths.
 *  The device replies, without errors, to gNMI Set requests, for different
    elements of the model.
 *  The device replies, with the expected and valid OpenConfig updated model,
    to gNMI Set requests, for different elements of the model.

Beyond testing the gNMI connection and transactions, I need to test the
validity and correctness of the OpenConfig models used in the configuration,
looking for:

 *  Syntax or semantic errors in the OpenConfig models issued and replied.

## How to use

### Install

Clone this repo and install the needed dependencies. Preferably, use a virtual environment:

```
cd oc_config_validate
virtualenv --clear venv
source venv/bin/activate
pip3 install -r requirements.txt
python3 -m oc_config_validate --version
```

Optionally, build a Python package with `python3 -m build`

### Demo

Run the script `run_demo.sh` to see a quick demonstration of `oc_config_validate`.

 1. It starts an instance of [gnmmi_target](https://github.com/google/gnxi/blob/master/gnmi_target/gnmi_target.go) with a sample configuration
 1. It runs `oc_config_validate` with a sample test file. 
 1. See `run_demo.sh -h` for additional TLS-connection options to use.

> For the first execution, it first create local TLS certificates.

### Use

 1. Prepare the initial OpenConfig configuration(s) for your device, in JSON format.

    This configuration(s) will be applied to the device before any other operation,
    to bring it to a known initial state.

    Each initial configuration file is sent in a single gNMI Set Update message to the indicated xpath.

    All initial configurations found on the test file and on the command line are applied, sequentially.

 1. Then, write your OpenConfig configuration tests, in YAML format.
 
    [docs/tests.md](docs/tests.md) describes how to create your own tests.

    The are some tests already created at [tests/](tests/).
 
 1. Finally, validate your configurations against a gNMI target using:
 
```
    python3 -m oc_config_validate --target HOSTNAME:PORT \
       -tests TESTS_FILE -results RESULTS_FILE \
       [--username USERNAME] [--password PASSWORD] \
       [--tls_host_override TLS_HOSTNAME | --no_tls ] \
       [-init INIT_CONFIG_FILE -xpath INIT_CONFIG_XPATH] \
       [--verbose] [--log_gnmi] [--stop_on_error]
```
    
    For an example:
    
```
    python3 -m oc_config_validate --target localhost:9339 \
       --username gnmi --password gnmi \
       --tls_host_override target.com \
       -tests tests/example.yaml -results results/example.json \
       -init init_configs/example.json -xpath "/system/" \
       --log_gnmi
```

 1. Expect an output similar to:
 
```
    INFO(__main__.py:275):Read tests file 'demo/example.yaml': 16 tests to run
    INFO(__main__.py:283):Using TLS for gNMI connection
    INFO(__main__.py:293):Testing gNMI Target localhost:9339.
    INFO(target.py:155):Obtaining Root CA certificate from Target
    INFO(__main__.py:301):Initial OpenConfig 'demo/init_config.json' applied on gNMI xpath /system/config
    INFO(runner.py:58):Running Test 'Get System Config'
    INFO(runner.py:62):Test 'Get System Config' took 1.031494 msecs: PASS
    
    INFO(runner.py:58):Running Test 'Check System Config'
    INFO(testbase.py:277):test0200 (oc_config_validate.testcases.get.GetJsonCheck) - FAIL:
    
     Traceback (most recent call last):
      File "/usr/local/google/home/itamayo/oc_config_validate/oc_config_validate/testcases/get.py", line 83, in test0200
        self.assertJsonModel(resp_val, self.model,
      File "/usr/local/google/home/itamayo/oc_config_validate/oc_config_validate/testbase.py", line 233, in assertJsonModel
        self.assertTrue(match, "%s: %s" % (msg, err))
    AssertionError: False is not true : Get response JSON does not match model: JSON object contained a key that did not exist (hostname)
    
    INFO(runner.py:62):Test 'Check System Config' took 4.437500 msecs: FAIL
    
    INFO(runner.py:58):Running Test 'Check AAA State'
    INFO(runner.py:62):Test 'Check AAA State' took 1.718262 msecs: PASS
    ...
    INFO(runner.py:58):Running Test 'Set timezone with a valid Json blob'
    INFO(runner.py:62):Test 'Set timezone with a valid Json blob' took 60.581787 msecs: PASS
    
    INFO(runner.py:58):Running Test 'Set timezone with a valid Json blob, and check it is Zurich'
    INFO(runner.py:62):Test 'Set timezone with a valid Json blob, and check it is Zurich' took 24.644043 msecs: PASS
    
    INFO(runner.py:58):Running Test 'Bad Set Clock Check State'
    WARNING(api.py:40):False is not true : key openconfig-system:timezone-name: Europe/Stockholm != Europe/Paris, retrying in 10 seconds...
    WARNING(api.py:40):False is not true : key openconfig-system:timezone-name: Europe/Stockholm != Europe/Paris, retrying in 10 seconds...
    WARNING(api.py:40):False is not true : key openconfig-system:timezone-name: Europe/Stockholm != Europe/Paris, retrying in 10 seconds...
    WARNING(api.py:40):False is not true : key openconfig-system:timezone-name: Europe/Stockholm != Europe/Paris, retrying in 10 seconds...
    INFO(testbase.py:277):test0300 (oc_config_validate.testcases.config_state.SetConfigCheckState) - FAIL:
    
     Traceback (most recent call last):
      File "/usr/local/google/home/itamayo/oc_config_validate/venv/lib/python3.9/site-packages/decorator.py", line 232, in fun
        return caller(func, *(extras + args), **kw)
      File "/usr/local/google/home/itamayo/oc_config_validate/venv/lib/python3.9/site-packages/retry/api.py", line 73, in retry_decorator
        return __retry_internal(partial(f, *args, **kwargs), exceptions, tries, delay, max_delay, backoff, jitter,
      File "/usr/local/google/home/itamayo/oc_config_validate/venv/lib/python3.9/site-packages/retry/api.py", line 33, in __retry_internal
        return f()
      File "/usr/local/google/home/itamayo/oc_config_validate/oc_config_validate/testcases/config_state.py", line 81, in test0300
        self.assertTrue(cmp, diff)
    AssertionError: False is not true : key openconfig-system:timezone-name: Europe/Stockholm != Europe/Paris
    
    INFO(runner.py:62):Test 'Bad Set Clock Check State' took 40082.661377 msecs: FAIL
    
    INFO(__main__.py:310):Results Summary: PASS 10, FAIL 6
    INFO(__main__.py:315):Test results written to demo/results.json
```

#### gNMI TLS connection options

By default, *oc_config_validate* will use TLS for the gNMI connection, will
validate the hostname presented by the Target in its certificate and will fetch
the Root CA certificate from the Target.

 *  Use the `--tls_host_override` flag to provide the hostname value present
    in the Target's certificate.
 *  Use the `--root_cat` flag to provide the Root CA certificate file to use.
 *  Use the `-key` and `-cert` flags to provide the TLS key and certificate 
    files that the client will present to the Target.

In case of errors, use the `--debug` flag to help understand the underlying TLS issue, if any.

 > With case, use the `--no_tls` flag not to use TLS for the gNMI connection. 
 >
 > Warning: All communication will be in plain text.

#### OpenConfig Models

`oc_config_validate` includes the Python bindings of some OpenConfig models, in the package `oc_config_validate.models`.

 *  To update the bindings to the latest model versions, run `update_models.sh`
 *  To add more OpenConfig models, update `update_models.sh`

The model bindings are used in tests that validate the gNMI JSON payloads against the expected OpenConfig model. See [docs/tests](docs/tests.md) for more details.

Run `python3 -m oc_config_validate -models` to get a list of the versions (revisions) of the OC models used. These are also added to the results file.

## Copyright

Contains code from https://github.com/google/gnxi/tree/master/gnmi_cli_py and
https://github.com/openconfig/gnmitest/blob/master/schemas/openconfig, under
Apache License, Version 2.0 license.
