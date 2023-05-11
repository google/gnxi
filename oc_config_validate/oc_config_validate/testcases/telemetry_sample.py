"""Test cases based on gNMI Subscribe SAMPLE requests."""

import collections

from oc_config_validate import schema, testbase

SECS_TO_NSEC = 1000000000


class SubsSampleTestCase(testbase.TestCase):
    """Subscribes SAMPLE and checks the response elements.

    Asserts that the Updates timestamps correspond to the sample interval,
        with a maximum tolerated drift.
    Asserts that all Update paths have the same number of updates.

    Args:
        xpath: gNMI path to subscribe to.
        sample_interval: Seconds interval between Updates.
        sample_timeout: Seconds to keep the Subscription and collect Updates.
        max_timestamp_drift_secs: Maximum drift for the timestamp(s) in the
            reply.
    """
    max_timestamp_drift_secs = 1
    xpath = None
    sample_interval = None
    sample_timeout = None

    def subscribeSample(self):
        """"""

        def _isTimestampDiffOk(time_diff, interval, max_drift):
            """Returns true if the timestamp differences is within limits.

            Timestamp diff must be interval +- max_drift.

            Args:
              time_diff: Timestamp diff in nsecs.
              interval: Sampling interval in nsecs.
              max_drift: Maximum drift in nsecs.

            """
            drift = time_diff - interval
            drift = drift * (-1) if drift < 0 else drift
            return (drift <= max_drift)

        self.assertArgs(["xpath", "sample_interval", "sample_timeout"])
        self.assertXpath(self.xpath)

        notifications = self.gNMISubsStreamSample(
            self.xpath,
            self.sample_interval * SECS_TO_NSEC,
            timeout=self.sample_timeout)
        self.assertIsNotNone(notifications, "No gNMI Subscribe response")

        self.responses = collections.defaultdict(list)
        for n in notifications:
            timestamp = n.timestamp
            for u in n.update:
                self.responses[schema.pathToString(u.path)].append(
                    (timestamp, schema.typedValueToPython(u.val)))

        want = (self.sample_timeout // self.sample_interval) + 1
        if self.responses:
            for path, updates in self.responses.items():
                got = len(updates)
                self.assertEqual(
                    got, want,
                    f"{got} Updates for path {path}, wanted {want}")
                if len(updates) > 1:
                    timestamp_0, _ = updates[0]
                    timeline = [0]
                    for timestamp_1, _ in updates[1:]:
                        timestamp_diff = timestamp_1 - timestamp_0
                        timeline.append(timestamp_diff)
                        self.assertTrue(
                            _isTimestampDiffOk(
                                timestamp_diff,
                                self.sample_interval * SECS_TO_NSEC,
                                self.max_timestamp_drift_secs * SECS_TO_NSEC),
                            f"Update timestamps for '{path}' out of interval: "
                            f"{timeline} diff in nsecs")
                        timestamp_0 = timestamp_1


class CountUpdates(SubsSampleTestCase):
    """Subscribes SAMPLE and checks the returned update paths.

    This tests that a Subscription consistenly reports all Update paths.

    Args:
        xpath: gNMI paths to subscribe to. Can contain wildcards.
        updates_count: Number of expected disctinct Update paths, per reply.
    """
    update_paths_count = None

    def testSubscribeSample(self):
        """"""
        self.assertArgs(["update_paths_count"])
        self.subscribeSample()
        got_paths = len(self.responses)
        self.assertEqual(
            got_paths, self.update_paths_count,
            f"Expected {self.update_paths_count} Update paths, "
            f"got: {got_paths}")


class CheckLeafs(SubsSampleTestCase):
    """Subscribes SAMPLE and checks the updates againts the state OC model.

    Tests that a Sample Subscription consistenly reports all Update paths, and
    that the paths corresponds to Leafs in the OC model.

    Args:
        xpath: gNMI paths to subscribe to. Can contain wildcards only on the
          keys.
        model: Python binding class to check the replies against.
    """
    model = None

    def testSubscribeSample(self):
        """"""
        self.assertArgs(["model", "xpath"])

        self.assertModelXpath(self.model, self.xpath)

        self.subscribeSample()

        got_paths = len(self.responses)
        self.assertGreater(got_paths, 0,
                           "There are no Update replies to the Subscription")
        if self.responses:
            for path, _ in self.responses.items():
                self.assertTrue(
                    schema.isPathInRequestedPaths(path, [self.xpath]),
                    f"Unexpected update path {path} for subscription")
