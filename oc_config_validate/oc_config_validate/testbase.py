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

import copy
import json
import logging
import time
import unittest
from functools import wraps
from typing import Any, Dict, List, Optional, Tuple, Union

from pyangbind.lib import pybindJSON
from pyangbind.lib.base import PybindBase
from retry.api import retry_call

from oc_config_validate import context, schema, target
from oc_config_validate.gnmi import gnmi_pb2  # type: ignore

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


def retryAssertionError(method):
    """Wrap a test to retry if AssertionError."""
    @wraps(method)
    def inner(self, *args, **kw):
        if hasattr(self, 'retries'):
            delay = getattr(self, 'retry_delay', 10)
            retry_call(method,  fargs=[self, *args], fkwargs=kw,
                       tries=getattr(self, 'retries'), delay=delay)
        else:
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
    duration_sec = 0
    start_time_sec = 0

    def phony(self):
        """A phony test method to be used for testing."""

    def run(self, result=None):
        """Hook the result such that it can be measured and logged."""
        logging.debug('Starting test %s ...', self)
        self.result = result
        if result:
            self.start_time_sec = int(time.time())
            super().run(result)
            self.duration_sec = int(time.time() - self.start_time_sec)
            logging.debug('Finished %s in %d secs', self, self.duration_sec)

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
        except Exception as err:
            self.log("Get(%s) <= Error: %s", xpath, err)
            return None
        try:
            resp_val = resp.notification[0].update[0].val
        except IndexError:
            self.log("Get(%s) <= Bad response: %s", xpath, resp)
            return None
        if LOG_GNMI:
            msg = ("gNMI Get(%s) <= %s", xpath, resp_val)
            self.log(*msg)
            logging.info(*msg)
        return resp_val

    def gNMIGetAssertJson(self, xpath: str) -> Union[str, bytes]:
        """Send a gNMI Get and asserts that the response is a JSON text.

        Args:
            xpath:  gNMI Path.

        Returns:
            a JSON text.

        Raises:
          AssertionError if unable to gNMI Get or if the response is not JSON.
        """
        resp = self.gNMIGetAssertGet(xpath)
        resp_val = resp.json_ietf_val
        self.assertIsNotNone(resp_val,
                             "The gNMI GET response is not JSON IETF")
        return resp_val

    def gNMIGetAssertBool(self, xpath: str) -> bool:
        """Send a gNMI Get and asserts that the response is a bool value.

        Args:
            xpath:  gNMI Path.

        Returns:
            a boolean value.

        Raises:
          AssertionError if unable to gNMI Get or if the response is not bool.
        """
        resp = self.gNMIGetAssertGet(xpath)
        self.assertIsNotNone(
            resp.bool_val,
            "The gNMI Get response is not a boolean: %s" % resp)
        return resp

    def gNMIGetAssertGet(self, xpath: str) -> gnmi_pb2.TypedValue:
        """Send a gNMI Get and asserts that there is a response.

        Args:
            xpath:  gNMI Path.

        Returns:
            the response to the Get requets.

        Raises:
          AssertionError if unable to gNMI Get or if the response is None.
        """
        resp = self.gNMIGet(xpath)
        self.assertIsNotNone(resp, "No gNMI GET response")
        return resp

    def gNMISetUpdate(self, xpath: str, value: Any) -> bool:
        """Send a gNMI Set Update message to the test's target

        Args:
            xpath: The gNMI path to update.
            value: Value to set; can be numeric, string or JSON-IETF.

        Returns:
          False if the gNMI Set did not succeed.
        """
        if LOG_GNMI:
            msg = ("gNMI Set Update(%s) => %s", xpath, value)
            self.log(*msg)
            logging.info(*msg)
        try:
            self.test_target.gNMISetUpdate(xpath, value)
        except Exception as err:
            self.log("Set(%s) <= Error: %s", xpath, err)
            return False
        return True

    def gNMISetDelete(self, xpath: str) -> bool:
        """Send a gNMI Set Delete message to the test's target

        Args:
            xpath: The gNMI path to delete.

        Args:
            xpath:  gNMI Path.

        Returns:
            A single response value, or None if error.
        """
        if LOG_GNMI:
            msg = ("gNMI Set Delete(%s) =>", xpath)
            self.log(*msg)
            logging.info(*msg)
        try:
            self.test_target.gNMISetDelete(xpath)
        except Exception as err:
            self.log("Set(%s) <= Error: %s", xpath, err)
            return False
        return True

    def gNMISubsOnce(self, xpaths: List[str]) -> Optional[
            List[gnmi_pb2.Notification]]:
        """Send a gNMI Subscribe message using ONCE mode.

        Gets all the Notification messages that came as response.

        Args:
            xpaths: List of gNMI paths to subscribe to.

        Returns:
          A list of Notifications received, or None if error.
        """
        try:
            resp = self.test_target.gNMISubsOnce(xpaths)
        except Exception as err:
            self.log("SubscribeOnce(%s) <= Error: %s", xpaths, err)
            return None
        if LOG_GNMI:
            msg = ("gNMI SubscribeOnce(%s) <= %s", xpaths,
                   schema.notificationsJsonString(resp))
            self.log(*msg)
            logging.info(*msg)
        return resp

    def gNMISubsStreamSample(
            self, xpath: str, sample_interval: int, timeout: int) -> Optional[
            List[gnmi_pb2.Notification]]:
        """Send a gNMI Subscribe message, using STREAM SAMPLE mode.

        Gets all the Notification messages that came as response. It returns
            a dictionary of all the Updates, keyed by Update path.

        Args:
            xpath: gNMI path to subscribe to.
            sample_interval: Nanoseconds between updates.
            timeout: Seconds to keep the gRPC channel open and receive
                    updates.

        Returns:
          A list of Notifications received, or None if error.
        """
        try:
            resp = self.test_target.gNMISubsStreamSample(
                xpath, sample_interval, timeout)
        except Exception as err:
            self.log("SubscribeStream(%s) <= Error: %s", xpath, err)
            return None
        if LOG_GNMI:
            msg = ("gNMI SubscribeSample(%s) <= %s", xpath,
                   schema.notificationsJsonString(resp))
            self.log(*msg)
            logging.info(*msg)
        return resp

    def gNMISubsStreamOnChange(
            self,
            xpath: str,
            timeout: int,
            sync_response_callback=None) -> Tuple[
                Optional[List[gnmi_pb2.Notification]],
                Optional[List[gnmi_pb2.Notification]]]:
        """Subscribes on-change and returns the Notifications received.

        It returns the Updates before and after receiving a sync_response.

        All Updates received are accumulated up to when the channel is
         closed (by server, by timeout or error).

        Optionally, it calls a callback function when sync_response is
          received.

        Does not support heartbeat. nor OnlyUpdates (the initial update
          is needed for value change comparison).

        Args:
            xpath: gNMI path to subscribe to.
            timeout: Seconds to keep the gRPC channel open.
            sync_response_callback: Function to call upon sync_response.

        Returns:
            2 Lists of gnmi_pb2.Notification objects received,
              before and after sync_response.
        """

        try:
            before, after = self.test_target.gNMISubsStreamOnChange(
                xpath, timeout, sync_response_callback)
        except Exception as err:
            self.log("SubscribeOnChange(%s) <= Error: %s", xpath, err)
            return None, None
        if LOG_GNMI:
            msg = ("gNMI SubscribeOnChange(%s) <= %s\n", xpath,
                   schema.notificationsJsonString(before + after))
            self.log(*msg)
            logging.info(*msg)
        return before, after

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
            self.assertIsNotNone(getattr(self, arg, None),
                                 "Missing class argument: %s" % arg)

    def assertJsonModel(self, json_value: Union[str, bytes],
                        model: PybindBase, msg: str):
        """Assert that the JSON text complies to the model schema.

        Args:
            json_value: JSON-IETF to check.
            model: the OC model class name in the oc_config_validate.models
              package, as `module.class`.
            xpath: the xpath to the OC container to create.
            mgs: Message to show when the check fails.

        Raises:
          AssertionError if the JSON text does not match the model.
        """
        match, error = True, None
        model_obj = copy.copy(model)
        try:
            schema.decodeJson(json_value, model_obj)
        except (AttributeError, ValueError) as err:
            match, error = False, err
        self.assertTrue(match, "%s: %s" % (msg, error))

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
        cmp, diff = schema.intersectCmp(
            json.loads(pybindJSON.dumps(want, mode='ietf')), got)
        self.assertTrue(cmp, diff)

    def assertXpath(self, xpath: str):
        """Assert that the xpath is valid.

        Args:
            xpath:  gNMI Path.

        Raises:
          AssertionError if xpath is not valid
        """
        self.assertTrue(schema.isPathOK(xpath), "'%s' not a gNMI path" % xpath)

    def assertModelXpath(self, model: str, xpath: str):
        """Assert that the model and xpath correspond to a valid OC container.

        Args:
            model: the OC model class name in the oc_config_validate.models
              package, as `module.class`.
            xpath: the xpath to the OC container to create.

        Raises: AssertionError.
        """
        match, error = True, None
        try:
            schema.ocContainerFromPath(model, xpath)
        except (AttributeError, schema.BaseError) as err:
            match, error = False, err
        self.assertTrue(
            match,
            "xpath %s does not point to a container in model %s: %s" % (
                xpath, model, error))


class TestResult(unittest.TestResult):
    """TestResult allows logging messages during the tests."""

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.successes = []
        self.log_message = ''
        self.test_name = 'Test'
        self.duration_sec = 0

    def log(self, formatter, *args):
        """Log a message during a test.

        Args:
            formatter: The base python string formatter.
            *args: The formatter parameters. Allows logging messages during
                test executions.
        """
        self.log_message += formatter % args + '\n'

    def startTest(self, test: TestCase):  # pytype: disable=signature-mismatch
        """Clean the log before every test."""
        super().startTest(test)
        self.log_message = ''

    @failfast
    def addError(self, test: TestCase, err):
        """Override parent addError to support logging."""
        log = '%s\n%s' % (
            self.log_message, self._exc_info_to_string(err, test))  # pytype: disable=attribute-error # noqa
        self.errors.append((test, log))
        self._mirrorOutput = True
        logging.error("%s - ERR:\n%s", test, log)

    @failfast
    def addFailure(self, test: TestCase, err):
        """Override parent addFailure to support logging."""
        log = '%s\n %s' % (
            self.log_message, self._exc_info_to_string(err, test))  # pytype: disable=attribute-error # noqa
        self.failures.append((test, log))
        self._mirrorOutput = True
        logging.error("%s - FAIL:\n%s", test, log)

    def addSuccess(self, test: TestCase):   # pytype: disable=signature-mismatch  # noqa
        """Store the log messages for every successful test."""
        self.successes.append((test, self.log_message))
        logging.debug("%s - PASS:\n%s", test, self.log_message)

    def addSkip(self, test: TestCase, reason: str):   # pytype: disable=signature-mismatch  # noqa
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
        self.start_time_sec = test_case.start_time_sec
        self.duration_sec = test_case.duration_sec
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
        self.duration_sec = test_result.duration_sec
        self.results = []
        # pytype: disable=wrong-arg-types
        for m in test_result.successes:
            self.results.append(MethodResult(m, 'PASS'))
        for m in test_result.failures:
            self.results.append(MethodResult(m, 'FAIL'))
        for m in test_result.errors:
            self.results.append(MethodResult(m, 'ERROR'))
        for m in test_result.skipped:
            self.results.append(MethodResult(m, 'SKIPPED'))
        # pytype: enable=wrong-arg-types


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
    start_time_sec = 0
    end_time_sec = 0
    tests_pass = 0
    tests_total = 0
    tests_fail = 0

    def __init__(self, ctx: context.TestContext, tgt: target.TestTarget):
        self.test_target = ctx.target.target
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
