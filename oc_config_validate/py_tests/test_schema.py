"""Copyright 2022 Google LLC.

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

import unittest

from parameterized import parameterized

from oc_config_validate import schema
from oc_config_validate.gnmi import gnmi_pb2


class TestIntersectCmp(unittest.TestCase):
    """Test for intersectCmp."""

    abc = {"a": True, "b": 2.5, "c": "str"}
    abcd = {"a": True, "b": 2.5, "c": "str", "d": None}
    ab = {"a": False, "b": 2.5}
    pq = {"p": False, "q": -2.5}
    list1 = [1, 2, 3]
    list2 = [2, 3]
    list3 = [2, abc]

    @parameterized.expand([
        ("allKeys_equal", abc, abc, True, ""),
        ("extraKeys_equal", abc, abcd, True, ""),
        ("missingKeys_equal", abcd, abc, False, "key d not found"),
        ("keys_diff", ab, abc, False, "key a: got 'True', wanted 'False'"),
        ("nested_equal", {"a": abc}, {"a": abcd, "b": True}, True, ""),
        ("nested_equal", {"a": abcd}, {
         "a": abc, "b": True}, False, "key d not found"),
        ("nested_diff", {"a": ab}, {"a": abc, "b": True},
         False, "key a: got 'True', wanted 'False'"),
        ("noKeys", abc, pq, False, "key a not found"),
        ("nested_list_equal", {"a": True, "b": list1},
         {"a": True, "b": list1}, True, ""),
        ("nested_list_equal", {"a": True, "b": list1},
         {"a": True, "b": list1[::-1]}, True, ""),
        ("nested_list_diff", {"a": True, "b": list2},
         {"a": True, "b": list1}, True, ""),
        ("nested_list_diff", {"a": True, "b": list1},  {
         "a": True, "b": list2}, False, "List element 1 not matched"),
        ("nested_list_diff", {"a": True, "b": list1},  {
         "a": True, "b": list3}, False, "List element 1 not matched"),
        ("nested_list_diff", {"a": True, "b": list3},  {"a": True, "b": list1},
         False, "List element {'a': True, 'b': 2.5, 'c': 'str'} not matched"),
        ("nested_list_keys_equal", {"b": [abc]}, {
         "a": True, "b": [abcd, pq]}, True, ""),
        ("nested_list_keys_diff", {"b": [ab]}, {"a": True, "b": [
         abc, pq]}, False, "List element {'a': False, 'b': 2.5} not matched"),
        ("nested_list_keys_diff", {"b": [ab]}, {"a": True, "b": [
         0, pq]}, False, "List element {'a': False, 'b': 2.5} not matched"),
    ])
    def test_intersectCmp(self, name, a, b, want_cmp, want_diff):
        self.assertEqual(schema.intersectCmp(a, b), (want_cmp, want_diff),
                         name)


class TestOcLeafsPaths(unittest.TestCase):
    """Test for ocSubscribePaths."""

    def testInterfaceOcModel(self):
        xpath = "/interfaces/interface[name='test']/state"
        got_paths = schema.ocLeafsPaths(
            "interfaces.openconfig_interfaces", xpath)
        want_paths = [
            "/admin-status",
            "/oper-status",
            "/counters/in-errors",
            "/counters/out-errors"
        ]
        for w in want_paths:
            self.assertIn(xpath + w, got_paths,
                          f"ocLeafsPaths({xpath}) does not contain {w}")


class TestTypedValue(unittest.TestCase):
    """Test for methods dealing with TypedValue."""

    @parameterized.expand([
        ("int", gnmi_pb2.TypedValue(int_val=-3), -3),
        ("str", gnmi_pb2.TypedValue(string_val="3.3"), "3.3"),
        ("bool", gnmi_pb2.TypedValue(bool_val=True), True),
        ("dict", gnmi_pb2.TypedValue(
            json_ietf_val=b'{\"value\": 3}'), {"value": 3})
    ])
    def test_typedValue(self, name, typed_val, val):
        self.assertEqual(schema.typedValueToPython(typed_val), val, name)
        self.assertEqual(schema.pythonToTypedValue(val), typed_val, name)

    def test_typedValue_float(self):
        typed_val = gnmi_pb2.TypedValue(float_val=0.3)
        val = 0.3
        self.assertAlmostEqual(schema.typedValueToPython(typed_val), val)
        self.assertAlmostEqual(schema.pythonToTypedValue(val), typed_val)

    @parameterized.expand([
        ("uint", gnmi_pb2.TypedValue(uint_val=3), 3),
        ("list", gnmi_pb2.TypedValue(
            leaflist_val=gnmi_pb2.ScalarArray(
                element=[gnmi_pb2.TypedValue(int_val=1),
                         gnmi_pb2.TypedValue(int_val=2),
                         gnmi_pb2.TypedValue(int_val=-3)])), [1, 2, -3])
    ])
    def test_typedValueToPython(self, name, typed_val, val):
        self.assertEqual(schema.typedValueToPython(typed_val), val)


class TestParsePaths(unittest.TestCase):
    """Test for methods dealing with Paths."""

    @parameterized.expand([
        ("_full_path",
         "/protocols/protocol[identifier=BGP][name=bgp]/bgp/global",
         True),
        ("_root_path", "/", True),
        ("_basic_path", "/system", True),
        ("_bad_index",
         "/network-instances/network-instance[default]",
         False)
    ])
    def test_pathStrings(self, name, path_str, is_ok):
        self.assertEqual(schema.isPathOK(path_str), is_ok, name)
        if is_ok:
            got_obj = schema.parsePath(path_str)
            self.assertEqual(path_str, schema.pathToString(got_obj), name)


class TestisPathInRequestedPathsRequestedPaths(unittest.TestCase):
    """Test for isPathInRequestedPaths."""

    @parameterized.expand([
        ("_in", "/interfaces/interface[name=ethernet1/2]/state", True),
        ("_deep_in",
         "/interfaces/interface[name=ethernet1/2]/state/counters/out-errors",
         True),
        ("_not_int",
         "/interfaces/interface[name=ethernet1/2]/config/name", False),
    ])
    def test_isPathInRequestedPathsRequestedPaths_singleReq(self, name, xpath,
                                                            want):
        request_xpaths = ["/interfaces/interface[name=ethernet1/2]/state"]
        self.assertEqual(schema.isPathInRequestedPaths(
            xpath, request_xpaths), want)

    @parameterized.expand([
        ("_in", "/interfaces/interface[name=ethernet1/0]/state", True),
        ("_deep_in",
         "/interfaces/interface[name=ethernet1/1]/state/counters/out-errors",
         True),
        ("_not_int",
         "/interfaces/interface[name=ethernet1/2]/config/name", False),
    ])
    def test_isPathInRequestedPathsRequestedPaths_singleWcardReq(self, name,
                                                                 xpath, want):
        request_xpaths = ["/interfaces/interface[name=*]/state"]
        self.assertEqual(schema.isPathInRequestedPaths(
            xpath, request_xpaths), want)

    @parameterized.expand([
        ("_in", "/interfaces/interface[name=ethernet1/0][id=0]/state", True),
        ("_deep_in",
         "/interfaces/interface[id=1][name=1/1]/state/counters/out-errors",
         True),
        ("_not_int",
         "/interfaces/interface[name=ethernet1/2]/config/state", False),
    ])
    def test_isPathInRequestedPathsRequestedPaths_manyKeysReq(self, name,
                                                              xpath, want):
        request_xpaths = ["/interfaces/interface[name=*][id=*]/state"]
        self.assertEqual(schema.isPathInRequestedPaths(
            xpath, request_xpaths), want)
        request_xpaths = ["/interfaces/interface[id=*][name=*]/state"]
        self.assertEqual(schema.isPathInRequestedPaths(
            xpath, request_xpaths), want)


if __name__ == '__main__':
    unittest.main()
