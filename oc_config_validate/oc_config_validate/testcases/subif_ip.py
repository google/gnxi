"""Test cases for subinterfaces with IPv4 addresses."""

import json

from pyangbind.lib import pybindJSON
from retry import retry

from oc_config_validate import schema, testbase
from oc_config_validate.models.interfaces import \
    openconfig_interfaces as oc_interfaces


class SetSubifDhcp(testbase.TestCase):
    """Tests configuring DHCP on a subinterface.

    1. A gNMI Set message is sent to configure the subinterface.
    1. A gNMI Get message on the /config container validates it.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        dhcp: True to enable DHCP, defaults to False.
    """
    interface = ""
    index = 0
    dhcp = False

    def test0100(self):
        self.assertArgs(["interface"])

        iface = oc_interfaces().interfaces.interface.add(self.interface)
        subif = iface.subinterfaces.subinterface.add(self.index)
        if self.index:
            subif.vlan.config.vlan_id = self.index
        subif.ipv4.config.dhcp_client = self.dhcp
        xpath = "/interfaces/interface[name=%s]" % self.interface
        _json_value = json.loads(pybindJSON.dumps(iface, mode='ietf'))
        schema.fixSubifIndex(_json_value)
        self.assertTrue(self.gNMISetUpdate(xpath, _json_value),
                        "gNMI Set did not succeed.")

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0200(self):
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/config/dhcp-client") % (
            self.interface, self.index)
        resp_val = self.gNMIGetAssertBool(xpath)
        self.assertEqual(resp_val, self.dhcp,
                         "Dhcp client config = %s, wanted %s" % (
                             resp_val, self.dhcp))

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0300(self):
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/state/dhcp-client") % (
            self.interface, self.index)
        resp_val = self.gNMIGetAssertBool(xpath)
        self.assertEqual(resp_val, self.dhcp,
                         "Dhcp client state = %s, wanted %s" % (
                             resp_val, self.dhcp))


class CheckSubifDhcpState(testbase.TestCase):
    """Checks the DHCP state on a subinterface.

    1. A gNMI Get message on the /state container.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        dhcp: True to enable DHCP, defaults to False.
    """
    interface = ""
    index = 0
    dhcp = False

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0100(self):
        """"""
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/state/dhcp-client") % (
            self.interface, self.index)
        resp_val = self.gNMIGetAssertBool(xpath)
        self.assertEqual(resp_val, self.dhcp,
                         "Dhcp client state = %s, wanted %s" % (
                             resp_val, self.dhcp))


class CheckSubifDhcpConfig(testbase.TestCase):

    """Checks the DHCP config on a subinterface.

    1. A gNMI Get message on the /config container.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        dhcp: True to enable DHCP, defaults to False.
    """
    interface = ""
    index = 0
    dhcp = False

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0100(self):
        """"""
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/config/dhcp-client") % (
            self.interface, self.index)
        resp_val = self.gNMIGetAssertBool(xpath)
        self.assertEqual(resp_val, self.dhcp,
                         "Dhcp client config = %s, wanted %s" % (
                             resp_val, self.dhcp))


class AddSubifIp(testbase.TestCase):
    """Tests configuring an IP on a subinterface.

    1. A gNMI Set message is sent to configure the subinterface.
    1. A gNMI Get message on the /config container validates it.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        address: IPv4 address to add.
        prefix_length: Prefix lenght of the IPv4 address to add.
    """
    interface = ""
    index = 0
    address = ""
    prefix_length = 32
    iface = None

    def setUp(self):
        self.iface = oc_interfaces().interfaces.interface.add(self.interface)
        subif = self.iface.subinterfaces.subinterface.add(self.index)
        if self.index:
            subif.vlan.config.vlan_id = self.index
        self.addr = subif.ipv4.addresses.address.add(self.address)
        self.addr.config.ip = self.address
        self.addr.config.prefix_length = self.prefix_length

    def test0100(self):
        self.assertArgs(["interface", "address", "prefix_length"])
        xpath = "/interfaces/interface[name=%s]" % self.interface
        _json_value = json.loads(pybindJSON.dumps(self.iface, mode='ietf'))
        schema.fixSubifIndex(_json_value)
        self.assertTrue(self.gNMISetUpdate(xpath, _json_value),
                        "gNMI Set did not succeed.")

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0200(self):
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/addresses"
                 "/address[ip=%s]/config") % (
            self.interface, self.index, self.address)
        resp_val = self.gNMIGetAssertJson(xpath)
        self.assertJsonModel(
            resp_val, self.addr.config,
            'gNMI Get on the /config container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(self.addr.config, mode='ietf')))
        self.assertJsonCmp(resp_val, want)

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0300(self):
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/addresses"
                 "/address[ip=%s]/state") % (
            self.interface, self.index, self.address)
        resp_val = self.gNMIGetAssertJson(xpath)

        self.addr.state._set_ip(self.address)
        self.addr.state._set_prefix_length(self.prefix_length)
        self.assertJsonModel(
            resp_val, self.addr.state,
            'gNMI Get on the /state container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(self.addr.state, mode='ietf')))
        self.assertJsonCmp(resp_val, want)


class RemoveSubifIp(testbase.TestCase):
    """Tests removing an IP on a subinterface.

    1. gNMI Get message on the /config container, to check it is configured.
    1. gNMI Set message to delete the ip.
    1. gNMI Get message on the /config container to check it is not there.
    1. gNMI Get message on the /state container to check it is not there.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        address: IPv4 address to remove.
        prefix_length: Prefix lenght of the IPv4 address to remove.
    """
    interface = ""
    index = 0
    address = ""
    prefix_length = 32
    xpath = ""

    def test0100(self):
        self.assertArgs(["interface", "address", "prefix_length"])
        self.xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                      "subinterface[index=%d]/ipv4/addresses"
                      "/address[ip=%s]") % (
            self.interface, self.index, self.address)
        if not self.gNMIGet(self.xpath + "/config"):
            self.log("IP %s not configured at %s.%s",
                     self.address, self.interface, self.index)
            return
        self.assertTrue(self.gNMISetDelete(self.xpath),
                        "gNMI Delete did not succeed.")

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0200(self):
        resp = self.gNMIGet(self.xpath + "/config")
        self.assertIsNone(
            resp,
            "IP %s still configured at %s.%s" % (self.address,
                                                 self.interface, self.index))

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0300(self):
        resp = self.gNMIGet(self.xpath + "/state")
        self.assertIsNone(
            resp,
            "IP %s still at %s.%s" % (self.address,
                                      self.interface, self.index))


class CheckSubifIpState(testbase.TestCase):
    """Checks the state on an ip address configured on a subinterface.

    1. A gNMI Get message on the /state container.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        address: IPv4 address.
        prefix_length: Prefix lenght of the IPv4 address.
    """
    interface = ""
    index = 0
    address = ""
    prefix_length = 32

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0100(self):
        """"""
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/addresses/"
                 "address[ip=%s]/state") % (
            self.interface, self.index, self.address)
        resp_val = self.gNMIGetAssertJson(xpath)
        iface = oc_interfaces().interfaces.interface.add(self.interface)
        subif = iface.subinterfaces.subinterface.add(self.index)
        if self.index:
            subif.vlan.config.vlan_id = self.index
        addr = subif.ipv4.addresses.address.add(self.address)
        addr.state._set_ip(self.address)
        addr.state._set_prefix_length(self.prefix_length)
        self.assertJsonModel(
            resp_val, addr.state,
            'gNMI Get on the /state container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(addr.state, mode='ietf')))
        self.assertJsonCmp(resp_val, want)


class CheckSubifIpConfig(testbase.TestCase):
    """Checks the config on an ip address on a subinterface.

    1. A gNMI Get message on the /config container.

    All arguments are read from the Test YAML description.

    Args:
        interface: Name of the physical interface.
        index: Index of the subinterface, defaults to 0.
        address: IPv4 address.
        prefix_length: Prefix lenght of the IPv4 address.
    """
    interface = ""
    index = 0
    address = ""
    prefix_length = 32

    @retry(exceptions=AssertionError, tries=5, delay=15)
    def test0100(self):
        """"""
        xpath = ("/interfaces/interface[name=%s]/subinterfaces/"
                 "subinterface[index=%d]/ipv4/addresses/"
                 "address[ip=%s]/config") % (
            self.interface, self.index, self.address)
        resp_val = self.gNMIGetAssertJson(xpath)
        iface = oc_interfaces().interfaces.interface.add(self.interface)
        subif = iface.subinterfaces.subinterface.add(self.index)
        if self.index:
            subif.vlan.config.vlan_id = self.index
        addr = subif.ipv4.addresses.address.add(self.address)
        addr.config.ip = self.address
        addr.config.prefix_length = self.prefix_length
        self.assertJsonModel(
            resp_val, addr.config,
            'gNMI Get on the /config container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(addr.config, mode='ietf')))
        self.assertJsonCmp(resp_val, want)
