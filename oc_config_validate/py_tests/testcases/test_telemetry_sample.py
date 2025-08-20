"""Copyright 2023 Google LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

You may obtain a copy of the License at
                https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

"""

import re
import time
import unittest
from unittest import mock

from parameterized import parameterized

from oc_config_validate.gnmi import gnmi_pb2  # type: ignore
from oc_config_validate.testcases import telemetry_sample

# Use 'check' instead of 'test', not to mix with test methods in oc_config_validate.testcases
unittest.TestLoader.testMethodPrefix = 'check'
mock.patch.TEST_PREFIX = 'check'


@mock.patch('oc_config_validate.testbase.TestCase.gNMISubsStreamSample')
class TestSubsSampleTestCase(telemetry_sample.SubsSampleTestCase):
    """Test for SubsSampleTestCase class."""

    def setUp(self):
        self.xpath = '/interfaces/interface[name=*]/state/oper-status'
        self.sample_interval = 10
        self.sample_timeout = 35

        # In Setup, so every test has a known set of updates to work with.
        self.updates_interface_status = [
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'eth0'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='UP')
            ),
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'eth1'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='DOWN')
            ),
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'mgmt'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='UP')
            )
        ]

    def check_bad_args(self, mock_gNMISubsStreamSample):
        """Test that subscribeSample raises an error for missing arguments."""
        self.xpath = None
        self.sample_interval = None
        self.sample_timeout = None

        with self.assertRaises(AssertionError):
            self.subscribeSample()

        self.xpath = '/valid/path'
        with self.assertRaises(AssertionError):
            self.subscribeSample()

        self.sample_interval = 10
        with self.assertRaises(AssertionError):
            self.subscribeSample()

        mock_gNMISubsStreamSample.assert_not_called()

        self.sample_timeout = 30
        self.subscribeSample()

    def check_bad_path(self, mock_gNMISubsStreamSample):
        """Test that subscribeSample raises an error for bad path."""
        self.xpath = '/network-instances/network-instance[default]'
        with self.assertRaises(AssertionError):
            self.subscribeSample()
        mock_gNMISubsStreamSample.assert_not_called()

    @parameterized.expand([
        ("none", None),
        ("empty", [])
    ])
    def check_no_responses(self, mock_gNMISubsStreamSample, name, response):
        """Test that subscribeSample raises an error when no responses are received."""
        mock_gNMISubsStreamSample.return_value = response
        with self.assertRaisesRegex(AssertionError, "No gNMI Subscribe response"):
            self.subscribeSample()
        self.assertTrue(mock_gNMISubsStreamSample.called)

    def check_bad_updates(self, mock_gNMISubsStreamSample):
        """Test that subscribeSample raises an error when missing updates for a path."""

        self.max_timestamp_drift_secs = 10
        now = int(time.time())
        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 11) * 1000000000,
                update=self.updates_interface_status[1:]
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 22) * 1000000000,
                update=self.updates_interface_status
            )
        ]
        with self.assertRaisesRegex(
            AssertionError,
            re.compile(
                r"2 != 4 : 2 Updates for /interfaces/interface\[name=eth0\]/state/oper-status, "
                r"wanted 4")):
            self.subscribeSample()

        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 8) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 16) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 24) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 32) * 1000000000,
                update=self.updates_interface_status
            )
        ]
        with self.assertRaisesRegex(
            AssertionError,
            re.compile(
                r"5 != 4 : 5 Updates for /interfaces/interface\[name=eth0\]/state/oper-status, "
                r"wanted 4")):
            self.subscribeSample()

    def check_bad_timestamp(self, mock_gNMISubsStreamSample):
        """Test that subscribeSample raises an error when timestamps are not within limits."""
        now = int(time.time())
        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 10) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 22) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 30) * 1000000000,
                update=self.updates_interface_status
            )
        ]
        with self.assertRaisesRegex(
            AssertionError,
            re.compile(
                r"False is not true : "
                r"Subscribe updates for '/interfaces/interface\[name=eth0\]/state/oper-status' "
                r"received out of sample interval 10 secs, "
                r"received updates intervals: \[10.0, 12.0\]")):
            self.subscribeSample()

        self.max_timestamp_drift_secs = 5
        self.subscribeSample()

    def check_bad_update_path(self, mock_gNMISubsStreamSample):
        """Test subscribeOnce when the path of an update is not as subscribed."""

        now = int(time.time())
        mock_gNMISubsStreamSample.return_value = mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 11) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 22) * 1000000000,
                update=self.updates_interface_status + [
                    gnmi_pb2.Update(
                        path=gnmi_pb2.Path(elem=[
                            gnmi_pb2.PathElem(name='system'),
                            gnmi_pb2.PathElem(name='state'),
                            gnmi_pb2.PathElem(name='hostname')
                        ]),
                        val=gnmi_pb2.TypedValue(string_val='foo')
                    )
                ]
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 31) * 1000000000,
                update=self.updates_interface_status
            ),
        ]

        with self.assertRaisesRegex(
                AssertionError,
                "False is not true : "
                "Unexpected update path /system/state/hostname for subscription"):
            self.subscribeSample()

    def check_ok(self, mock_gNMISubsStreamSample):
        """Test that subscribeSample works as expected."""

        now = int(time.time())
        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 11) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 20) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 29) * 1000000000,
                update=self.updates_interface_status
            )
        ]
        self.subscribeSample()


@mock.patch('oc_config_validate.testbase.TestCase.gNMISubsStreamSample')
class TestCountUpdatePaths(telemetry_sample.CountUpdatePaths):
    """Test for CountUpdatePaths class."""

    def setUp(self):
        self.sample_interval = 10
        self.sample_timeout = 30
        self.xpath = '/interfaces/interface[name=*]/state/oper-status'

        # In Setup, so every test has a known set of updates to work with.
        self.updates_interface_status = [
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'eth0'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='UP')
            ),
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'eth1'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='DOWN')
            ),
            gnmi_pb2.Update(
                path=gnmi_pb2.Path(elem=[
                    gnmi_pb2.PathElem(name='interfaces'),
                    gnmi_pb2.PathElem(
                        name='interface', key={'name': 'mgmt'}),
                    gnmi_pb2.PathElem(name='state'),
                    gnmi_pb2.PathElem(name='oper-status')
                ]),
                val=gnmi_pb2.TypedValue(string_val='UP')
            )
        ]

    def check_bad_args(self, mock_gNMISubsStreamSample):
        """Test that CountUpdatePaths raises an error for missing arguments."""

        with self.assertRaises(AssertionError):
            self.testSubscribeSample()

        mock_gNMISubsStreamSample.assert_not_called()

    def check_bad_count_updates(self, mock_gNMISubsStreamSample):
        """Test CountUpdatePaths when the update paths are not as expected."""
        self.update_paths_count = 3

        now = int(time.time())

        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status[2:]
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 10) * 1000000000,
                update=self.updates_interface_status[2:]
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 20) * 1000000000,
                update=self.updates_interface_status[2:]
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 30) * 1000000000,
                update=self.updates_interface_status[2:]
            )
        ]
        with self.assertRaisesRegex(
            AssertionError,
            re.compile(
                r"1 != 3 : Expected 3 Update paths, "
                r"got: \['/interfaces/interface\[name=mgmt\]/state/oper-status'\]")):
            self.testSubscribeSample()

    def check_ok(self, mock_gNMISubsStreamSample):
        """Test that CountUpdatePaths works as expected."""
        self.update_paths_count = 3
        now = int(time.time())

        mock_gNMISubsStreamSample.return_value = [
            gnmi_pb2.Notification(
                timestamp=now * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 11) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 20) * 1000000000,
                update=self.updates_interface_status
            ),
            gnmi_pb2.Notification(
                timestamp=(now + 30) * 1000000000,
                update=self.updates_interface_status
            )
        ]

        self.testSubscribeSample()


if __name__ == '__main__':
    unittest.main()
