"""Test cases based on gNMI Subscribe ON requests."""

from oc_config_validate import schema, testbase


class SubsOnChangeTestCase(testbase.TestCase):
    """"""
    notifications_before_sync = []
    notifications_after_sync = []
    xpath = ""
    assert_all_paths_updated = False

    def assertNotifications(self):
        """Asserts that the received Updates have different values.

        Args:
          assert_all_paths_updated: If True, it asserts that all paths in
          the initial Update message had a change after.
        """
        self.assertTrue(bool(self.notifications_before_sync),
                        "No gNMI Subscribe response before sync_response")
        initial_updates = {}
        for n in self.notifications_before_sync:
            for u in n.update:
                path = schema.pathToString(u.path)
                self.assertTrue(
                    schema.isPathInRequestedPaths(path, [self.xpath]),
                    f"Unexpected update path {path} for subscription")
                initial_updates[schema.pathToString(
                    u.path)] = schema.typedValueToPython(u.val)

        self.assertTrue(bool(self.notifications_after_sync),
                        "No gNMI Subscribe response after sync_response")
        last_updates = initial_updates.copy()
        changed_paths = set()
        for n in self.notifications_after_sync:
            for u in n.update:
                path = schema.pathToString(u.path)
                value = schema.typedValueToPython(u.val)
                self.assertNotEqual(
                    last_updates[path], value,
                    f"Update for {path} has the same previous value {value}")
                last_updates[path] = value
                changed_paths.add(path)

        if self.assert_all_paths_updated:
            paths_not_updated = set(
                initial_updates.keys()).difference(changed_paths)
            self.assertEqual(
                len(paths_not_updated), 0,
                f"No Updates after sync_response for paths: "
                f"{paths_not_updated}")


class SubscribeAndSet(SubsOnChangeTestCase):
    """Subscribes ON_CHANGE and checks the response elements.

    1. Subscribes ON_CHANGE to the path.
    1. Collects initial Notifications.
    1. Issues a gNMI Set request.
    1. Collects Notifications until timeout.

    Does not support heartbeat.
    Does not support OnlyUpdates, the initial update is needed for comparison.

    Checks that the paths in the inital Updates are within the Request path.
    Checks that the Update values after the change are different from before
      the gNMI Set.

    Args:
        xpath: gNMI path to subscribe to.
        set_xpath: gNMI path to change.
        set_value: Value to change to.
        timeout_secs: Seconds to wait to receive an Update after Set.
        assert_all_paths_updated: If True, it asserts that all paths in the
          initial Update message had a change after. Defaults to False.
    """
    xpath = ""
    set_xpath = ""
    set_value = None
    timeout_secs = 10

    def testSubscribeOnChange(self):
        """"""

        def changeSet():
            self.gNMISetUpdate(self.set_xpath, self.set_value)

        self.assertArgs(["xpath", "set_xpath", "set_value"])
        self.assertXpath(self.xpath)
        self.assertXpath(self.set_xpath)

        (self.notifications_before_sync,
         self.notifications_after_sync) = self.gNMISubsStreamOnChange(
            self.xpath, self.timeout_secs, changeSet)

        self.assertNotifications()


class Subscribe(SubsOnChangeTestCase):
    """Subscribes ON_CHANGE and checks the response elements.

    1. Subscribes ON_CHANGE to the path.
    1. Collects initial Notifications.
    1. Collects further Notifications until timeout.

    Does not support heartbeat.
    Does not support OnlyUpdates, the initial update is needed for comparison.

    Checks that the paths in the inital Updates are within the Request path.
    Checks that the Update values are different from previous.

    Args:
        xpath: gNMI path to subscribe to.
        timeout_secs: Seconds to wait to receive an Update after Set.
        assert_all_paths_updated: If True, it asserts that all paths in
          the initial Update message had a change after. Defaults to False.
    """
    xpath = ""
    timeout_secs = 10

    def testSubscribeOnChange(self):
        """"""

        self.assertArgs(["xpath"])
        self.assertXpath(self.xpath)

        (self.notifications_before_sync,
         self.notifications_after_sync) = self.gNMISubsStreamOnChange(
            self.xpath, timeout=self.timeout_secs)

        self.assertNotifications()
