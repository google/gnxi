"""Test cases based on gNMI Subscribe ONCE requests."""

import time

from oc_config_validate import schema, testbase


class SubsOnceTestCase(testbase.TestCase):
    """Subscribes ONCE and checks the response elements.

    The arguments are optional, only checked when defined.

    Args:
        xpaths: List of gNMI paths to subscribe to.
        notifications_count: Number of expected Notifications.
        max_delay_secs: Maximum delay of the replies, measured with the
          timestamp value(s).
    """
    notifications_count = None
    max_delay_secs = None
    xpaths = None

    def subscribeOnce(self):
        """"""
        for xpath in self.xpaths:
            self.assertXpath(xpath)
        self.now = int(time.time())
        self.responses = self.gNMISubsOnce(self.xpaths)
        self.assertIsNotNone(self.responses, "No gNMI Subscribe response")

        if self.max_delay_secs:
            for n in self.responses:
                timestamp_diff = (n.timestamp // 1000000000) - self.now
                self.assertTrue(
                    (self.max_delay_secs * -
                     1) < timestamp_diff < self.max_delay_secs,
                    f"Timestamp diff too long: {timestamp_diff} secs")

        if self.notifications_count:
            notifications = len(self.responses)
            self.assertEqual(
                notifications, self.notifications_count,
                f"Expected {self.notifications_count} notifications, "
                f"got {notifications}")


class CountUpdatesCheckType(SubsOnceTestCase):
    """Subscribes ONCE and checks the returned updates and their value types.

    Args:
        xpaths: List of /state gNMI paths to subscribe to.
        updates_count: Number of expected Update messages.
        values_type: Python type of the values of the Updates.

    """
    values_type = None
    updates_count = None

    def test100(self):
        """"""
        self.assertArgs(["xpaths", "values_type", "updates_count"])

        self.subscribeOnce()

        updates = 0
        for n in self.responses:
            updates += len(n.update)
            for u in n.update:
                self.assertTrue(
                    u.val.HasField(self.values_type),
                    f"Value of Update {schema.pathToString(u.path)} "
                    f"is not of type {self.values_type}: {u.val}")

        self.assertEqual(
            updates, self.updates_count,
            f"Expected {self.updates_count} Updates, got: {updates}")


class CheckStateLeafs(SubsOnceTestCase):
    """Subscribes ONCE and checks the updates againts the state OC model.

    All arguments are read from the Test YAML description.

    Args:
        xpaths: List of /state gNMI paths to subscribe to.
                If the paths do not end in ''/state', it will be added.
        model: Python binding class to check the JSON reply against.
        check_missing_model_paths: If True, missing OC Model leaf paths in the
                                   Updates replies are checked.
    """
    model = None
    check_missing_model_paths = False

    def test100(self):
        """"""
        self.assertArgs(["xpaths", "model"])

        self.subscribeOnce()

        want_paths = []
        for xpath in self.xpaths:
            # Ensure the path ends in /state.
            elems = xpath.split("/")
            if elems[-1] != "state":
                elems.append("state")
            xpath = "/".join(elems)
            self.assertModelXpath(self.model, xpath)
            want_paths.extend(schema.ocLeafsPaths(self.model, xpath))

        got_updates = []
        for n in self.responses:
            got_updates.extend(n.update)

        self.assertGreater(
            len(got_updates), 0,
            "There are no Updates as reply to the Subscription")

        got_paths = []
        for u in got_updates:
            got_path = schema.pathToString(u.path)
            got_paths.append(got_path)
            self.assertTrue(
                schema.isPathIn(got_path, want_paths),
                f"Unexpected update path {got_path} for OC model {self.model},"
                f"expected {want_paths}")

        if self.check_missing_model_paths:
            for want_path in want_paths:
                self.assertIn(
                    want_path, got_paths,
                    f"Missing update path {want_path} for OC model"
                    f" {self.model}, got {got_paths}")
