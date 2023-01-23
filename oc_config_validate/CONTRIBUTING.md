# Contributing

## Install and run from source

1. Install Python3.9.

1. Clone this repo and install the needed dependencies. Preferably, use a virtual environment:

```
git clone https://github.com/google/gnxi.git
cd gnxi/oc_config_validate
virtualenv --clear venv
source venv/bin/activate
pip3 install -r requirements.txt
```

1. Check it is all working as expected:

```
python3 -m oc_config_validate --version
python3 -m oc_config_validate -models
```

### Update OC models

To get the latest OC Models from GitHub, running `oc_config_validate/update_models.sh`.

This will take some time, be patient.

#### Use specific OpenConfig models

If specific OpenConfig models need to be used, place them in a directory (here `$MODELS_DIRECTORY`) and have **pyang** create the bindings and store them in `oc_config_validate/models/` (also write their revisions in `oc_config_validate/models/versions`):

See [docs/oc_models.md](docs/oc_models.md) for details.

## Augmenting `oc_config_validate`

There are 2 main ways to contribute to this tool:

### Write testcases

With a little of Python, you can extend the testcases that oc_config_validate runs. Compare this to writing [unittest.TestCase](https://docs.python.org/3/library/unittest.html#basic-example) classes (see [docs/testcases.md](docs/testcases.md) for details).

If you have a particular behavior of your device under test, or you would like to test something more than gNMI SET/GET operations, you can create a new TestCase class and use it in your YAML file (see [docs/tests.md](docs/tests.md) for details).

#### Commit

If you would like to commit new testcases, follow the recommendations in [#commit-code](#commit-code).

### Write tests

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

#### Commit

If you would like to commit new tests, make sure you test your YAML and JSON files on a real device and there are no parsing errors.

Have your YAML and JSON files analyzed and validated with a parser. Simply the run `lint.sh` script.

## Committing code

Before committing JSON or YAML files, make sure:

* Format your JSON code using `json.tool`.
* Format your YAML code using `yamllint`.

Before committing Python code, make sure:

 * Format your code using `autopep8`
 * Lint your code using `pylama`
 * Analyze your code using `pytype`

> To make things a bit easier, there is the `lint.sh` script that helps you with this.

### Unit Tests

For your Python code, write relevant Unit tests, using the `unittest` library.

Place your testing code in `py_tests/test_*.py` and run the tests with `python3 -m unittest discover py_tests`.

## Updating the `oc_config_validate` PyPi package

1. Update the OC models, by running `./update_models.sh`
1. Update the version of oc_config_validate, by editing `__init__.py`
1. Run the demo to verify it all works, by running `demo/run_demo.sh`
1. Run `oc_config_validate` with a real gNMI target to verify it all works.
1. Create a cleared virtual environment and call `python3 -m build` to build the packages
1. Upload the latest packages to TestPyPi

```
python3 -m twine upload --verbose --repository testpypi --skip-existing dist/*
```
