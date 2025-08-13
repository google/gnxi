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

import time
import unittest
from unittest import mock
import re

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
        self.xpaths = ['/network-instances/network-instance[default]']
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
                        gnmi_pb2.PathElem(name='interaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='oper-status')
                    ]),
                    val=gnmi_pb2.TypedValue(string_val='UP')
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth1'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='oper-status')
                    ]),
                    val=gnmi_pb2.TypedValue(string_val='DOWN')
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interaces'),
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
                        gnmi_pb2.PathElem(name='interaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='config'),
                        gnmi_pb2.PathElem(name='enabled')
                    ]),
                    val=gnmi_pb2.TypedValue(bool_val=True)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interaces'),
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
                        gnmi_pb2.PathElem(name='interaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='oper-status')
                    ]),
                    val=gnmi_pb2.TypedValue(string_val='UP')
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interaces'),
                        gnmi_pb2.PathElem(
                            name='interface', key={'name': 'eth0'}),
                        gnmi_pb2.PathElem(name='state'),
                        gnmi_pb2.PathElem(name='mtu')
                    ]),
                    val=gnmi_pb2.TypedValue(int_val=1500)
                ),
                gnmi_pb2.Update(
                    path=gnmi_pb2.Path(elem=[
                        gnmi_pb2.PathElem(name='interaces'),
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
        self.xpaths = []

        with self.assertRaises(AssertionError):
            self.test100()

        self.xpaths = ['/valid/path']
        with self.assertRaises(AssertionError):
            self.test100()

        self.values_type = 'string_val'
        with self.assertRaises(AssertionError):
            self.test100()
        
        self.values_type = None
        self.updates_count = 1
        with self.assertRaises(AssertionError):
            self.test100()

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
            self.test100()

        self.updates_count = 1
        self.xpaths = ['/interaces/interface[name=*]/state/oper-status']
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(AssertionError, "3 != 1 : Expected 1 Updates, got: 3"):
            self.test100()

    def check_bad_type(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the value type is not as expected."""

        self.values_type = 'string_val'
        self.updates_count = 3
        self.xpaths = ['/interaces/interface[name=*]/state/oper-status']
        self.response_interface_status.update[1].val.ClearField('string_val')
        self.response_interface_status.update[1].val.int_val = 1
        
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(
                AssertionError,
                re.compile(r"False is not true : "
                           r"Value of Update /interaces/interface\[name=eth1\]/state/oper-status "
                           r"is not of type string_val: .*")):
            self.test100()

    def check_bad_update_path(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the path of an update is not as expected."""

        self.xpaths = ['/interaces/interface[name=*]/state/oper-status']
        self.values_type = 'string_val'
        self.updates_count = 4
        self.response_interface_status.update.append(
            gnmi_pb2.Update(
                path=gnmi_pb2.Path( 
                    elem=[
                gnmi_pb2.PathElem(name='system'),
                gnmi_pb2.PathElem(name='state'),
                gnmi_pb2.PathElem(name='hostname')],

                ),
                val=gnmi_pb2.TypedValue(string_val='localhost')
            
            ))
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status
        ]
        with self.assertRaisesRegex(
                AssertionError,
                "False is not true : Unexpected update path /system/state/hostname for subscription"):
            self.test100()

    def check_container(self, mock_gNMISubsOnce):
        """Test CountUpdatesCheckType when the path of an update is a container."""

        self.xpaths = ['/interaces/interface[name=eth0]/state']
        self.values_type = 'string_val'
        self.updates_count = 3
        mock_gNMISubsOnce.return_value = [
            self.response_interface_eth0_state
        ]
        with self.assertRaisesRegex(
                AssertionError,
                re.compile(r"False is not true : "
                           r"Value of Update /interaces/interface\[name=eth0\]/state/mtu "
                           r"is not of type string_val: .*")):
            self.test100()

    def check_ok(self, mock_gNMISubsOnce):
        """Test that CountUpdatesCheckType works as expected."""
        self.xpaths = [
            '/system/state/hostname',
            '/interaces/interface[name=*]/state/oper-status']
        self.updates_count = 4
        self.values_type = 'string_val'
        mock_gNMISubsOnce.return_value = [
            self.response_interface_status,
            self.response_hostname
        ]
        self.test100()

        self.xpaths = ['/interaces/interface[name=eth0]/*/enabled']
        self.updates_count = 2
        self.values_type = 'bool_val'
        mock_gNMISubsOnce.return_value = [
            self.response_interface_eth0_enabled
        ]
        self.test100()


if __name__ == '__main__':
    unittest.main()
