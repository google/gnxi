"""Test cases for static IPv4 routes."""

import json

from pyangbind.lib import pybindJSON
from retry import retry

from oc_config_validate import schema, testbase
from oc_config_validate.models.local_routing import openconfig_local_routing


class AddStaticRoute(testbase.TestCase):
    """Tests configuring a static route.

    1. A gNMI Set message is sent to configure the route.
    1. A gNMI Get message on the /config and /state containers validates it.

    All arguments are read from the Test YAML description.

    Args:
        description: Optional text description of the route.
        prefix: Destination prefix of the static route.
        next_hop: IP of the next hop of the route.
        index: Index of the next hop for the prefix. Defaults to 0.
        metric: Optional numeric metric of the next hop for the prefix.
    """
    description = None
    prefix = ""
    next_hop = ""
    index = 0
    metric = None
    route = None

    def setUp(self):
        self.route = openconfig_local_routing(
        ).local_routes.static_routes.static.add(self.prefix)
        self.route.config.prefix = self.prefix
        if self.description is not None:
            self.route.config.description = self.description
        self.nh = self.route.next_hops.next_hop.add(self.index)
        self.nh.config.index = self.index
        self.nh.config.next_hop = self.next_hop
        if self.metric is not None:
            self.nh.config.metric = self.metric

    def test0001(self):
        self.assertArgs(["prefix", "next_hop"])

    def test0100(self):
        xpath = "/local-routes/static-routes/static[prefix=%s]" % self.prefix
        _json_value = json.loads(pybindJSON.dumps(self.route, mode='ietf'))
        self.assertTrue(self.gNMISetUpdate(xpath, _json_value),
                        "gNMI Set did not succeed.")

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0200(self):
        xpath = ("/local-routes/static-routes/static[prefix=%s]/"
                 "next-hops/next-hop[index=%d]/config") % (
            self.prefix, self.index)
        resp_val = self.gNMIGetAssertJson(xpath)
        self.assertJsonModel(
            resp_val, self.nh.config,
            'gNMI Get on the /config container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(self.nh.config, mode='ietf')))
        self.assertJsonCmp(resp_val, want)

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0300(self):
        xpath = ("/local-routes/static-routes/static[prefix=%s]/"
                 "next-hops/next-hop[index=%d]/state") % (
            self.prefix, self.index)
        resp_val = self.gNMIGetAssertJson(xpath)
        self.nh.state._set_index(self.index)
        self.nh.state._set_next_hop(self.next_hop)
        if self.metric is not None:
            self.nh.state._set_metric(self.metric)
        self.assertJsonModel(
            resp_val, self.nh.state,
            'gNMI Get on the /state container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(self.nh.state, mode='ietf')))
        self.assertJsonCmp(resp_val, want)


class RemoveStaticRoute(testbase.TestCase):
    """Tests removing a static route.

    1. gNMI Get message on the /config container, to check it is configured.
    1. gNMI Set message to delete the route.
    1. gNMI Get message on the /config container to check it is not there.

    All arguments are read from the Test YAML description.

    Args:
        prefix: Destination prefix of the static route.
        index: Index of the next hop for the prefix. Defaults to 0.
    """
    prefix = ""
    index = 0

    def test0000(self):
        self.assertArgs(["prefix"])

    def test0100(self):
        xpath = ("/local-routes/static-routes/static[prefix=%s]"
                 "/next-hops/next-hop[index=%d]") % (
            self.prefix, self.index)
        if not self.gNMIGet(xpath + "/config"):
            self.log("Route to %s via next-hop %d not configured",
                     self.prefix, self.index)
            return
        self.assertTrue(self.gNMISetDelete(xpath),
                        "gNMI Delete did not succeed.")

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0200(self):
        xpath = ("/local-routes/static-routes/static[prefix=%s]"
                 "/next-hops/next-hop[index=%d]/config") % (
            self.prefix, self.index)
        resp = self.gNMIGet(xpath)
        self.assertIsNone(
            resp,
            "Route to %s via next-hop %d still configured" % (self.prefix,
                                                              self.index))

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0300(self):
        xpath = ("/local-routes/static-routes/static[prefix=%s]"
                 "/next-hops/next-hop[index=%d]/state") % (
            self.prefix, self.index)
        resp = self.gNMIGet(xpath)
        self.assertIsNone(
            resp,
            "Route to %s via next-hop %d still present" % (self.prefix,
                                                           self.index))


class CheckRouteState(testbase.TestCase):
    """Checks the state on a static route.

    1. A gNMI Get message on the /state container.

    All arguments are read from the Test YAML description.

    Args:
        description: Optional text description of the route.
        prefix: Destination prefix of the static route.
        next_hop: IP of the next hop of the route.
        index: Index of the next hop for the prefix. Defaults to 0.
        metric: Optional numeric metric of the next hop for the prefix.
    """
    description = None
    prefix = ""
    next_hop = ""
    index = 0
    metric = None

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0100(self):
        """"""
        xpath = ("/local-routes/static-routes/static[prefix=%s]"
                 "/next-hops/next-hop[index=%d]/state") % (
            self.prefix, self.index)

        ste = openconfig_local_routing().local_routes.static_routes.static.\
            add(prefix=self.prefix).next_hops.next_hop.add(index=self.index).\
            state
        ste._set_index(self.index)
        ste._set_next_hop(self.next_hop)
        if self.metric is not None:
            ste._set_metric(self.metric)
        resp_val = self.gNMIGetAssertJson(xpath)
        self.assertJsonModel(
            resp_val, ste,
            'gNMI Get on the /state container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(ste, mode='ietf')))
        self.assertJsonCmp(resp_val, want)


class CheckRouteConfig(testbase.TestCase):
    """Checks the configuration on a static route.

    1. A gNMI Get message on the /config container.

    All arguments are read from the Test YAML description.

    Args:
        description: Optional text description of the route.
        prefix: Destination prefix of the static route.
        next_hop: IP of the next hop of the route.
        index: Index of the next hop for the prefix. Defaults to 0.
        metric: Optional numeric metric of the next hop for the prefix.
    """
    description = None
    prefix = ""
    next_hop = ""
    index = 0
    metric = None

    @retry(exceptions=AssertionError, tries=3, delay=10)
    def test0100(self):
        """"""
        xpath = ("/local-routes/static-routes/static[prefix=%s]"
                 "/next-hops/next-hop[index=%d]/config") % (
            self.prefix, self.index)

        cfg = openconfig_local_routing().local_routes.static_routes.static.\
            add(prefix=self.prefix).next_hops.next_hop.add(index=self.index).\
            config.config()
        cfg._set_index(self.index)
        cfg._set_next_hop(self.next_hop)
        if self.metric is not None:
            cfg._set_metric(self.metric)
        resp_val = self.gNMIGetAssertJson(xpath)
        self.assertJsonModel(
            resp_val, cfg,
            'gNMI Get on the /config container does not match model')
        want = json.loads(
            schema.removeOpenConfigPrefix(
                pybindJSON.dumps(cfg, mode='ietf')))
        self.assertJsonCmp(resp_val, want)
