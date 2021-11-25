"""Test cases based on gNMI Set requests."""

import json

from oc_config_validate import schema, testbase


class SetUpdate(testbase.TestCase):
    """Sends gNMI Set of an xpath with a value.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to write.
        value: Value to set; can be numeric, string or JSON-IETF.
    """
    xpath = ""
    value = None

    def test0200(self):
        """"""
        self.assertArgs(["xpath", "value"])
        self.assertXpath(self.xpath)
        self.assertTrue(self.gNMISetUpdate(self.xpath, self.value),
                        "gNMI Set did not succeed.")


class SetDelete(testbase.TestCase):
    """Sends gNMI Set of an xpath to delete.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to delete.
    """
    xpath = ""

    def test0200(self):
        """"""
        self.assertArgs(["xpath"])
        self.assertXpath(self.xpath)
        self.assertTrue(self.gNMISetDelete(self.xpath),
                        "gNMI Set did not succeed.")


class JsonCheckSetUpdate(testbase.TestCase):
    """Sends gNMI Set with a schema-checked JSON-IETF value.

    The intended JSON-IETF configuration is first checked for schema validity.

    All arguments are read from the Test YAML description.

    Args:
        xpath: gNMI path to write.
        json_value: JSON-IETF value to check and set.
        model: Python binding class to check the JSON reply against.
    """
    xpath = ""
    json_value = None
    model = ""

    def test0100(self):
        """"""
        self.assertArgs(["xpath", "json_value", "model", ])
        self.assertXpath(self.xpath)
        self.assertIsNotNone(schema.containerFromName(self.model),
                             "Unable to find model '%s' binding" % self.model)
        self.assertIsInstance(self.json_value, dict,
                              "The value is not a valid JSON object")
        self.assertJsonModel(json.dumps(self.json_value), self.model,
                             "JSON to Set does not match the model")
        self.assertTrue(self.gNMISetUpdate(self.xpath, self.json_value),
                        "gNMI Set did not succeed.")
