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
    labels = []
    tests = []
    description = ""

    def __init__(self, description, labels, tests):
        self.description = description
        self.labels = labels
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


def fromFile(yaml_stream):
    """Create a TestContext object from a YAML file.

    Args:
        yaml_stream: A file-like object that supports the .read() method.

    """
    context = yaml.safe_load(yaml_stream)
    return context
