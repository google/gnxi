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

import unittest

from oc_config_validate import context, runner, target, testcases


class TestGetTestSuites(unittest.TestCase):
    """Test for method runner.getTestSuites()."""

    def test_getTestSuites(self):
        tgt = target.TestTarget(context.Target())
        tests = [
            context.TestCase("SetConfigCheckState",
                             "config_state.SetConfigCheckState",
                             None),
            context.TestCase("AddStaticRoute",
                             "static_route.AddStaticRoute",
                             {"arg1": "foo", "arg2": "bar"}),
        ]
        ctx = context.TestContext("example", [], [], None, tests)
        suites = runner.getTestSuites(ctx, tgt)
        self.assertIsInstance(
            suites[0]._tests[0], testcases.config_state.SetConfigCheckState)
        self.assertIsInstance(suites[1]._tests[0],
                              testcases.static_route.AddStaticRoute)
        for s in suites:
            self.assertEqual(s._tests[0].test_target, tgt)
        for t in suites[1]._tests:
            self.assertEqual(getattr(t, "arg1", None), "foo")
            self.assertEqual(getattr(t, "arg2", None), "bar")

    def test_getTestSuites_error(self):
        tgt = target.TestTarget(context.Target())
        tests = [
            context.TestCase("notAClass",
                             "config_state.notAClass",
                             None),
            context.TestCase("NotAModule",
                             "not_a_module.aClass",
                             None),
        ]
        ctx = context.TestContext("example", [], [], None, tests)
        suites = runner.getTestSuites(ctx, tgt)
        self.assertEqual(suites, [])


if __name__ == '__main__':
    unittest.main()
