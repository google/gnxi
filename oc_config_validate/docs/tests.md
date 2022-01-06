## How to create oc_config_validate tests

`oc_config_validate` executes the test listed in a YAML file with the following structure:

```
!TestContext

# Informative text about this test run.
description: [:string]

# Optional list of text labels, passed to the Results. Can be used for tagging and filtering.
labels:
- [:string]
- [:string]

# Target to connect to. Overridden by command arguments.
target:
  !Target
  target: [:string]  # As "hostname:port".
  username: [:string]
  password: [:string]  # Attention, this is clear text.
  no_tls: [:bool]
  private_key: [:string]  # Path to file.
  cert_chain: [:string]  # Path to file.
  root_ca_cert: [:string]  # Path to file.
  tls_host_override: [:string]
  gnmi_set_cooldown_secs: [:int]  # Time to wait after a successful gNMI Set.

# Optional list of initial configurations and xpaths to apply before the tests.
init_configs:
- !InitConfig
  # Path to the initial configuration JSON file.
  filename: [:string]
  # Xpath where to send a gNMI Set Update messages with the configuration.
  xpath: [:string]

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

> The Tags like `!TestContext` and `!TestCase` are needed to easily parse the YAML into objects.

A test run is made of a sequence of tests. Each test is made of a Class, in
the `oc_config_validate.testcases` module, that tests something (it can be as simple as testing a gNMI Get is successful).

Each Class accepts or needs some arguments, which are passed in the `args:` section. See the class documentation to provide its arguments.

> Some `oc_config_validate` tests files for common network devices are already curated in the [tests/](../tests) folder.

> The available test classes are listed in [testclasses](testclasses.md).

### JSON Comparison

The way to compare the JSON values given in the tests with the responses obtained form the gNMI Target is:

 1.  Only the keys from the JSON in the tests are compared. The gNMI response can have more members in the JSON, those are not compared.
 1.  For key in the JSON in the test, the values are compared for equality in the gNMI response.
 1.  If there are nested JSON objects, the same comparison is applied to them.
 1.  If a key contains a list of JSON objects, all objects in the list in the JSON text in the tests must be present in the gNMI response.
     The gNMI response can have more elements in the list, those are not compared.
 1.  The same comparison mechanism applies to JSON objects contained in a list, in the JSON text in the tests.

This means that a test can get a whole model container and compare only a subset of the values, like:

```
xpath : interfaces/interface[name=mgmt]/state
want_json: {
    "enabled": true,
    "oper-status": "UP"
  }
```

or

```
xpath : /local-routes/static-routes/static[prefix=10.10.20.0/24]/next-hops/
want_json:{
    "next-hop": [
      {
        "index": "1",
        "next-hop": "192.168.1.125"
      }
    ]
  }
```
