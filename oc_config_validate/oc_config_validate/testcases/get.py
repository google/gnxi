"""Test cases based on gNMI Get requests."""

import json

from oc_config_validate import schema, target, testbase


class Get(testbase.TestCase):
    """Tests that gNMI Get of an xpath returns no error.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to read.
    """
    xpath = ""

    def test0200(self):
        """"""
        self.assertArgs(["xpath"])
        self.assertXpath(self.xpath)
        self.gNMIGet(self.xpath)


class GetCompare(testbase.TestCase):
    """Compares that gNMI Get of an xpath returns the expected value.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to read.
        want: Expected value; can be numeric, string or JSON-IETF.
    """

    xpath = ""
    want = ""

    @testbase.retryAssertionError
    def test0200(self):
        """"""
        self.assertArgs(["xpath", "want"])
        self.assertXpath(self.xpath)
        resp_val = self.gNMIGet(self.xpath)
        self.assertIsNotNone(resp_val, "No gNMI GET response")
        got = target.typedValueToPython(resp_val, type(self.want))
        self.assertEqual(type(got), type(self.want),
                         "Values of different types")
        if isinstance(self.want, dict):
            cmp, diff = target.intersectCmp(self.want, got)
            self.assertTrue(cmp,  diff)
        else:
            self.assertEqual(self.want, got)


class GetJsonCheck(testbase.TestCase):
    """Tests that gNMI Get of an xpath returns a schema-valid JSON response.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to read.
        model: Python binding class to check the JSON reply against.
    """

    xpath = ""
    model = ""

    @testbase.retryAssertionError
    def test0200(self):
        """"""
        self.assertArgs(["xpath", "model"])
        self.assertXpath(self.xpath)
        self.assertIsNotNone(schema.containerFromName(self.model),
                             "Unable to find model '%s' binding" % self.model)
        resp = self.gNMIGet(self.xpath)
        self.assertIsNotNone(resp, "No gNMI GET response")
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        self.assertJsonModel(resp_val, self.model,
                             "Get response JSON does not match model")


class GetJsonCheckCompare(testbase.TestCase):
    """Checks for schema validity and compares a gNMI Get response.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to read.
        model: Python binding class to check the JSON reply against.
        want_json: Expected JSON-IETF value.
    """

    xpath = ""
    model = ""
    want_json = None

    @testbase.retryAssertionError
    def test0200(self):
        """"""
        self.assertArgs(["xpath", "want_json", "model"])
        self.assertXpath(self.xpath)
        self.assertIsNotNone(schema.containerFromName(self.model),
                             "Unable to find model '%s' binding" % self.model)
        self.assertIsInstance(self.want_json, dict,
                              "'want_json' is not a valid JSON object")
        resp = self.gNMIGet(self.xpath)
        self.assertIsNotNone(resp, "No gNMI GET response")
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        self.assertJsonModel(resp_val, self.model,
                             "Get response does not match the model")
        got = json.loads(resp_val)
        cmp, diff = target.intersectCmp(self.want_json, got)
        self.assertTrue(cmp, diff)
