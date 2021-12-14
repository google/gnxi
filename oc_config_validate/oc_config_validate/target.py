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

---

Code taken and adapted from

  https://github.com/google/gnxi/tree/master/gnmi_cli_py, under
  Apache License, Version 2.0 license
"""


import json
import logging
import re
import ssl
from typing import Any, Dict, List, Optional, Tuple, Union

from oc_config_validate.gnmi import gnmi_pb2, gnmi_pb2_grpc

_RE_GNMI_PATH_ELEM = re.compile(r'''^
(?P<name>[^[]+)   # gNMI path name
(\[(?P<key>\w\D+) # gNMI path key
=
(?P<value>.*)     # gNMI path value
\])?$
''', re.VERBOSE)

_RE_GNMI_XPATH = re.compile(r'''/(?=(?:[^\[\]]|\[[^\[\]]+\])*$)''')

_RE_GNMI_KEY = re.compile(r'\[([^]]*)\]')


class BaseError(Exception):
    """Base Error for the target class"""


class XpathError(BaseError):
    """Error parsing gNMI xpath."""


class TargetError(BaseError):
    """Error connecting to Target."""


class GnmiError(BaseError):
    """Error using gNMI."""


class TestTarget():
    """gNMI Target to run tests against.

    Args:
        hostname_port: Target to connect to, as 'hostname:port'.
    """

    def __init__(self, hostname_port: str):
        self.address = hostname_port
        self.notls = False
        self.username = None
        self.password = None
        self.host_tls_override = None
        self.key = None
        self.cert = None
        self.root_ca = None
        self.connection = None
        self.stub = None

    def __repr__(self):
        return ('TestTarget(address=%r, username=%r, notls=%r)' %
                (self.address, self.username, self.notls))

    def __str__(self):
        return self.address

    def setCredentials(self, username: str, password: str):
        """Set the username and password to use on connections.

        Args:
            username: The username.
            password: The password value, passed verbatim to the target.
        """
        self.username = username
        self.password = password

    def setCertificates(self, key: Optional[str] = None,
                        cert: Optional[str] = None,
                        root_ca: Optional[str] = None,):
        """Read the TLS certificates to use on connections.

        Args:
            key: Path to the private key file.
                 If None, no key will be used.
            cert: Path to the certificate to present to the target.
                  If None, no certificate will be used.
            root_ca: Path to the trusted CA certificate. If None, it will
                     be requested from the Target.
        """
        def _fileReadOrNone(filename: str) -> Optional[bytes]:
            try:
                with open(filename, 'rb') as file:
                    return file.read()
            except IOError as io_error:
                logging.error("Unable to read %s: %s", filename, io_error)
                return None
        if key:
            self.key = _fileReadOrNone(key)
        if cert:
            self.cert = _fileReadOrNone(cert)
        if root_ca:
            self.root_ca = _fileReadOrNone(root_ca)

    def _gNMIConnnect(self):
        """Create a gNMI connection."""
        if self.stub:
            return
        if self.notls:
            channel = gnmi_pb2_grpc.grpc.insecure_channel(self.address)
        else:
            creds = self._buildCredentials()
            logging.debug("creds: %s", creds)
            if self.host_tls_override:
                channel = gnmi_pb2_grpc.grpc.secure_channel(
                    self.address, creds,
                    (('grpc.ssl_target_name_override',
                      self.host_tls_override,),))
            else:
                channel = gnmi_pb2_grpc.grpc.secure_channel(
                    self.address, creds)
        self.stub = gnmi_pb2_grpc.gNMIStub(channel)

    def _buildCredentials(self) -> gnmi_pb2_grpc.grpc.ssl_channel_credentials:
        """Build credentials used in gNMI Requests.

        Returns:
            A gRPC credentials object.

        Raises:
            TargetError if unable to build the credentials.
        """
        if not self.root_ca:
            logging.info('Obtaining Root CA certificate from Target')
            address_elems = self.address.split(':')
            try:
                self.root_ca = ssl.get_server_certificate(
                    (address_elems[0], int(address_elems[1]))).encode('utf-8')
            except ConnectionRefusedError as err:
                raise TargetError(
                    "Unable to get Root CA from Target: %s") from err

        return gnmi_pb2_grpc.grpc.ssl_channel_credentials(
            root_certificates=self.root_ca,
            private_key=self.key,
            certificate_chain=self.cert)

    def _buildGnmiStubMetadata(self) -> Dict[str, List[Tuple[str, str]]]:
        """Build the gNMI metadata used for authentication."""
        metadata = {}
        if self.username:  # User/pass supplied for Authentication.
            metadata = {'metadata': [
                ('username', self.username), ('password', self.password)]}
        return metadata

    def gNMIGet(self, xpath: str) -> gnmi_pb2.GetResponse:
        """Create a gNMI GetRequest.

        Args:
            xpath: gNMI Path

        Returns:
            A gnmi_pb2.GetResponse object.

        Raises:
            grpc._channel._InactiveRpcError  if unable to connect to Target.
        """
        path = parsePath(xpath)
        self._gNMIConnnect()
        metadata = self._buildGnmiStubMetadata()
        return self.stub.Get(
            gnmi_pb2.GetRequest(path=[path], encoding='JSON_IETF'),
            **metadata)

    def _gNMISet(self, xpath: str, set_type: str,
                 value: Optional[Any] = None) -> gnmi_pb2.SetResponse:
        """Create a gNMI SetRequest.

        Args:
            xpath: The gNMI path to set.
            set_type: UPDATE | DELETE | REPLACE
            value: Value to set; can be numeric, string or JSON-IETF.

        Returns:
            A gnmi_pb2.SetResponse object.

        Raises:
            GnmiError if the set_type is not valid.
            grpc._channel._InactiveRpcError if unable to connect to Target.
        """
        if set_type.lower() not in ['delete', 'update', 'replace']:
            raise GnmiError('%s is not a valid gNMI Set type' % set_type)

        path = parsePath(xpath)

        path_val = None
        if value:
            val = pythonToTypedValue(value)
            path_val = gnmi_pb2.Update(path=path, val=val)
        self._gNMIConnnect()
        metadata = self._buildGnmiStubMetadata()

        if set_type.lower() == 'delete':
            return self.stub.Set(gnmi_pb2.SetRequest(delete=[path]),
                                 **metadata)
        if set_type.lower() == 'update':
            return self.stub.Set(gnmi_pb2.SetRequest(update=[path_val]),
                                 **metadata)
        if set_type.lower() == 'replace':
            return self.stub.Set(gnmi_pb2.SetRequest(replace=[path_val]),
                                 **metadata)

    def gNMISetUpdate(self, xpath: str,
                      value: Any) -> gnmi_pb2.SetResponse:
        """Create a gNMI SetRequest Update.

        Args:
            xpath: The gNMI path to replace.
            value: Value to set; can be numeric, string or JSON-IETF.

        Returns:
            A gnmi_pb2.SetResponse object.
        """
        return self._gNMISet(xpath, 'update', value)

    def gNMISetReplace(self, xpath: str,
                       value: Any) -> gnmi_pb2.SetResponse:
        """Create a gNMI SetRequest Replace.

        Args:
            xpath: The gNMI path to replace.
            value: Value to set; can be numeric, string or JSON-IETF.

        Returns:
            A gnmi_pb2.SetResponse object.
        """
        return self._gNMISet(xpath, 'replace', value)

    def gNMISetDelete(self, xpath: str) -> gnmi_pb2.SetResponse:
        """Create a gNMI SetRequest Delete.

        Args:
            xpath: The gNMI path to delete.

        Returns:
            A gnmi_pb2.SetResponse object.
        """
        return self._gNMISet(xpath, 'delete')

    def gNMISetConfigFile(self, file_path: str, xpath: str):
        """Apply the JSON-formated configuration to the target.

        Args:
            file_path: Path to a JSON file with the configuration to apply.
            xpath: gNMI xpath where to apply the configuration.

        Raises:
            IOError json.JSONDecodeError if unable to read and parse the file.
            TargetError, GnmiError or XpathError if unable to set the config.
        """
        with open(file_path, encoding="utf8") as file:
            json_data = json.load(file)
        self.gNMISetUpdate(xpath, json_data)


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
            raise XpathError('xpath component parse error: %s' % part_search)
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


def pythonToTypedValue(value: Any) -> gnmi_pb2.TypedValue:
    """Return the gNMI TypedValue for a Python variable.

    Args:
        value: Value to set; can be int, float, bool, str or dict.

    Raises:
        GnmiError if type is unsupported.
    """
    translate_map = {int: 'int_val', float: 'float_val', bool: 'bool_val',
                     str: 'string_val', dict: 'json_ietf_val'}
    val = gnmi_pb2.TypedValue()
    if type(value) not in translate_map:
        raise GnmiError("Value '%s' cannot be converted to gNMI TypedValue" %
                        value)
    if isinstance(value, dict):
        setattr(val, translate_map[type(value)], json.dumps(value).encode())
    else:
        setattr(val, translate_map[type(value)], value)
    return val


def typedValueToPython(value: gnmi_pb2.TypedValue, tp: type) -> Union[
        bool, int, float, dict, str]:
    """Return the Python variable for a gNMI TypedValue.

    Args:
        value: gNMI TypedValue to convert.
        tp: Type to extract, from [bool, int, float, dict, str]

    Raises:
        GnmiError if type is unsupported.
    """
    if tp == dict:
        return json.loads(value.json_ietf_val)
    if tp == bool:
        return bool(value.bool_val)
    if tp == int:
        return int(value.int_val)
    if tp == float:
        return float(value.float_val)
    if tp == str:
        return str(value.string_val)
    raise GnmiError("Unsupported TypedValue")


def intersectCmp(a: dict, b: dict) -> Tuple[bool, str]:
    """Return true if the common keys have the same values.

    E.g.: Taking {"foo": 0, "bla": "test"} and
        {"foo": 1, "bar": true, "bla": "test"}, the intersected keys are
        "foo" and "bla", and the values are not equal.

    If there are no common keys, returns False.

    The same comparison applies to nested dictionaries

    Args:
        a, b: Dictionaries to intersect and compare.

    Returns:
        A tuple with the comparison result and a diff text, if different.
    """
    common_keys = a.keys() & b.keys()
    if not common_keys:
        return False, "got %s, wanted %s" % (a.keys(), b.keys())
    for k in common_keys:
        x = a[k]
        y = b[k]
        if isinstance(x, dict) and isinstance(y, dict):
            cmp, diff = intersectCmp(x, y)
            if not cmp:
                return False, diff
        else:
            if x != y:
                return False, "key %s: %s != %s" % (k, x, y)
    return True, ""
