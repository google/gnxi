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

import json
import logging
import time
import unittest
from functools import wraps
from typing import Any, Dict, List, Optional, Tuple, Union

from pyangbind.lib import pybindJSON

from oc_config_validate import context, schema, target
from oc_config_validate.gnmi import gnmi_pb2

# Switch to enable logging of gNMI messages in the testcase logs.
LOG_GNMI = False


def failfast(method):
    """Wrap a test to stop upon failure."""
    @wraps(method)
    def inner(self, *args, **kw):
        if getattr(self, 'failfast', False):
            self.stop()
        method(self, *args, **kw)
    return inner


class TestSuite(unittest.suite.TestSuite):
    """TestSuite from which all other suites inherit.

    Allows inserting arguments to the Tests, which are made attributes of the
        class.
    """

    # Instructs parent class to not cleanup after performing tests.
    _cleanup = False
    test_name = ""
    test_class_name = ""

    def insertArgs(self, args: Dict[str, Any]):
        """Insert test arguments to all test cases in this suite.

        The keys of the argument dictionary are made attributes of the class.

        Args:
            args: dictionary of arguments.
        """
        for test in self._tests:
            TestCase.insertArgs(test, args)

    def insertTarget(self, tgt: target.TestTarget):
        """Insert the gNMI Target to all test cases in this suite.

        Args:
            tgt: A gNMI Target.
        """
        for test in self._tests:
            TestCase.insertTarget(test, tgt)


class TestCase(unittest.case.TestCase):
    """TestCase from which all other tests inherit.

    self.test_target is the gNMI Target to test.

    """
    result = None
    test_target = None
    duration_msec = 0
    start_time_msec = 0

    def phony(self):
        """A phony test method to be used for testing."""

    def run(self, result):
        """Hook the result such that it can be measured and logged."""
        logging.debug('Starting test %s ...', self)
        self.result = result
        if result:
            self.start_time_msec = int(round(time.time() * 1000))
            super().run(result)
            self.duration_msec = (int(round(time.time() * 1000)) -
                                  self.start_time_msec)
            logging.debug('Finished %s in %d msec', self, self.duration_msec)

    def log(self, formatter, *args):
        """Log a message during a test.

        Args:
            formatter: The base python string formatter.
            *args: The formatter parameters. Allows logging messages during
                test executions.
        """
        if self.result:
            self.result.log(formatter, *args)
        logging.debug(formatter, *args)

    def gNMIGet(self, xpath: str) -> Optional[gnmi_pb2.TypedValue]:
        """Send a gNMI Get message to the test's target.

        If the path produces several values in the response, only the first is
          returned.

        Args:
            xpath:  gNMI Path.

        Returns:
            A single response value, or None if error.
        """
        try:
            resp = self.test_target.gNMIGet(xpath)
        except target.RpcError as err:
            if LOG_GNMI:
                self.log("gNMI Get(%s) => ", xpath)
            self.log("gRCP Error: %s", err)
            return None
        resp_val = resp.notification[0].update[0].val
        if LOG_GNMI:
            self.log("gNMI Get(%s) <= %s", xpath, resp_val)
        return resp_val

    def gNMISetUpdate(self, xpath: str, value: Any) -> bool:
        """Send a gNMI Set Update message to the test's target

        Args:
            xpath: The gNMI path to update.
            value: Value to set; can be numeric, string or JSON-IETF.

        Returns:
          False if the gNMI Set did not succeed.
        """
        if LOG_GNMI:
            self.log("gNMI Set Update(%s) => %s", xpath, value)
        try:
            self.test_target.gNMISetUpdate(xpath, value)
        except target.RpcError as err:
            self.log("gRCP Error: %s", err)
            return False
        return True

    def gNMISetDelete(self, xpath: str) -> bool:
        """Send a gNMI Set Delete message to the test's target

        Args:
            xpath: The gNMI path to delete.

        Returns:
          False if the gNMI Set did not succeed.
        """
        if LOG_GNMI:
            self.log("gNMI Set Delete(%s) =>", xpath)
        try:
            self.test_target.gNMISetDelete(xpath)
        except target.RpcError as err:
            self.log("gRCP Error: %s", err)
            return False
        return True

    @classmethod
    def insertArgs(cls, test: unittest.TestCase, args: Dict[str, Any]):
        """Insert test arguments to the test.

        The keys of the argument dictionary are made attributes of the class.

        Args:
            args: dictionary of strings keyed by argument name.
        """
        for key, value in args.items():
            setattr(test, key, value)

    @classmethod
    def insertTarget(cls,
                     test: unittest.TestCase,
                     tgt: target.TestTarget):
        """Insert the gNMI Target to the test.

        Args:
            tgt: A gNMI Target.
        """
        setattr(test, "test_target", tgt)

    def assertArgs(self,
                   args: List[str]):
        """Assert that the arguments are present in the class.

        Args:
            args: A list of argument names to check.

        Raises:
            AssertionError if any argument is not in the class.
        """
        for arg in args:
            self.assertTrue(bool(getattr(self, arg)),
                            "Missing class argument: %s" % arg)

    def assertJsonModel(self, json_value: Union[str, bytes], model: str,
                        msg: str):
        """Assert that the JSON text complies to the model schema.

        Args:
            json_value: JSON-IETF to check.
            model: Python binding class name, to check the JSON reply against.
            mgs: Message to show when the check fails.

        Raises:
          AssertionError if unable to find the model's binding of the JSON
            does not match the model.
        """
        model_obj = schema.containerFromName(model)
        self.assertIsNotNone(model_obj,
                             "Unable to find model '%s' binding" % model)
        match, error = True, None
        try:
            schema.decodeJson(json_value, model_obj)
        except AttributeError as err:
            match, error = False, err
        self.assertTrue(match, "%s: %s" % (msg, error))

    def gNMIGetJson(self, xpath: str) -> Union[str, bytes]:
        """Send a gNMI Get and asserts that the response is a JSON text.

        Args:
            xpath:  gNMI Path.

        Returns:
            a JSON text.

        Raises:
          AssertionError if unable to gNMI Get or if the response is not JSON.
        """
        resp = self.gNMIGet(xpath)
        self.assertIsNotNone(resp, "No gNMI GET response")
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        return resp_val

    def assertJsonCmp(self, json_value: Union[str, bytes], want: dict):
        """Assert that the JSON response matches the wanted values..

        Args:
            json_value: JSON-IETF gNMI response to check.
            want: Dictionary of wanted values.

        Raises:
          AssertionError if unable to compare, of the the wanted values are not
            present in the response.
        """
        got = json.loads(json_value)
        cmp, diff = target.intersectCmp(
            got, json.loads(pybindJSON.dumps(want, mode='ietf')))
        self.assertTrue(cmp, diff)

    def assertXpath(self, xpath: str):
        """Assert that the xpath is valid.

        Args:
            xpath:  gNMI Path.

        Raises:
          AssertionError if xpath is not valid
        """
        self.assertTrue(target.isPathOK(xpath), "'%s' not a gNMI path" % xpath)


class TestResult(unittest.TestResult):
    """TestResult allows logging messages during the tests."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.successes = []
        self.log_message = ''
        self.test_name = 'Test'
        self.duration_msec = 0.0

    def log(self, formatter, *args):
        """Log a message during a test.

        Args:
            formatter: The base python string formatter.
            *args: The formatter parameters. Allows logging messages during
                test executions.
        """
        self.log_message += formatter % args + '\n'

    def startTest(self, test: TestCase):
        """Clean the log before every test."""
        super().startTest(test)
        self.log_message = ''

    @failfast
    def addError(self, test: TestCase, err):
        """Override parent addError to support logging."""
        log = '%s\n%s' % (
            self.log_message, self._exc_info_to_string(err, test))
        self.errors.append((test, log))
        self._mirrorOutput = True
        logging.error("%s - ERR:\n%s", test, log)

    @failfast
    def addFailure(self, test: TestCase, err):
        """Override parent addFailure to support logging."""
        log = '%s\n %s' % (
            self.log_message, self._exc_info_to_string(err, test))
        self.failures.append((test, log))
        self._mirrorOutput = True
        logging.error("%s - FAIL:\n%s", test, log)

    def addSuccess(self, test: TestCase):
        """Store the log messages for every successful test."""
        self.successes.append((test, self.log_message))
        logging.debug("%s - PASS:\n%s", test, self.log_message)

    def addSkip(self, test: TestCase, reason: str):
        """Override parent addFailure to support logging."""
        self.skipped.append((test, reason))
        logging.debug("%s - SKIP:\n%s", test, reason)


class MethodResult():
    """Result of a method called in a TestCase"""

    def __init__(self, test_method: Tuple[TestCase, str], result: str):
        """Create a Result.

        Args:
          test_method: A tuple of TestCase and logs.
          result: A text for the result, as PASS or FAIL.
        """
        test_case, log = test_method
        split_trace = test_case.id().split('.')
        self.test_case = split_trace[-1]
        self.test_class = split_trace[-2]
        self.test_module = '.'.join(split_trace[0:-3])
        self.start_time_msec = test_case.start_time_msec
        self.duration_msec = test_case.duration_msec
        self.result = result
        self.log = log


class TestcaseResult():
    """Result of a TestCase"""

    def __init__(self, test_result):
        """Create a Result.

        Args:
          test_result: A TestResult object.

        """
        self.test_name = test_result.test_name
        self.success = test_result.wasSuccessful()
        self.duration_msec = test_result.duration_msec
        self.results = []
        for m in test_result.successes:
            self.results.append(MethodResult(m, 'PASS'))
        for m in test_result.failures:
            self.results.append(MethodResult(m, 'FAIL'))
        for m in test_result.errors:
            self.results.append(MethodResult(m, 'ERROR'))
        for m in test_result.skipped:
            self.results.append(MethodResult(m, 'SKIPPED'))


class TestRun():
    """Results of a completed test run.

    Including time, context and method details.

    Args:
        target: gNMI Target against which the tests run.
        ctx: Test Context.

    """
    description = ""
    labels = []
    results = []
    start_time_sec = 0.0
    end_time_sec = 0.0
    tests_pass = 0
    tests_total = 0
    tests_fail = 0

    def __init__(self, tgt: str, ctx: context.TestContext):
        self.test_target = tgt
        self.oc_models = schema.getOcModelsVersions()
        if ctx.description:
            self.description = ctx.description
        if ctx.labels:
            self.labels = ctx.labels.copy()

    def copyResults(self, results: List[TestResult], start_time_sec: float,
                    end_time_sec: float):
        """Copy the run test results.

        Args:
            results: A list of test results to copy.
            start_time_sec: EPOCH time when the tests started.
            end_time_sec: EPOCH time when the tests ended.
        """

        self.start_time_sec = start_time_sec
        self.end_time_sec = end_time_sec
        self.results = []
        n_pass = 0
        for result in results:
            n_pass += result.wasSuccessful()
            self.results.append(TestcaseResult(result))
        self.tests_pass = n_pass
        self.tests_total = len(results)
        self.tests_fail = self.tests_total - self.tests_pass

    def summary(self) -> str:
        """Return a text with the results summary. 'PASS x, FAIL y'.
        """

        return "PASS %d, FAIL %d" % (self.tests_pass, self.tests_fail)
