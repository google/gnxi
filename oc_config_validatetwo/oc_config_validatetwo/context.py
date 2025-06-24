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

import yaml


class TestContext(yaml.YAMLObject):
    """Object parsed from the TestContext YAML file."""

    yaml_loader = yaml.SafeLoader
    yaml_tag = u'!TestContext'
    init_configs = []
    labels = []
    target = None
    tests = []
    description = ""

    def __init__(self, description, init_configs, labels, target, tests):
        self.description = description
        self.init_configs = init_configs
        self.labels = labels
        self.target = target
        self.tests = tests

    def __repr__(self):
        return ('TestContext(description=%r, labels=%r, tests=%r)' % (
                self.description, self.labels, self.tests))


class TestCase(yaml.YAMLObject):
    """Object parsed from the TestContext YAML file."""

    yaml_loader = yaml.SafeLoader
    yaml_tag = u'!TestCase'

    def __init__(self, name, class_name, args):
        self.name = name
        self.class_name = class_name
        self.args = args

    def __repr__(self):
        return ('TestCase(name=%r, class_name=%r, args=%r)' %
                (self.name, self.class_name, self.args))


class InitConfig(yaml.YAMLObject):
    """Object parsed from the TestContext YAML file."""

    yaml_loader = yaml.SafeLoader
    yaml_tag = u'!InitConfig'

    def __init__(self, filename, xpath):
        self.filename = filename
        self.xpath = xpath

    def __repr__(self):
        return ('InitConfig(filename=%r, xpath=%r)' %
                (self.filename, self.xpath))


class Target(yaml.YAMLObject):
    """Object parsed from the TestContext YAML file."""

    yaml_loader = yaml.SafeLoader
    yaml_tag = u'!Target'

    target = ""
    username = ""
    password = ""
    private_key = ""
    root_ca_cert = ""
    cert_chain = ""
    no_tls = False
    tls_host_override = ""
    target_cert_as_root_ca = False
    gnmi_set_cooldown_secs = 10

    def __repr__(self):
        return 'Target(target=%r, no_tls=%r)' % (self.target, self.no_tls)

    def validate(self):
        """Ensures the Target is defined appropriately.

        Raises:
            ValueError when the Target is not defined correctly.
        """
        parts = self.target.split(":")
        if len(parts) != 2 or not bool(parts[0]) or not parts[1].isdigit():
            raise ValueError("Needed valid target HOSTNAME:PORT")

        # If using client certificates for TLS, provide key and cert
        if not self.no_tls and (
                bool(self.private_key) ^ bool(self.cert_chain)):
            raise ValueError("TLS key and cert are both needed.")


def fromFile(file_path) -> TestContext:
    """Create a TestContext object from a YAML file.

    Args:
        file_path: Path to a YAML file with test profile.

    Raises:
        IOError: An error occurred while trying to read the file.
        YAMLError: An error occurred while trying to parse the YAML
           file.
    """
    with open(file_path, encoding="utf8") as raw_profile_data:
        ctx = yaml.safe_load(raw_profile_data)

    # If no Target is defined in the tests file, create a default one.
    if not ctx.target:
        ctx.target = Target()

    return ctx
