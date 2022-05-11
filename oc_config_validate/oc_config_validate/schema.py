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

import importlib
import json
import os
import re
from inspect import isclass
from typing import List, Union

from pyangbind.lib import yangtypes
from pyangbind.lib.base import PybindBase
from pyangbind.lib.serialise import pybindJSONDecoder

from oc_config_validate import models, target


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


def ocContainerFromPath(model: str, xpath: str) -> PybindBase:
    """Create an empty PybindBase instance of the model for the path.

    This method look for the model class in the oc_config_validate.models
      package

    Args:
        model: the OC model class name in the oc_config_validate.models
        package, as `module.class`.
        xpath: the xpath to the OC container to create.

    Returns:
        An PybindBase object of the class.

    Raises:
        Error if unable to find the Python class or if the class is not derived
          from PybindBase.
        target.XpathError if the xpath is invalid.
        AttributeError if unable to find an xpath element in the OC class.
    """

    parts = model.split('.')
    if len(parts) != 2:
        raise Error("%s is not module.class" % model)
    model_mod = importlib.import_module(
        "oc_config_validate.models." + parts[0])
    model_cls = getattr(model_mod, parts[1])
    if not isclass(model_cls):
        raise Error(
            "%s is not a class in oc_config_validate.models package" % model)

    gnmi_xpath = target.parsePath(xpath)

    model_inst = model_cls()
    for e in gnmi_xpath.elem:
        model_inst = getattr(model_inst, yangtypes.safe_name(e.name))
        if e.key:
            save_key = {}
            for k, v in e.key.items():
                save_key[yangtypes.safe_name(k)] = v
            model_inst = model_inst.add(**save_key)
    if not issubclass(model_inst.__class__, PybindBase):
        raise Error("%s:%s is not a valid container class" % (model, xpath))
    return model_inst


def fixSubifIndex(json_value: dict):
    """Rewrite the index of a pybindJSON-produced subinterface as int.

    pybindJSON dumps the index as a str value, instead of int.

    https://github.com/robshakir/pyangbind/issues/139

    """
    index = json_value['openconfig-interfaces:subinterfaces'][
        'subinterface'][0]['index']
    json_value['openconfig-interfaces:subinterfaces']['subinterface'][0][
        'index'] = int(index)


def getOcModelsVersions() -> List[str]:
    """Returns a list of the OC models versions used.

     Returns an empty list if unable to read the models/versions file.
     """
    versions_file = os.path.join(models.__path__[0], "versions")
    if os.path.isfile(versions_file):
        with open(versions_file) as f:
            return [line.strip() for line in f]
    return []
