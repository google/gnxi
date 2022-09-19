"""Test cases based on gNMI Subscribe SAMPLE requests."""

from oc_config_validate import testbase


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
            drift = time_diff - interval
            drift = drift * (-1) if drift < 0 else drift
            return (drift < max_drift)

        self.assertArgs(["xpath", "sample_interval", "sample_timeout"])
        self.assertXpath(self.xpath)

        self.responses = self.gNMISubsStreamSample(
            self.xpath,
            self.sample_interval * 1000000000,
            timeout=self.sample_timeout)
        self.assertIsNotNone(self.responses, "No gNMI Subscribe response")

        want = (self.sample_timeout // self.sample_interval) + 1
        if self.responses:
            for path, updates in self.responses.items():
                got = len(updates)
                self.assertEqual(
                    got, want,
                    f"{got} Updates for path {path}, wanted {want}")
                if len(updates) > 1:
                    timestamp_0, _ = updates[0]
                    timeline = [timestamp_0]
                    for timestamp_1, _ in updates[1:]:
                        timestamp_diff = (
                            timestamp_1 - timestamp_0) // 1000000000
                        timeline.append(timestamp_1)
                        self.assertTrue(
                            _isTimestampDiffOk(timestamp_diff,
                                               self.sample_interval,
                                               self.max_timestamp_drift_secs),
                            f"Update timestamps for '{path}' out of interval: "
                            f"{timeline}")
                        timestamp_0 = timestamp_1


class CountUpdates(SubsSampleTestCase):
    """Subscribes SAMPLE and checks the returned update paths.

    This tests that a Subscription consistenly reports all Update paths.

    Args:
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
