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
from oc_config_validate.testcases import telemetry_once

# Use 'check' instead of 'test', not to mix with test methods in oc_config_validate.testcases
unittest.TestLoader.testMethodPrefix = 'check'
mock.patch.TEST_PREFIX = 'check'


@mock.patch('oc_config_validate.testbase.TestCase.gNMISubsOnce')
class TestSubsOnceTestCase(telemetry_once.SubsOnceTestCase):
    """Test for SubsOnceTestCase class."""

    def setUp(self):
        self.xpaths = ['/valid/path']

    def check_bad_path(self, mock_gNMISubsOnce):
        """Test that subscribeOnce raises an error for bad path."""
        self.xpaths.append('/network-instances/network-instance[default]')
        with self.assertRaises(AssertionError):
            self.subscribeOnce()
        mock_gNMISubsOnce.assert_not_called()

    @parameterized.expand([
        ("none", None),
        ("empty", [])
    ])
    def check_no_responses(self, mock_gNMISubsOnce, name, response):
        """Test that subscribeOnce raises an error when no responses are received."""
        mock_gNMISubsOnce.return_value = response
        with self.assertRaisesRegex(AssertionError, "No gNMI Subscribe response"):
            self.subscribeOnce()
        self.assertTrue(mock_gNMISubsOnce.called)

    @parameterized.expand([
        ("want_1_got_3", 1, 3),
        ("want_3_got_1", 3, 1)
    ])
    def check_bad_notifications(self, mock_gNMISubsOnce, name, want, got):
        """Test that subscribeOnce raises an error when the notifications are not as expected."""
        self.notifications_count = want
        mock_gNMISubsOnce.return_value = [
            gnmi_pb2.Notification(
                timestamp=(int(time.time()) + 10) * 1000000000)  # 10 second later
        ] * got
        with self.assertRaisesRegex(AssertionError, f"Expected {want} notifications, got {got}"):
            self.subscribeOnce()

    def check_bad_update_path(self, mock_gNMISubsOnce):
        """Test subscribeOnce when the path of an update is not as subscribed."""

        self.xpaths = ['/interfaces/interface[name=*]/state/oper-status']
        mock_gNMISubsOnce.return_value = [
            gnmi_pb2.Notification(
                timestamp=(int(time.time()) + 10) *
                1000000000,  # 10 second later
                update=[
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
                        path=gnmi_pb2.Path(
                            elem=[
                                gnmi_pb2.PathElem(name='system'),
                                gnmi_pb2.PathElem(name='state'),
                                gnmi_pb2.PathElem(name='hostname')],

                        ),
                        val=gnmi_pb2.TypedValue(string_val='localhost')

                    )
                ])

        ]

        with self.assertRaisesRegex(
                AssertionError,
                "False is not true : "
                "Unexpected update path /system/state/hostname for subscription"):
            self.subscribeOnce()

    def check_ok_nothing_checked(self, mock_gNMISubsOnce):
        """Test that subscribeOnce does NOT check Notifications nor delays if not told so."""
        mock_gNMISubsOnce.return_value = [
            gnmi_pb2.Notification(
                timestamp=(int(time.time()) + 30) * 1000000000),
            gnmi_pb2.Notification(
                timestamp=(int(time.time()) + 35) * 1000000000)
        ]
        self.subscribeOnce()

        self.max_delay_secs = 3
        with self.assertRaisesRegex(AssertionError, "Timestamp diff too long: 30 secs"):
            self.subscribeOnce()

        self.max_delay_secs = None
        self.notifications_count = 1
        with self.assertRaisesRegex(AssertionError, "Expected 1 notifications, got 2"):
            self.subscribeOnce()


@mock.patch('oc_config_validate.testbase.TestCase.gNMISubsOnce')
class TestCountUpdatesCheckType(telemetry_once.CountUpdatesCheckType):
    """Test for CountUpdatesCheckType class."""

    def setUp(self):

        self.response_hostname = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='system'),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='hostname')
                    ]),
                    val=gnmi_pb2.TypedValue(string_val='localhost')
                )
            ]
        )

        self.response_interface_status = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
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
        )
        self.response_interface_eth0_enabled = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='config'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=True)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=False)
                )
            ]
        )

        self.response_interface_eth0_state = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
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
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='mtu')
                    ]),
                    val=gnmi_pb2.TypedValue(int_val=1500)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=True)
                )
            ]
        )

    def check_bad_args(self, mock_gNMISubsOnce):
        """Test that CountUpdatesCheckType raises an error for missing arguments."""
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        self.xpaths = ['/valid/path']
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        self.values_type = 'string_val'
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        self.values_type = None
        self.updates_count = 1
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        mock_gNMISubsOnce.assert_not_called()

    def check_bad_count_updates(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the updates are not as expected."""
        self.updates_count = 3
        self.values_type = 'string_val'
        self.xpaths = ['/system/state/hostname']
        mock_gNMISubsOnce.return_value = [
            self.response_hostname
        ]
        with self.assertRaisesRegex(AssertionError, "1 != 3 : Expected 3 Updates, got: 1"):
            self.testSubscribeOnce()

        self.updates_count = 1
        self.xpaths = ['/interfaces/interface[name=*]/state/oper-status']
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(AssertionError, "3 != 1 : Expected 1 Updates, got: 3"):
            self.testSubscribeOnce()

    def check_bad_type(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the value type is not as expected."""

        self.values_type = 'string_val'
        self.updates_count = 3
        self.xpaths = ['/interfaces/interface[name=*]/state/oper-status']
        self.response_interface_status.update[1].val.ClearField('string_val')
        self.response_interface_status.update[1].val.int_val = 1

        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(
                AssertionError,
                re.compile(r"False is not true : "
                           r"Value of Update /interfaces/interface\[name=eth1\]/state/oper-status "
                           r"is not of type string_val: .*")):
            self.testSubscribeOnce()

    def check_container(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the path of an update is a container."""

        self.xpaths = ['/interfaces/interface[name=eth0]/state']
        self.values_type = 'string_val'
        self.updates_count = 3
        mock_gNMISubsOnce.return_value = [
            self.response_interface_eth0_state
        ]
        with self.assertRaisesRegex(
                AssertionError,
                re.compile(r"False is not true : "
                           r"Value of Update /interfaces/interface\[name=eth0\]/state/mtu "
                           r"is not of type string_val: .*")):
            self.testSubscribeOnce()

    def check_wildcards_ok(self, mock_gNMISubsOnce):
        """Test that CountUpdatesCheckType works as expected."""
        self.xpaths = [
            '/system/state/hostname',
            '/interfaces/interface[name=*]/state/oper-status']
        self.updates_count = 4
        self.values_type = 'string_val'
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status,
            self.response_hostname
        ]
        self.testSubscribeOnce()

        self.xpaths = ['/interfaces/interface[name=eth0]/*/enabled']
        self.updates_count = 2
        self.values_type = 'bool_val'
        mock_gNMISubsOnce.return_value = [
            self.response_interface_eth0_enabled
        ]
        self.testSubscribeOnce()

@mock.patch('oc_config_validate.testbase.TestCase.gNMISubsOnce')
class TestCountUpdatePaths(telemetry_once.CountUpdatePaths):
    """Test for CountUpdatePaths class."""

    def setUp(self):

        self.response_hostname = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='system'),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='hostname')
                    ]),
                    val=gnmi_pb2.TypedValue(string_val='localhost')
                )
            ]
        )

        self.response_interface_status = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
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
        )
        self.response_interface_eth0_enabled = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='config'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=True)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=False)
                )
            ]
        )

        self.response_interface_eth0_state = gnmi_pb2.Notification(
            timestamp=(int(time.time())) * 1000000000,
            update=[
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
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='mtu')
                    ]),
                    val=gnmi_pb2.TypedValue(int_val=1500)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interfaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=True)
                )
            ]
        )

    def check_bad_args(self, mock_gNMISubsOnce):
        """Test that CountUpdatePaths raises an error for missing argument."""
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        self.xpaths = ['/valid/path']
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        self.update_paths_count = 3
        self.xpaths = None
        with self.assertRaises(AssertionError):
            self.testSubscribeOnce()

        mock_gNMISubsOnce.assert_not_called()

    def check_diff_count_updates(self, mock_gNMISubsOnce):
        """Test CountUpdatePaths when the updates are not as expected."""
        self.update_paths_count = 4
        self.xpaths = ['/interfaces/interface[name=*]/state/oper-status']
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(
            AssertionError, 
            re.compile(
            r"3 != 4 : Expected 4 Update paths, got: .*")            ):
            self.testSubscribeOnce()

        self.update_paths_count = 1
        self.xpaths = ['/system/state']
        mock_gNMISubsOnce.return_value = [
            self.response_hostname,
            self.response_hostname,
        ]
        self.testSubscribeOnce()

    def check_wildcards_ok(self, mock_gNMISubsOnce):
        """Test that CountUpdatesCheckType works as expected."""
        self.xpaths = [
            '/interfaces/interface[name=eth0]/*/enabled',
            '/interfaces/interface[name=*]/state/oper-status']
        self.update_paths_count = 5
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status,
            self.response_interface_eth0_enabled,
            self.response_interface_status,
        ]
        self.testSubscribeOnce()


if __name__ == '__main__':
    unittest.main()
