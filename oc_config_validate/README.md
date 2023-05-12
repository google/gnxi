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

1. Install Python3.9 or later.

1. Preferably, use a virtual environment:

    ```
    virtualenv --clear venv
    source venv/bin/activate
    ```

1. Install from Pypi

    ```
    pip3 install oc-config-validate
    ```

1. Check it is all working as expected:

    ```
    python3 -m oc_config_validate --version
    python3 -m oc_config_validate -models
    ```

### Prepare

 1. Write the initial OpenConfig configuration(s) for your device, in JSON format.

    This configuration(s) will be applied to the device before any other operation,
    to bring it to a known initial state.

 1. Write your OpenConfig configuration tests, in YAML format.

    See [here](https://github.com/google/gnxi/blob/master/oc_config_validate/docs/tests.md) to create your OpenConfig tests.

    The YAML file is organized as follows:

    ```
    !TestContext

    # Informative text about this test run.
    description: [:string]

    # Optional list of text labels, passed to the Results. Can be used for tagging and filtering.
    labels:
    - [:string]

    # Target to connect to. Overridden by command arguments.
    target:
      !Target
      [:object]

    # Optional list of initial configurations and xpaths to apply before the tests.
    init_configs:
    - !InitConfig
      [:object]

    # List of Tests to run.
    tests:
    - !TestCase
      # Short description of what is being tested.
      name: [:string]
      # A valid class in the oc_config_validate.testcases module
      class_name: [:string]
      # Optional list of arguments required by the class.
      args:
        [:string]: [:any]

    ```

    The gNMI Target definition can be given in the YAML test file or via command arguments. The command line option override the values in the test file.

    All initial configurations are applied sequentially, before any test runs. Reference there the JSON files prepared before.

    The tests are executed sequentially, sending gNMI Set and Get messages for xpaths with JSON payloads.

    The tests can validate the conformance of the data to OpenConfig models, as well as compare the values.
    The selected `class_name` defines the test behavior, and the `args` values needed. See [here](https://github.com/google/gnxi/blob/master/oc_config_validate/docs/testclasses.md) for the list of available tests.

    The are some example test files [here](https://github.com/google/gnxi/tree/master/oc_config_validate/tests).

### Run

 1. Validate your configurations against a gNMI target using:

    ```
        python3 -m oc_config_validate -tests TESTS_FILE -results RESULTS_FILE \
           [--target HOSTNAME:PORT] \
           [--username USERNAME] [--password PASSWORD] \
           [--tls_host_override TLS_HOSTNAME | --no_tls] \
           [--target_cert_as_root_ca | --root_ca_cert FILE ] \
           [--cert_chain FILE --private_key FILE]
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

### Options

#### gNMI TLS connection options

By default, *oc_config_validate* will use TLS for the gNMI connection, will
validate the hostname presented by the Target in its certificate.

Optionally, self-signed certificates presented by the Target can be trusted.

 *  Use the `tls_host_override` option to provide the hostname value present
    in the Target's certificate, if different from the Target hostname.
 *  Use the `target_cert_as_root_ca` option to fetch the TLS certificate from
    the Target and use it as the client Root CA. This effectively trusts
    self-signed certificates from the Target.
 *  Use the `root_ca_cert` option to provide the Root CA certificate file to use.
 *  Use the `private_key` and `cert_chain` options to provide the TLS key and certificate
    files that the client will present to the Target.

In case of errors, use the `--debug` flag to help understand the underlying TLS issue, if any.

 > With care, use the `no_tls` option not to use TLS for the gNMI connection.
 >
 > Warning: All communication will be in plain text.

#### OpenConfig Models

`oc_config_validate` includes the Python bindings of most OpenConfig models, in the package `oc_config_validate.models`.

Run `python3 -m oc_config_validate -models` to get a list of the versions (revisions) of the OC models used.

#### Timing in gNMI Sets and Gets

The Target can take some time to process configurations done via gNMI Set messages. By default, `oc_config_validate` waits 10 seconds after a successful gNMI Set message.
This time is customized with the `gnmi_set_cooldown_secs` option in the targe configuration or in a command line argument.

Similarly, the Target can take some time to show the expected values in a gNMI Get message.
Some testcases have a retry mechanism, that will repeat the gNMI Get message and comparisons (against an OC model, against an expected JSON text, or both) until it succeeds or times out.

See [here](https://github.com/google/gnxi/blob/master/oc_config_validate/docs/testclasses.md) for more details.

## Copyright

Copyright 2022

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
