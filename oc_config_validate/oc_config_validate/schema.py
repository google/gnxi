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
import re
from inspect import isclass
from typing import Any, List, Tuple, Union

from pyangbind.lib import yangtypes
from pyangbind.lib.base import PybindBase
from pyangbind.lib.serialise import pybindJSONDecoder

from oc_config_validate import models
from oc_config_validate.gnmi import gnmi_pb2  # type: ignore

_RE_GNMI_PATH_ELEM = re.compile(r'''^
(?P<name>[^[]+)   # gNMI path name
(\[(?P<key>\w\D+) # gNMI path key
=
(?P<value>.*)     # gNMI path value
\])?$
''', re.VERBOSE)

_RE_GNMI_XPATH = re.compile(r'''/(?=(?:[^\[\]]|\[[^\[\]]+\])*$)''')

_RE_GNMI_KEY = re.compile(r'\[([^]]*)\]')

type_translate_map = {int: 'int_val', float: 'float_val', bool: 'bool_val',
                      str: 'string_val', dict: 'json_ietf_val'}
value_translate_map = {'int_val': int, 'float_val': float, 'bool_val': bool,
                       'string_val': str, 'json_ietf_val': dict,
                       'json_val': dict, 'uint_val': int, 'leaflist_val': list}


class BaseError(Exception):
    """Base Exception raised by this module."""


class ModelClassError(BaseError):
    """Error parsing or loading an OC Model Class."""


class XpathError(BaseError):
    """Error parsing gNMI xpath."""


class GnmiError(BaseError):
    """Error using gNMI."""


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
        ModelClassError if unable to find the Python class or if the class is
          not derived from PybindBase.
        XpathError if the xpath is invalid.
        AttributeError if unable to find an xpath element in the OC class.
    """
    model_inst = _ocObjFromPath(model, xpath)
    if not issubclass(model_inst.__class__, PybindBase):
        raise ModelClassError(
            "%s:%s is not a valid container class" % (model, xpath))
    return model_inst


def ocLeafFromPath(model: str, xpath: str) -> Any:
    """Create an empty value of the model leaf pointed by the xpath.

    This method look for the model type in the oc_config_validate.models
      package

    Args:
        model: the OC model class name in the oc_config_validate.models
          package, as `module.class`.
        xpath: the xpath to the OC leaf to create.

    Returns:
        A value object of the leaf, like int, float, str, etc.

    Raises:
        ModelClassError if unable to find the Python class or if the instance
          is not a Leaf.
        XpathError if the xpath is invalid.
        AttributeError if unable to find an xpath element in the OC class.
    """
    model_inst = _ocObjFromPath(model, xpath)
    if issubclass(model_inst.__class__, PybindBase):
        raise ModelClassError(
            "%s:%s is not a valid leaf class" % (model, xpath))
    return model_inst


def _ocObjFromPath(model: str, xpath: str) -> Any:
    """Create an empty Object of the model for the path.

    This method look for the model class in the oc_config_validate.models
      package

    Args:
        model: the OC model class name in the oc_config_validate.models
        package, as `module.class`.
        xpath: the xpath to the OC container to create.

    Returns:
        An object of the class.

    Raises:
        ModelClassError if unable to find the Python class or if the class.
        XpathError if the xpath is invalid.
        AttributeError if unable to find an xpath element in the OC class.
    """

    parts = model.split('.')
    if len(parts) != 2:
        raise ModelClassError("%s is not module.class" % model)
    model_mod = importlib.import_module(
        "oc_config_validate.models." + parts[0])
    model_cls = getattr(model_mod, parts[1])
    if not isclass(model_cls):
        raise ModelClassError(
            "%s is not a class in oc_config_validate.models package" % model)

    gnmi_xpath = parsePath(xpath)

    model_inst = model_cls()
    for e in gnmi_xpath.elem:
        model_inst = getattr(model_inst, yangtypes.safe_name(e.name))
        if e.key:
            save_key = {}
            for k, v in e.key.items():
                save_key[yangtypes.safe_name(k)] = v
            model_inst = model_inst.add(**save_key)
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


def getOcModelsVersions() -> str:
    """Returns a multi-line string of the OC models versions used."""
    return models.__model_versions__.strip()


def ocLeafsPaths(model: str, xpath: str) -> List[str]:
    """Returns a list of xpaths to leafs, rooted at xpath.

    This method look for the model class in the oc_config_validate.models
      package

    Args:
        model: the OC model class name in the oc_config_validate.models
        package, as `module.class`.
        xpath: the xpath to the OC container to root at.

    Returns:
        A list of paths to leafs, rooted at the xpath.

    Raises:
        Error if unable to find the Python class or if the class is not derived
          from PybindBase.
        XpathError if the xpath is invalid.
        AttributeError if unable to find an xpath element in the OC class.

    """
    obj = ocContainerFromPath(model, xpath)

    paths = []
    for leaf, val in obj.get().items():
        if isinstance(val, dict):
            paths.extend(ocLeafsPaths(model, xpath + "/" + leaf))
        else:
            paths.append(xpath + "/" + leaf)
    return paths


def parsePath(xpath: str) -> gnmi_pb2.Path:
    """Parse an xpath string into gNMI.Path object.

    Args:
        xpath: String representation of a gNMI path.

    Returs:
        A gNMI.Path oject.

    Raises:
        XpathError: If unable to parse the xpath provided.
    """

    if not xpath or xpath == '/':
        return gnmi_pb2.Path(elem=[])
    xpath = xpath.strip('/').strip('/')  # Removes leading/trailing '/'.
    path_parts = _RE_GNMI_XPATH.split(xpath)
    gnmi_elems = []
    for part in path_parts:
        part_search = _RE_GNMI_PATH_ELEM.search(part)
        if not part_search:  # Invalid path specified.
            raise XpathError('xpath component parse error: %s' % part)
        if part_search.group('key') is not None:  # A path key was provided.
            key = {}
            for k in _RE_GNMI_KEY.findall(part):
                key[k.split('=')[0]] = k.split('=')[-1]
            gnmi_elems.append(
                gnmi_pb2.PathElem(name=part_search.group('name'), key=key))
        else:
            gnmi_elems.append(gnmi_pb2.PathElem(name=part, key={}))
    return gnmi_pb2.Path(elem=gnmi_elems)


def isPathOK(xpath: str) -> bool:
    """Return True is the xpath is parsed correctly into a gNMI Path

    Args:
        xpath: The xpath to parse
    """
    try:
        parsePath(xpath)
    except XpathError:
        return False
    return True


def isPathInRequestedPaths(xpath: str, xpaths: List[str]) -> bool:
    """Returns True if xpath in included or equal to any request path.

    The request paths can contain paths with wildcard '*'.

    Args:
      xpath: Non-wildcard xpath to check.
      xpaths: List of xpaths (can have wildcards) to contain xpath.
    """
    def _inOrEqual(x: str):
        # Ensure the path has keys sorted
        x = pathToString(parsePath(x))
        x_regex = re.escape(x).replace(r"\*", r"(.+)")
        return re.match(x_regex, xpath) is not None

    xpath = pathToString(parsePath(xpath))
    return any(map(_inOrEqual, xpaths))


def pathToString(path: gnmi_pb2.Path) -> str:
    """Parse a gNMI Path to a URI-like string.

    The keys are sorted alphabetically.

    Args:
        A gNMI.Path oject.

    Returs:
        xpath: String representation of a gNMI path.

    Raises:
        XpathError: If unable to parse the xpath provided.
    """
    path_elems = []
    for e in path.elem:
        elem = e.name
        if hasattr(e, "key"):
            keys = [f"[{k}={v}]" for k, v in e.key.items()]
            elem += f"{''.join(sorted(keys))}"
        path_elems.append(elem)
    return "/" + "/".join(path_elems)


def notificationsJsonString(notifications: List[gnmi_pb2.Notification],
                            indent: int = 2) -> str:
    """Returns a JSON-idented string of the Notification messages.

    Args:
        notifications: List of Notification messages to format.
        indent: number of JSON indentation spaces
    """
    notifs = []
    for n in notifications:
        notif = {"timestamp": n.timestamp, "updates": [None] * len(n.update)}
        for i, u in enumerate(n.update):
            notif["updates"][i] = {"path": pathToString(
                u.path), "value": typedValueToPython(u.val)}
        notifs.append(notif)
    return json.dumps(notifs, indent=2, default=str)


def pythonToTypedValue(value: Any) -> gnmi_pb2.TypedValue:
    """Return the gNMI TypedValue for a Python variable.

    Args:
        value: Value to set.

    Raises:
        GnmiError if type is unsupported.
    """

    typed_val = gnmi_pb2.TypedValue()
    try:
        value_type = type_translate_map[type(value)]
    except KeyError:
        raise GnmiError(
            f"Value type '{type(value)}' cannot be made a gNMI TypedValue")

    if isinstance(value, dict):
        setattr(typed_val, value_type, json.dumps(value).encode())
    else:
        setattr(typed_val, value_type, value)
    return typed_val


def typedValueToPython(value: gnmi_pb2.TypedValue) -> Union[
        bool, int, float, dict, str, list]:
    """Return the Python variable for a gNMI TypedValue.

    Args:
        value: gNMI TypedValue to convert.
    """
    value_type = value.WhichOneof("value")
    if value_type in ["json_ietf_val", "json_val"]:
        return json.loads(getattr(value, value_type))
    if value_type == "leaflist_val":
        return [typedValueToPython(v) for v in value.leaflist_val.element]
    else:
        tp = value_translate_map[value_type]
        return tp(getattr(value, value_type))
    raise GnmiError(f"Unsupported Type {value_type}")


def intersectCmp(want: dict, got: dict) -> Tuple[bool, str]:
    """Return true if all the keys of want are the same in got.

    E.g.: Taking want={"foo": 0, "bla": "test"} and
        got={"foo": 1, "bar": true, "bla": "test"}, the keys "foo"
        and "bla" are looked for and compared in got.

    The same comparison applies to nested dictionaries and lists.

    Args:
        want, got: Dictionaries to compare.

    Returns:
        A tuple with the comparison result and a diff text, if different.
    """
    for k in want.keys():
        x = want[k]
        y = got.get(k)
        if y is None:
            return False, "key %s not found" % k
        if isinstance(x, dict) and isinstance(y, dict):
            return intersectCmp(x, y)
        if isinstance(x, list) and isinstance(y, list):
            return intersectListCmp(x, y)
        if x != y:
            return False, "key %s: got '%s', wanted '%s'" % (k, y, x)
    return True, ""


def intersectListCmp(want: list, got: list) -> Tuple[bool, str]:
    """Return true if all the elements in want are the same in got.

    For dictionaries, intersectCmp() is used for comparison.
    Order is not considered.

    E.g.: Taking [0, "bla"] and [1, "bar", "bla"], the elements 0 and "bla"
        are looked for and compared in got.

    The same comparison applies to nested dictionaries and lists.

    Args:
        want, got: Lists to compare.

    Returns:
        A tuple with the comparison result and a diff text, if different.
    """
    for i in want:
        for j in got:
            matched = False
            if not isinstance(i, type(j)):
                continue
            if isinstance(i, dict):
                cmp, _ = intersectCmp(i, j)
            else:
                cmp = bool(i == j)
            if cmp:
                matched = True
                break
        if not matched:
            return False, "List element %s not matched" % i
    return True, ""


def gNMISubscriptionOnceRequest(
    xpaths: List[str],
    encoding: str = 'JSON_IETF'
) -> gnmi_pb2.SubscribeRequest:
    """
    Returns a gNMI Subscription ONCE requests for the xpaths.

    Args:
        xpaths: List of gNMI xpaths to subscribe to.
        encoding: Encoding requested for the Updates.
    """
    paths = [parsePath(xpath) for xpath in xpaths]
    return gnmi_pb2.SubscribeRequest(subscribe=gnmi_pb2.SubscriptionList(
        subscription=[gnmi_pb2.Subscription(path=path)
                      for path in paths],
        mode=gnmi_pb2.SubscriptionList.ONCE,
        encoding=encoding))


def gNMISubscriptionStreamSampleRequest(
    xpaths: List[str],
    sample_interval: int,
    encoding: str = 'JSON_IETF'
) -> gnmi_pb2.SubscribeRequest:
    """
    Returns a gNMI Subscription STREAM SAMPLE request for the xpaths.

    Args:
        xpaths: gNMI xpaths to subscribe to.
        sample_interval: Nanoseconds interval between Updates.
        encoding: Encoding requested for the Updates.
    """
    paths = [parsePath(xpath) for xpath in xpaths]
    return gnmi_pb2.SubscribeRequest(subscribe=gnmi_pb2.SubscriptionList(
        subscription=[gnmi_pb2.Subscription(
            path=path,
            mode=gnmi_pb2.SubscriptionMode.SAMPLE,
            sample_interval=sample_interval) for path in paths],
        mode=gnmi_pb2.SubscriptionList.STREAM,
        encoding=encoding))


def gNMISubscriptionStreamOnChangeRequest(
    xpaths: List[str],
    encoding: str = 'JSON_IETF'
) -> gnmi_pb2.SubscribeRequest:
    """
    Returns a gNMI Subscription STREAM ON_CHANGE request for the xpaths.

    Args:
        xpaths: gNMI xpaths to subscribe to.
        encoding: Encoding requested for the Updates.
    """
    paths = [parsePath(xpath) for xpath in xpaths]
    return gnmi_pb2.SubscribeRequest(subscribe=gnmi_pb2.SubscriptionList(
        subscription=[gnmi_pb2.Subscription(
            path=path,
            mode=gnmi_pb2.SubscriptionMode.ON_CHANGE) for path in paths],
        mode=gnmi_pb2.SubscriptionList.STREAM,
        encoding=encoding))
