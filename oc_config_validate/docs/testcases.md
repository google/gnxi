## How to create oc_config_validate testcases

`oc_config_validate` executes the Python classes in the `oc_config_validate.testcases` module.
These classes are written similarly to a [unittests.TestCase](https://docs.python.org/3/library/unittest.html#basic-example) class.

The signature of a testcase Class is:

```
from oc_config_validate import test_base


class TestSomething(test_base.TestCase):
    """Tests something.

    All arguments are read from the Test YAML description.

    Args:
        a_path: path to read.
        an_int: number to check.
    """

    def test0100(self):
        self.log("Running TestSomething(a_path=%s)", self.a_path)
        self.log("Using Target %s", self.test_target)
        resp = self.gNMIGet(self.a_path)
        self.assertEqual(resp.int_val, self.an_int )

```

Important to notice:

 * All Class inherits from `test_base.TestCase`. 
 
 * It is recommended to interact with the gNMI Target with methods like `self.gNMIGet()` and `self.gNMISetUpdate()`, but direct access to the Target is possible using `self.test_target`.

 * The Class will get as attributes the arguments passed from the Test YAML description. It would be beneficial to first test that the arguments were passed (check the existence of the attributes) at first.

 * The Class has test methods, that **MUST** start with `test` prefix. The methods are executed in alphabetical order.
 
 * The methods can call `unittests` methods, such as `assertTrue()`, `assertEqual()`, `log()`, `fail()`, etc.

> Some testcases for simple gNMI GET and SET tests are already provided for use in the `oc_config_validate.testcases` module.
