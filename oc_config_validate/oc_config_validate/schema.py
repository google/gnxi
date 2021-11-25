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
import re
from inspect import isclass
from typing import Union

from pyangbind.lib.base import PybindBase
from pyangbind.lib.serialise import pybindJSONDecoder

from oc_config_validate import models


class Error(Exception):
    """Base Exception raised by this module."""


def decodeJson(json_text: Union[str, bytes], obj: PybindBase):
    """Decode a JSON text into a PybindBase object.

    This method is used to validate the JSON text adheres to the OC schema.

    Args:
        json_text: The JSON-IETF text to decode.
        obj: The PybindBase object to decode the texto into.

    Raises:
        Error if unable to parse the JSON text.
    """
    json_text = removeOpenConfigPrefix(json_text)
    pybindJSONDecoder.load_ietf_json(json.loads(json_text), None, None, obj)


def removeOpenConfigPrefix(json_text: Union[str, bytes]) -> Union[str, bytes]:
    """Remove open-config prefixed so JSON text can be processed by PyBind.

    When JSON-IETF is used for a gNMI response, type references will
    prepend the corresponding model name(ie. openconfig-aaa:RADIUS) when
    referencing a server of type RADIUS in a leaf.  These must be removed
    before being processed by PyBind.

    Args:
        json_text: The JSON-IETF text to correct.

    Returns:
        The string that can be fed directly to pybindJSONDecoder.
    """
    # https://regex101.com/r/xiZj4Q/1
    if isinstance(json_text, bytes):
        return re.sub(b'(openconfig(-[a-z]+)+\:)', b'',
                      json_text)  # noqa
    return re.sub(r'(openconfig(-[a-z]+)+\:)', '', json_text)


def containerFromName(class_name: str) -> PybindBase:
    """Create an empty PybindBase instance of the model class.

    This method interprets the class_name as part of the
        oc_config_validate.models package

    Args:
        class_name: a string with the class name.

    Returns:
        An PybindBase object of the class.

    Raises:
        AttributeError is unable to find the Python class.
        Error if the class is not derived from PybindBase.
    """
    cls = models
    for part in class_name.split('.'):
        cls = getattr(cls, part)
    if not isclass(cls):
        raise Error("%s is not a class" % class_name)
    if not issubclass(cls, PybindBase):
        raise Error("%s is not derived from PybindBase" % class_name)
    return cls()


def fixSubifIndex(json_value: dict):
    """Rewrite the index of a pybindJSON-produced subinterface as int.

    pybindJSON dumps the index as a str value, instead of int.

    https://github.com/robshakir/pyangbind/issues/139

    """
    index = json_value['openconfig-interfaces:subinterfaces'][
        'subinterface'][0]['index']
    json_value['openconfig-interfaces:subinterfaces']['subinterface'][0][
        'index'] = int(index)
