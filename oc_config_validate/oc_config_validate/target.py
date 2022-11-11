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
import ssl
import time
from typing import Any, Dict, Iterator, List, Optional, Tuple

import grpc
from grpc._channel import _InactiveRpcError, _MultiThreadedRendezvous

from oc_config_validate import context, schema
from oc_config_validate.gnmi import gnmi_pb2, gnmi_pb2_grpc  # type: ignore


class BaseError(Exception):
    """Base Error for the target class"""


class TargetError(BaseError):
    """Error connecting to Target."""


class RpcError(BaseError):
    """Error on the RPC."""


class TestTarget():
    """gNMI Target to run tests against.

    Args:
        hostname_port: Target to connect to, as 'hostname:port'.
    """

    def __init__(self, tgt: context.Target):
        self.address = tgt.target
        self.username = tgt.username
        self.password = tgt.password
        self.notls = tgt.no_tls
        self.host_tls_override = tgt.tls_host_override
        self.target_cert_as_root_ca = tgt.target_cert_as_root_ca
        self.gnmi_set_cooldown_secs = tgt.gnmi_set_cooldown_secs
        self.key = None
        self.cert = None
        self.root_ca = None
        self.connection = None
        self.stub = None

        if not self.notls:
            self.setCertificates(key=tgt.private_key, cert=tgt.cert_chain,
                                 root_ca=tgt.root_ca_cert)

    def __repr__(self):
        return ('TestTarget(address=%r, username=%r, notls=%r)' %
                (self.address, self.username, self.notls))

    def __str__(self):
        return self.address

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
        if self.target_cert_as_root_ca:
            logging.info('Obtaining Root CA certificate from Target')
            address_elems = self.address.split(':')
            try:
                self.root_ca = ssl.get_server_certificate(
                    (address_elems[0], int(address_elems[1]))).encode('utf-8')
            except ConnectionRefusedError as err:
                raise TargetError(
                    "Unable to get Root CA from Target") from err

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
            RpcError if unable to connect to Target.
        """
        path = schema.parsePath(xpath)
        self._gNMIConnnect()
        metadata = self._buildGnmiStubMetadata()
        try:
            return self.stub.Get(
                gnmi_pb2.GetRequest(path=[path], encoding='JSON_IETF'),
                **metadata)
        except _InactiveRpcError as err:
            raise RpcError(err) from err

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
            TypeError if the set_type is not valid.
            RpcError if unable to connect to Target.
        """
        if set_type.lower() not in ['delete', 'update', 'replace']:
            raise TypeError('%s is not a valid gNMI Set type' % set_type)

        path = schema.parsePath(xpath)
        path_val = None
        if value:
            val = schema.pythonToTypedValue(value)
            path_val = gnmi_pb2.Update(path=path, val=val)
        self._gNMIConnnect()
        metadata = self._buildGnmiStubMetadata()

        response = None
        try:
            if set_type.lower() == 'delete':
                response = self.stub.Set(gnmi_pb2.SetRequest(delete=[path]),
                                         **metadata)
            if set_type.lower() == 'update':
                response = self.stub.Set(gnmi_pb2.SetRequest(
                    update=[path_val]), **metadata)
            if set_type.lower() == 'replace':
                response = self.stub.Set(gnmi_pb2.SetRequest(
                    replace=[path_val]), **metadata)
        except _InactiveRpcError as err:
            raise RpcError(err) from err

        if self.gnmi_set_cooldown_secs:
            time.sleep(self.gnmi_set_cooldown_secs)

        return response

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

    def gNMISubsOnce(self, xpaths: List[str]) -> List[gnmi_pb2.Notification]:
        """Subscribes using ONCE mode, returns the Notifications received.

        Args:
            xpaths: List of gNMI paths to subscribe to.

        Returns:
            A list of gnmi_pb2.Notification objects received.
            None on error.

        Raises:
            RpcError if unable to connect to Target.
            ValueError if the SubscribeResponse is invalid.
        """
        def _subscribe_requet_generator(
            requests: List[gnmi_pb2.SubscribeRequest]
        ) -> Iterator[gnmi_pb2.SubscribeRequest]:
            for r in requests:
                yield r

        paths = [schema.parsePath(xpath) for xpath in xpaths]
        self._gNMIConnnect()
        metadata = self._buildGnmiStubMetadata()
        notifications = []
        stream_requests = _subscribe_requet_generator(
            [gnmi_pb2.SubscribeRequest(subscribe=gnmi_pb2.SubscriptionList(
                subscription=[
                    gnmi_pb2.Subscription(path=path)
                    for path in paths],
                mode=gnmi_pb2.SubscriptionList.ONCE,
                encoding='JSON_IETF'))
             ])
        try:
            for resp in self.stub.Subscribe(
                    stream_requests,
                    **metadata):
                if resp.sync_response:
                    pass
                elif resp.update:
                    notifications.append(resp.update)
                else:
                    raise ValueError("Invalid SubscribeResponse %s" % resp)
            return notifications
        except grpc._channel._InactiveRpcError as err:
            raise RpcError(err) from err

    def validate(self):
        """Ensures the Target is defined appropriately.

        Raises:
          ValueError when the Target is not defined correctly.
        """
        parts = self.address.split(":")
        if len(parts) != 2 or not bool(parts[0]) or not parts[1].isdigit():
            raise ValueError("Needed valid target HOSTNAME:PORT")

        # If using client certificates for TLS, provide key and cert
        if not self.notls and (bool(self.key) ^ bool(self.cert)):
            raise ValueError("TLS key and cert are both needed.")
