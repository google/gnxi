"""Copyright 2021 Google LLC.

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

from oc_config_validate import target


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
        self.assertEqual(target.intersectCmp(a, b), (want_cmp, want_diff),
                         name)


if __name__ == '__main__':
    unittest.main()
