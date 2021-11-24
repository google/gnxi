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
# List of Tests to run
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

 1.  Only the common JSON Keys are compared. If there are no common keys, it is considered as unequal.
 1.  For each common key, the values are compared for equality.
 1.  If there are nested JSON objects, the same comparison is applied to them.

This means that a test can get a whole model container (like `interfaces/interface[name=mgmt]/state`) and compare only a subset of the values, like:

```
{
  "enabled": true,
  "oper-status": "UP"
}
```