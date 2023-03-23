"""Test cases based on gNMI Set/Get requests."""

import json

from oc_config_validate import schema, testbase


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

    def setUp(self):
        self.assertArgs(["xpath", "model", "json_value"])
        xpath = self.xpath.rstrip("/") + "/config"
        self.assertXpath(xpath)
        self.assertModelXpath(self.model, xpath)
        self.assertIsInstance(self.json_value, dict,
                              "The value is not a valid JSON object")
        model = schema.ocContainerFromPath(self.model, xpath)
        self.assertJsonModel(json.dumps(self.json_value), model,
                             "JSON value to Set does not match the model")

    @testbase.retryAssertionError
    def _get_check_retry(self):
        # gNMI Get on /config
        xpath = self.xpath.rstrip("/") + "/config"
        resp_val = self.gNMIGetAssertJson(xpath)
        model = schema.ocContainerFromPath(self.model, xpath)
        self.assertJsonModel(resp_val, model,
                             "Get /config does not match the model")
        got = json.loads(resp_val)
        cmp, diff = schema.intersectCmp(self.json_value, got)
        self.assertTrue(cmp, diff)

        # gNMI Get on /state
        xpath = self.xpath.rstrip("/") + "/state"
        resp = self.gNMIGet(xpath)
        self.assertIsNotNone(resp, "No gNMI GET response")
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        model = schema.ocContainerFromPath(self.model, xpath)
        self.assertJsonModel(resp_val, model,
                             "Get /state does not match the model")
        got = json.loads(resp_val)
        cmp, diff = schema.intersectCmp(self.json_value, got)
        self.assertTrue(cmp, diff)

    def testSetConfigCheckState(self):
        # gNMI Set on /config
        xpath = self.xpath.rstrip("/") + "/config"
        self.assertTrue(
            self.gNMISetUpdate(xpath, self.json_value),
            "gNMI Set did not succeed.")
        self._get_check_retry()


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

    def setUp(self):
        self.assertArgs(["xpath"])
        self.assertXpath(self.xpath)

    @testbase.retryAssertionError
    def _get_check_retry(self):
        # gNMI Get on /config to verify it is deleted
        self.assertFalse(
            self.gNMIGet(self.xpath.rstrip("/") + "/config"),
            "There is still a /config container for the xpath.")

        # gNMI Get on /state to verify it is deleted
        self.assertFalse(
            self.gNMIGet(self.xpath.rstrip("/") + "/state"),
            "There is still a /state container for the xpath.")

    def testDeleteConfigCheckState(self):

        # gNMI Get on /config to verify it is configured
        self.assertTrue(
            self.gNMIGet(self.xpath.rstrip("/") + "/config"),
            "There is no configured /config container for the xpath.")

        # gNMI Delete on the xpath
        self.assertTrue(
            self.gNMISetDelete(self.xpath.rstrip("/")),
            "gNMI Delete did not succeed.")

        # gNMI Gets to confirm deletion
        self._get_check_retry()
