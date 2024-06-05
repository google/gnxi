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
import logging
import os
import sys
import time
from typing import Any, Dict

import yaml

from oc_config_validate import (__version__, context, formatter, runner,
                                schema, target, testbase)

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
        help="The gNMI Target, as hostname:port.",
    )
    parser.add_argument(
        "-user",
        "--username",
        type=str,
        help="Username to use when establishing a gNMI Channel to the Target.",
    )
    parser.add_argument(
        "-pass",
        "--password",
        type=str,
        help="Password to use when establishing a gNMI Channel to the Target.",
    )
    parser.add_argument(
        "-key",
        "--private_key",
        type=str,
        help="Path to the Private key to use when establishing"
        "a gNMI Channel to the Target.",
    )
    parser.add_argument(
        "-ca",
        "--root_ca_cert",
        type=str,
        help="Path to Root CA to use when building the gNMI Channel.",
    )
    parser.add_argument(
        "-cert",
        "--cert_chain",
        type=str,
        help="Path to Certificate chain to use when"
        "establishing a gNMI Channel to the Target.")
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
        "--init_set_replace",
        action="store_true",
        help="Use gNMI SetReplace method for the initial configs.")
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
        help="Format of the test results file. Default=JSON.",
        choices=formatter.SUPPORTED_FORMATS,
        default="json")
    parser.add_argument(
        "-v",
        "--version",
        help="Print program version",
        action="store_true")
    parser.add_argument(
        "-V",
        "--verbose",
        help="Enable gRPC debugging and extra logging.",
        action="store_true")
    parser.add_argument(
        "-models",
        "--oc_models_versions",
        help="Print OC models versions.",
        action="store_true")
    parser.add_argument(
        "--no_tls",
        help="gRPC insecure mode.",
        action="store_true")
    parser.add_argument(
        "-o",
        "--tls_host_override",
        type=str,
        action="store",
        help="Hostname to use during the TLS certificate check.",
    )
    parser.add_argument(
        "--target_cert_as_root_ca",
        action="store_true",
        help="Fetch the Target TLS cert and use it as client Root CA cert.",
    )
    parser.add_argument(
        "-set_cooldown",
        "--gnmi_set_cooldown_secs",
        type=int,
        action="store",
        help="Seconds to wait after a successful gNMI Set message.",
    )
    parser.add_argument(
        "--stop_on_error",
        action="store_true",
        help="Stop the execution if a test fails.",
    )
    parser.add_argument(
        "--log_gnmi",
        action="store_true",
        help="Log the gnmi requests to the tests results.",
    )
    return parser


def validateArgs(args: Dict[str, Any]):
    """Returns True if the arguments are valid.

     Raises:
       ValueError if any argument is invalid.
       IOError is unable to open a file given in argument.

    """

    def isFileOK(filename: str, writable: bool = False):
        try:
            file = open(filename, "w+" if writable else "r", encoding="utf8")
            file.close()
        except IOError as io_error:
            logging.error("Unable to open %s: %s", filename, io_error)
            raise

    # Mandatory args for tests
    for arg, write in [("tests_file", False), ("results_file", True)]:
        if not args[arg]:
            raise ValueError("Needed --%s file" % arg)
        isFileOK(args[arg], write)

    if args["init_config_file"]:
        isFileOK(args["init_config_file"], False)

    # Output format supported
    if (args["format"] and
            args["format"].lower() not in formatter.SUPPORTED_FORMATS):
        raise ValueError("Output format %s is not supported.")


def main():  # noqa
    """Executes this library."""
    argparser = createArgsParser()
    args = vars(argparser.parse_args())

    if args["version"]:
        print(__version__)
        sys.exit()
    if args["oc_models_versions"]:
        print(schema.getOcModelsVersions())
        sys.exit()

    if args["verbose"]:
        # os.environ["GRPC_TRACE"] = "all"
        os.environ["GRPC_VERBOSITY"] = "DEBUG"
    logging.basicConfig(
        level=logging.DEBUG if args["verbose"] else logging.INFO,
        format=LOGGING_FORMAT)

    try:
        validateArgs(args)
    except (IOError, ValueError) as error:
        sys.exit("Invalid arguments: %s" % error)

    if args["log_gnmi"]:
        testbase.LOG_GNMI = args["log_gnmi"]

    try:
        ctx = context.fromFile(args["tests_file"])
    except IOError as io_error:
        sys.exit("Unable to read %s: %s" % (args["tests_file"], io_error))
    except yaml.YAMLError as yaml_error:
        sys.exit("Unable to parse YAML file %s: %s" % (args["tests_file"],
                 yaml_error))

    logging.info("Read tests file '%s': %d tests to run",
                 args["tests_file"], len(ctx.tests))

    # Override Target options
    for arg in ["target", "username", "password", "no_tls", "private_key",
                "cert_chain", "root_ca_cert", "tls_host_override",
                "target_cert_as_root_ca", "gnmi_set_cooldown_secs"]:
        if args[arg]:
            setattr(ctx.target, arg, args[arg])
    try:
        ctx.target.validate()
    except ValueError as error:
        sys.exit("Invalid Target: %s" % error)

    # Append initial configuration paths and file to Context
    if args["init_config_file"]:
        ctx.init_configs.append(context.InitConfig(args["init_config_file"],
                                                   args["init_config_xpath"]))

    with target.TestTarget(ctx.target) as tgt:
        try:
            runner.setInitConfigs(ctx, tgt,
                                  stop_on_error=args["stop_on_error"],
                                  set_replace=args["init_set_replace"])
        except runner.InitConfigError as err:
            sys.exit("Unable to apply init config(s): %s" % err)

        start_t = time.time()
        results = runner.runTests(
            ctx, tgt, stop_on_error=args["stop_on_error"])
        end_t = time.time()

    test_run = testbase.TestRun(ctx, tgt)
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
