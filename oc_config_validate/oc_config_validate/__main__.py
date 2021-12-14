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
import argparse
import json
import logging
import os
import sys
import time
from typing import Any, Dict, Optional

import yaml

from oc_config_validate import context, formatter, runner, target, testbase

__version__ = "1.0.0"

LOGGING_FORMAT = "%(levelname)s(%(filename)s:%(lineno)d):%(message)s"


def createArgsParser() -> argparse.ArgumentParser:
    """Create parser for arguments passed into the program from the CLI.

     Returns:
        argparse.ArgumentParser object.

     """
    parser = argparse.ArgumentParser(
        description="OpenConfig Configuration Validation utility.")
    parser.add_argument(
        "-tgt",
        "--target",
        type=str,
        help="The gNMI Target, as hostname:port",
    )
    parser.add_argument(
        "-user",
        "--username",
        type=str,
        help="Username to use when establishing a gNMI Channel to the Target",
    )
    parser.add_argument(
        "-pass",
        "--password",
        type=str,
        help="Password to use when establishing a gNMI Channel to the Target",
    )
    parser.add_argument(
        "-key",
        "--private_key",
        type=str,
        help="Path to the Private key to use when establishing"
        "a gNMI Channel to the Target",
    )
    parser.add_argument(
        "-ca",
        "--root_ca_cert",
        type=str,
        help="Path to Root CA to use when building the gNMI Channel",
    )
    parser.add_argument(
        "-cert",
        "--cert_chain",
        type=str,
        help="Path to Certificate chain to use when"
        "establishing a gNMI Channel to the Target")
    parser.add_argument(
        "-tests",
        "--tests_file",
        type=str,
        action="store",
        help="YAML file to read the test to run.")
    parser.add_argument(
        "-init",
        "--init_config_file",
        type=str,
        action="store",
        help="JSON file with the initial full OpenConfig configuration to "
        "apply.")
    parser.add_argument(
        "-xpath",
        "--init_config_xpath",
        type=str,
        action="store",
        help="gNMI xpath where to apply the initial config.",
        default="/")
    parser.add_argument(
        "-results",
        "--results_file",
        type=str,
        action="store",
        help="Filename where to write the test results.")
    parser.add_argument(
        "-f",
        "--format",
        type=str,
        action="store",
        help="Format "
        "of the GetResponse to be printed. Default=JSON.",
        choices=["json", "protobuff"],
        default="json")
    parser.add_argument(
        "-v", "--version", help="Print program version", action="store_true")
    parser.add_argument(
        "-V", "--verbose", help="Enable gRPC debugging and extra logging",
        action="store_true")
    parser.add_argument(
        "--no_tls", help="gRPC insecure mode", action="store_true")
    parser.add_argument(
        "-o",
        "--tls_host_override",
        type=str,
        action="store",
        help="Hostname to use during the TLS certificate check",
    )
    parser.add_argument(
        "--stop_on_error",
        action="store_true",
        help="Stop the execution if a test fails",
    )
    parser.add_argument(
        "--log_gnmi",
        action="store_true",
        help="Log the gnmi requests to the tests results",
    )
    return parser


def validateArgs(args: Dict[str, Any]) -> bool:  # noqa
    """Returns True if the arguments are valid.

    To run tests:
     * A valid target is needed, as HOSTNAME:PORT.
     * All readable tests, results and initial config files are needed.
     * --notls and TLS arguments are exclusive.

    """

    def isTarget(tgt: str) -> bool:
        parts = tgt.split(":")
        return len(parts) == 2 and bool(parts[0]) and parts[1].isdigit()

    def isFileOK(filename: str, writable: bool = False) -> bool:
        try:
            file = open(filename, "w+" if writable else "r", encoding="utf8")
            file.close()
        except IOError as io_error:
            logging.error("Unable to open %s: %s", filename, io_error)
            return False
        return True

    # Mandatory args for tests
    if not args["target"] or not isTarget(args["target"]):
        logging.error("Needed valid --target HOSTNAME:PORT argument")
        return False

    for arg, write in [("tests_file", False), ("results_file", True)]:
        if not args[arg]:
            logging.error("Needed --%s file", arg)
            return False
        if not isFileOK(args[arg], write):
            return False

    if args["init_config_file"] and not isFileOK(args["init_config_file"],
                                                 False):
        return False

    # Either TLS or not
    if any([args[arg] for arg in [
            "private_key", "cert_chain", "root_ca_cert",
            "tls_host_override"]]) and args["no_tls"]:
        logging.error(
            "TLS arguments and --notls option are mutually exclusive.")
        return False

    # If using client certificates for TLS, provide key and cert
    if bool(args["private_key"]) ^ bool(args["cert_chain"]):
        logging.error(
            "TLS arguments -private_key and -cert_chain are both needed.")
        return False

    # Output format supported
    if (args["format"] and
            args["format"].lower() not in formatter.SUPPORTED_FORMATS):
        logging.error("Output format %s is not supported.")
        return False

    return True


def loadTestContextFromFile(file_path) -> Optional[context.TestContext]:
    """Loads the tests description file and returns a TestContext.

     Args:
        file_path: Path to a YAML file with test profile.

     Returns:
        A TestContext object with all test information, or None if
           error.

     Raises:
        IOError: An error occurred while trying to read the file.
        YAMLError: An error occurred while trying to parse the YAML
           file.
     """
    ctx = None
    try:
        with open(file_path, encoding="utf8") as raw_profile_data:
            ctx = context.fromFile(raw_profile_data)
    except IOError as io_error:
        logging.error("Unable to read %s: %s", file_path, io_error)
        return None
    except yaml.YAMLError as yaml_error:
        logging.error("Unable to parse YAML file %s: %s", file_path,
                      yaml_error)
        return None
    return ctx


def setConfigFile(file_path: str, xpath: str,
                  tgt: target.TestTarget) -> bool:
    """Applies the JSON-formated configuration to the target.

    Args:
        file_path: Path to a JSON file with the configuration to apply.
        xpath: gNMI xpath where to apply the configuration.
        tgt: Target to configure.

    Returns:
        True if the configuration was applied successfully.
    """
    try:
        tgt.gNMISetConfigFile(file_path, xpath)
    except (IOError, json.JSONDecodeError, target.BaseError) as err:
        logging.error("Unable to set configuration '%s' via gNMI: %s",
                      file_path, err)
        return False
    return True


def main():  # noqa
    """Executes this library."""
    argparser = createArgsParser()
    args = vars(argparser.parse_args())
    if args["version"]:
        print(__version__)
        sys.exit()
    if args["verbose"]:
        # os.environ["GRPC_TRACE"] = "all"
        os.environ["GRPC_VERBOSITY"] = "DEBUG"

    logging.basicConfig(
        level=logging.DEBUG if args["verbose"] else logging.INFO,
        format=LOGGING_FORMAT)

    if not validateArgs(args):
        sys.exit(1)

    ctx = loadTestContextFromFile(args["tests_file"])
    if ctx is None:
        sys.exit(1)
    logging.info("Read tests file '%s': %d tests to run",
                 args["tests_file"], len(ctx.tests))

    tgt = target.TestTarget(args["target"])
    tgt.setCredentials(args["username"], args["password"])
    if args["no_tls"]:
        tgt.notls = True
    else:
        logging.info("Using TLS for gNMI connection")
        tgt.setCertificates(key=args["private_key"],
                            cert=args["cert_chain"],
                            root_ca=args["root_ca_cert"])
        if args["tls_host_override"]:
            tgt.host_tls_override = args["tls_host_override"]

    if args["log_gnmi"]:
        testbase.LOG_GNMI = args["log_gnmi"]

    logging.info("Testing gNMI Target %s.", args["target"])

    # Apply initial configuration
    if args["init_config_file"]:
        if not setConfigFile(args["init_config_file"],
                             args["init_config_xpath"],
                             tgt):
            logging.error("Unable to set the initial configuration")
            sys.exit(1)
        logging.info("Initial OpenConfig '%s' applied on gNMI xpath %s",
                     args["init_config_file"], args["init_config_xpath"])

    start_t = time.time()
    results = runner.runTests(ctx, tgt, stop_on_error=args["stop_on_error"])
    end_t = time.time()

    test_run = testbase.TestRun(args["target"], ctx)
    test_run.copyResults(results, start_t, end_t)
    logging.info("Results Summary: %s", test_run.summary())

    try:
        fmtr = formatter.makeFormatter(args["format"])
        fmtr.writeResultsToFile(test_run, args["results_file"])
        logging.info("Test results written to %s", args["results_file"])
    except IOError as io_error:
        logging.exception("Unable to write file %s: %s", args["results_file"],
                          io_error)
    except TypeError as type_error:
        logging.exception("Unable to parse results into a JSON text: %s",
                          type_error)


if __name__ == "__main__":
    main()
