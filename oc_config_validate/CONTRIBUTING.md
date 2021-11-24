# Contributing

There are 2 main ways to contribute and enhance this tool:

## Write testcases

With a little of Python, you can extend the testcases that oc_config_validate runs. Compare this to writing [unittest.TestCase](https://docs.python.org/3/library/unittest.html#basic-example) classes (see [docs/testcases.md](docs/testcases.md) for details).

If you have a particular behavior of your device under test, or you would like to test something more than gNMI SET/GET operations, you can create a new TestCase class and use it in your YAML file (see [docs/tests.md](docs/tests.md) for details).

### Commit

If you would like to commit new testcases, follow the recommendations in [#commit-code](#commit-code).

## Write tests

Without any coding, you can extensively test your OpenConfig configuration by writing Tests to runs (see [docs/tests.md](docs/tests.md) for details).

You list all related OpenConfig configuration validations, as a sequence of basic (or more advanced, if available) tests, in a YAML-formated file. 
The sequence and individual tests can reflect very closely your real configuration sequence, and the OpenConfig models you are using.

For instance:
 1. Get the hostname.
 1. Get the current device time.
 1. Get an interface named "mgmt" and check its model is as expected.
 1. Get the default route and check it points to the correct gateway.
 1. Set a static route and check it was configured
 1. Set a frequency channel to a radio and check the radio is using it.

 > Remember that *oc_config_validate* applies an initial JSON-formated configuration to the device, before running any test.

The capabilities of what you can test depend on the available testcases. At minimum, testcases to do gNMI SET/GET operations are already provided.

### Commit

If you would like to commit new tests, make sure you test your YAML and JSON files on a real device and there are no parsing errors.

Have your YAML and JSON files analyzed and validated with a parser. Simply the run `lint_project.sh` script.

## Commit code

Before committing Python code, make sure:

 * Format your code using `autopep8`
 * Lint your code using `pylama`
 * Analyze your code using `pytype`
 
To make things a bit easier, there is the `lint_project.sh` script that helps you with this.
