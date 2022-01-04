"""Test cases based on gNMI Set/Get requests."""

import json

from oc_config_validate import schema, target, testbase


class SetConfigCheckState(testbase.TestCase):
    """Configures on the /config container and checks the /state container.

    1. The intended JSON-IETF configuration is checked for schema validity.
    1. It is sent in a gNMI Set request,  to the /config container.
    1. The same container is fetched in a gNMI Get request and checked for
         schema validity. It is compared with the sent configuration.
    1. The /state container is fetched in a gNMI Get request and checked for
         schema validity. It is compared with the sent configuration.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to write and read, without ending /config or /state.
        json_value: JSON-IETF value to check set, get and compare.
        model: Python binding class to check the JSON reply against.
    """
    xpath = ""
    json_value = None
    model = ""

    def test0100(self):
        """gNMI Set on /config"""
        self.assertArgs(["xpath", "model", "json_value"])
        self.assertTrue(target.isPathOK(self.xpath),
                        "Xpath '%s' not a gNMI path" % self.xpath)
        self.assertIsNotNone(schema.containerFromName(self.model +
                                                      ".config.config"),
                             "Unable to find model '%s' binding" % self.model)
        self.assertIsInstance(self.json_value, dict,
                              "The value is not a valid JSON object")
        self.assertJsonModel(json.dumps(self.json_value),
                             self.model + ".config.config",
                             "JSON value to Set does not match the model")
        self.assertTrue(
            self.gNMISetUpdate(
                self.xpath.rstrip("/") + "/config", self.json_value),
            "gNMI Set did not succeed.")

    @testbase.retryAssertionError
    def test0200(self):
        """gNMI Get on /config"""
        self.assertArgs(["xpath", "model", "json_value"])
        resp_val = self.gNMIGetJson(self.xpath.rstrip("/") + "/config")
        self.assertJsonModel(resp_val,
                             self.model + ".config.config",
                             "Get /config does not match the model")
        got = json.loads(resp_val)
        cmp, diff = target.intersectCmp(self.json_value, got)
        self.assertTrue(cmp, diff)

    @testbase.retryAssertionError
    def test0300(self):
        """gNMI Get on /state"""
        self.assertArgs(["xpath", "model", "json_value"])
        resp = self.gNMIGet(self.xpath.rstrip("/") + "/state")
        self.assertIsNotNone(resp, "No gNMI GET response")
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        self.assertJsonModel(resp_val,
                             self.model + ".state.state",
                             "Get /state does not match the model")
        got = json.loads(resp_val)
        cmp, diff = target.intersectCmp(self.json_value, got)
        self.assertTrue(cmp, diff)


class DeleteConfigCheckState(testbase.TestCase):
    """Deletes the xpath and checks the /config and /state container after.

    1. gNMI Get request validates that the /config container exists.
    1. gNMI Delete request removes the xpath.
    1. gNMI Get request validates that the /config container no longer exists.
    1. gNMI Get request validates that the /state container no longer exists.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to delete, without ending /config or /state.
    """
    xpath = ""

    def test0100(self):
        """gNMI Get on /config to verify it is configured"""
        self.assertArgs(["xpath"])
        self.assertTrue(target.isPathOK(self.xpath),
                        "Xpath '%s' not a gNMI path" % self.xpath)
        self.assertTrue(
            self.gNMIGet(self.xpath.rstrip("/") + "/config"),
            "There is no configured /config container for the xpath.")

    def test0200(self):
        """gNMI Delete on the xpath"""
        self.assertTrue(
            self.gNMISetDelete(self.xpath.rstrip("/")),
            "gNMI Delete did not succeed.")

    @testbase.retryAssertionError
    def test0300(self):
        """gNMI Get on /config to verify it is deleted"""
        self.assertFalse(
            self.gNMIGet(self.xpath.rstrip("/") + "/config"),
            "There is still a /config container for the xpath.")

    @testbase.retryAssertionError
    def test0400(self):
        """gNMI Get on /state to verify it is deleted"""
        self.assertFalse(
            self.gNMIGet(self.xpath.rstrip("/") + "/state"),
            "There is still a /state container for the xpath.")
