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

from oc_config_validate import testbase

SUPPORTED_FORMATS = ["json"]


class FormatterError(Exception):
    """Error while working with Formatter"""


class Formatter():
    """Interface for concrete results formatters."""

    def writeResultsToFile(self, test_run: testbase.TestRun, filename: str):
        """Templates writing to a file."""
        raise FormatterError("writeResultsToFile() not implemented.")


class JSONFormatter(Formatter):
    """Writes the test results to a JSON file"""

    def writeResultsToFile(self, test_run: testbase.TestRun, filename: str):
        """Write the results of a test run to a JSON file.

        This  method assumes the filename is valid and writable.

        """
        with open(filename, 'w+', encoding='utf-8') as file:
            json.dump(test_run, file, indent=2, default=lambda o: o.__dict__)


def makeFormatter(format_name: str) -> Formatter:
    """Return a Formatter for the format, if supported.

    Args:
        format_name: a valid format from SUPPORTED_FORMATS.

    Raises:
        FormatterError if the format is not supported.
    """
    if format_name.lower() == "json":
        return JSONFormatter()
    raise FormatterError("%s is not a supported format." % format_name)
