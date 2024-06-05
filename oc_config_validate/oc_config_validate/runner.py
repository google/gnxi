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

import io
import json
import logging
import threading
import time
import unittest
from typing import List, Optional

from oc_config_validate import context, schema, target, testbase, testcases


class BaseError(Exception):
    """Base Error for the runner class"""


class InitConfigError(BaseError):
    """Error applying initial configs."""


class TestClassError(BaseError):
    """Error rendering a testclass."""


class TestThread(threading.Thread):
    """Threaded class that runs tests."""

    def __init__(self, test_suite: testbase.TestSuite,
                 connection_lock: threading.Lock,
                 stop_on_error: bool = False):
        """Initialize the threaded class.

        Args:
            test_suite: A TestSuite to run in this Thread.
            node_locks: A mutex to control the use of the connection.
            stop_on_error: True if the run stops after the first error.
        """
        super().__init__()
        self.test_suite = test_suite
        self.conn_lock = connection_lock
        self.results = testbase.TestResult()
        self.stream = io.StringIO()
        self.duration_sec = 0
        self.stop_on_error = stop_on_error

    def run(self):
        """Run the test suite."""
        text_test_runner = getRunner(self.stream)
        text_test_runner.failfast = self.stop_on_error
        if self.conn_lock:
            self.conn_lock.acquire()
        start_time = int(time.time())
        self.results = text_test_runner.run(self.test_suite)
        self.duration_sec = int(time.time() - start_time)
        if self.conn_lock:
            self.conn_lock.release()
        self.results.test_class_name = self.test_suite.test_class_name
        self.results.test_name = self.test_suite.test_name
        self.results.duration_sec = self.duration_sec


def getRunner(stream: io.StringIO) -> unittest.TextTestRunner:
    """Generate a unittest text runner that uses the the StringIO stream.

    Args:
        stream: A io.StringIO object.

    Returns:
        A unittest.TextTestRunner object.
    """
    return unittest.TextTestRunner(
        stream=stream,
        buffer=False,
        verbosity=2,
        resultclass=testbase.TestResult)


def getTestSuites(ctx: context.TestContext,
                  tgt: target.TestTarget):
    suites = []
    for t in ctx.tests:
        s = makeTestSuite(t.class_name, t.name)
        if s is not None:
            s.insertTarget(tgt)
            if hasattr(t, 'args') and t.args:
                s.insertArgs(t.args)
            suites.append(s)
    return suites


def makeTestSuite(test_class_name: str,
                  test_name: str) -> Optional[testbase.TestSuite]:
    """Build a test suite of the provided test class.

    Args:
        test_class_name: The class name.
        test_name: The name of the test.

    Returns:
        A unittest.TestSuite object,,
        None if the test_class_name was not found.
    """

    def strCmp(a, b):
        """Comparer method to get alphabetically-ordered strings."""
        return (a > b) - (a < b)

    try:
        test_class = _getClassFromName(test_class_name)
    except TestClassError as err:
        logging.error(err)
        return None
    suite = unittest.makeSuite(test_class,  # pytype: disable=wrong-arg-types
                               sortUsing=strCmp,
                               prefix='test',
                               suiteClass=testbase.TestSuite)
    suite.test_class_name = test_class_name
    suite.test_name = test_name
    return suite    # pytype: disable=bad-return-type


def _getClassFromName(class_name: str) -> testbase.TestCase:
    """Find a class type from a string.

    This method interprets the class_name as module.class

    Args:
        class_name: a string with the class name

    Returns:
        A class type derived from test_testbase.TestCase.

    Raises:
        TestClassError if not found or not derived from test_testbase.TestCase.

    """
    try:
        cls = testcases
        for part in class_name.split('.'):
            cls = getattr(cls, part)
    except AttributeError as err:
        raise TestClassError(
            f"Unable to find test class {class_name}") from err
    if not issubclass(cls, testbase.TestCase):
        raise TestClassError(
            f"Test class {class_name} not derived from test_testbase.TestCase")
    return cls


def runTests(ctx: context.TestContext,
             tgt: target.TestTarget,
             stop_on_error: bool = False) -> List[testbase.TestResult]:
    """Run all the tests in the TestContext against the TestTarget.

    Args:
        ctx: A context.TestContext object.
        tgt: A target.TestTarget with a connection established.
        stop_on_error: If True, execution stop if there is an error.

    Returns:
        A list of testbase.TestResult objects.

    """
    suites = getTestSuites(ctx, tgt)
    errored_test = len(ctx.tests) - len(suites)
    if errored_test:
        logging.error("Running %d/%d Tests, %d failed to load.",
                      len(suites), len(ctx.tests), errored_test)
    if stop_on_error:
        logging.info("Stopping if a Test fails.")
    results = []
    conn_lock = threading.Lock()
    for i, suite in enumerate(suites):
        thread = TestThread(suite, conn_lock, stop_on_error)
        logging.debug(f"Running Test {i+1} '{suite.test_name}'")
        thread.start()
        thread.join()
        results.append(thread.results)
        logging.info(
            f"Test {i+1} '{suite.test_name}' took "
            f"{thread.results.duration_sec} secs: "
            f"{'' if thread.results.wasSuccessful() else 'NOT '}"
            "PASSED\n")
        if stop_on_error and not thread.results.wasSuccessful():
            break
    return results


def setInitConfigs(ctx: context.TestContext,
                   tgt: target.TestTarget,
                   stop_on_error: bool = False,
                   set_replace: bool = False) -> bool:
    """Applies the initial configurations to the target.

    Args:
        ctx: A context.TestContext object.
        tgt: A target.TestTarget with a connection established.
        stop_on_error: If True, stop if there is an error.
        set_replace: If True, gNMI SetReplace is used. Else, SetUpdate is used.

    Raises:
        InitConfigError if failed to apply the config. Does not raise if
        stop_on_error.

    """

    if set_replace:
        logging.info("Using gNMI Set Replace method for each Initial Config.")

    for init_config in ctx.init_configs:
        if (not getattr(init_config, "filename", None) or
                not getattr(init_config, "xpath", None)):
            msg = "Initial configuration file and xpath are both needed."
            if stop_on_error:
                raise InitConfigError(msg)
            logging.error(msg)
        try:
            schema.parsePath(init_config.xpath)
            tgt.gNMISetConfigFile(init_config.filename,
                                  init_config.xpath,
                                  set_replace)
            logging.info(f"Initial OpenConfig '{init_config.filename}' "
                         f"applied at {init_config.xpath}")
        except (IOError, json.JSONDecodeError, target.BaseError) as err:
            msg = (f"Unable to set configuration '{init_config.filename}' "
                   f"at '{init_config.xpath}': {err}")
            if stop_on_error:
                raise InitConfigError(msg) from err
            logging.error(msg)
