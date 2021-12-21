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

    @parameterized.expand([
        ("allKeys_equal", {"a": True, "b": 2.5, "c": "str"},
         {"a": True, "b": 2.5, "c": "str"}, True, ""),
        ("allKeys_diffType", {"a": True, "b": 2.5, "c": "str"},
         {"a": True, "b": "str", "c": "str"}, False, "key b: 2.5 != str"),
        ("allKeys_diffValue", {"a": True, "b": 2.5, "c": "str"},
         {"a": True, "b": 1.5, "c": "str"}, False, "key b: 2.5 != 1.5"),
        ("nested_equal", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"a": True, "b": {"b": 2.5, "c": "str"}}, True, ""),
        ("nested_diffType", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"a": True, "b": {"b": "str", "c": "str"}}, False,
         "key b: 2.5 != str"),
        ("nested_diffValue", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"a": True, "b": {"b": 2.5, "c": "notstr"}}, False,
         "key c: str != notstr"),
        ("someKeys_equal", {"a": True, "b": 2.5, "c": "str"},
         {"a": True, "c": "str", "d": 2.5}, True, ""),
        ("someKeys_diffType", {"a": {}, "b": 2.5, "c": "str"},
         {"a": True, "c": "str", "d": 2.5}, False, "key a: {} != True"),
        ("someKeys_diffValue", {"a": False, "b": 2.5, "c": "str"},
         {"a": True, "c": "str", "d": 2.5}, False, "key a: False != True"),
        ("nested_someKs_equal", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"z": True, "b": {"c": "str", "d": 2.5}}, True, ""),
        ("nested_someKS_diffType", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"z": True, "b": {"c": False, "d": 2.5}}, False,
         "key c: str != False"),
        ("nested_someKs_diffValue", {"a": True, "b": {"b": 2.5, "c": "str"}},
         {"z": True, "b": {"c": "nostr", "d": 2.5}}, False,
         "key c: str != nostr"),
        ("noKeys", {"a": True, "b": 2.5},
         {"c": True, "d": 2.5}, False,
         "got dict_keys(['a', 'b']), wanted dict_keys(['c', 'd'])")
    ])
    def test_intersectCmp(self, name, a, b, want_cmp, want_diff):
        self.assertEqual(target.intersectCmp(a, b), (want_cmp, want_diff),
                         name)


if __name__ == '__main__':
    unittest.main()
