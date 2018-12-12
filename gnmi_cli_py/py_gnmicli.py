"""Copyright 2018 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

python3 CLI utility for interacting with Network Elements using gNMI.

This utility can be utilized as a reference, or standalone utility, for
interacting with Network Elements which support OpenConfig and gNMI.

Current supported gNMI features:
- GetRequest
- Target hostname override
- Auto-loads Target cert from Target if not specified
- User/password based authentication
- Certifificate based authentication

Current unsupported gNMI features:
- Subscribe
- SetRequest
- Multiple xpaths
"""

from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
import argparse
import logging
import os
import re
import ssl
import six
try:
  import gnmi_pb2  # pylint: disable=g-import-not-at-top
except ImportError:
  print('ERROR: Ensure you have grpcio-tools installed; eg\n'
        'sudo apt-get install -y pip\n'
        'sudo pip install --no-binary=protobuf -I grpcio-tools==1.15.0')
import gnmi_pb2_grpc  # pylint: disable=g-import-not-at-top

__version__ = '0.2'

_RE_PATH_COMPONENT = re.compile(r'''
^
(?P<pname>[^[]+)  # gNMI path name
(\[(?P<key>\w+)   # gNMI path key
=
(?P<value>.*)    # gNMI path value
\])?$
''', re.VERBOSE)


class Error(Exception):
  """Module-level Exception class."""


class XpathError(Error):
  """Error parsing xpath provided."""


def _create_parser():
  """Create parser for arguments passed into the program from the CLI.

  Returns:
    Argparse object.
  """
  parser = argparse.ArgumentParser(description='gNMI CLI utility.')
  parser = argparse.ArgumentParser(
      formatter_class=argparse.RawDescriptionHelpFormatter, epilog='\nExample'
      ' GetRequest without user/password and over-riding Target certificate CN:'
      '\npython py_gnmicli.py -t 127.0.0.1 -p 8080 -x \'/access-points/'
      'access-point[hostname=test-ap]/\' -rcert ~/certs/target-cert.crt -o '
      'openconfig.example.com')
  parser.add_argument('-t', '--target', type=str, help='The gNMI Target',
                      required=True)
  parser.add_argument('-p', '--port', type=str, help='The port the gNMI Target '
                      'is listening on', required=True)
  parser.add_argument('-u', '--username', type=str, help='Username to use'
                      'when establishing a gNMI Channel to the Target',
                      required=False)
  parser.add_argument('-w', '--password', type=str, help='Password to use'
                      'when establishing a gNMI Channel to the Target',
                      required=False)
  parser.add_argument('-m', '--mode', choices=['get', 'set', 'subscribe'],
                      help='Mode of operation when interacting with network'
                      ' element. Default=get', default='get')
  parser.add_argument('-pkey', '--private_key', type=str, help='Fully'
                      'quallified path to Private key to use when establishing'
                      'a gNMI Channel to the Target', required=False)
  parser.add_argument('-rcert', '--root_cert', type=str, help='Fully quallified'
                      'Path to Root CA to use when building the gNMI Channel',
                      required=False)
  parser.add_argument('-cchain', '--cert_chain', type=str, help='Fully'
                      'quallified path to Certificate chain to use when'
                      'establishing a gNMI Channel to the Target', default=None,
                      required=False)
  parser.add_argument('-x', '--xpath', type=str, help='The gNMI path utilized'
                      'in the GetRequest or Subscirbe', required=True)
  parser.add_argument('-o', '--host_override', type=str, help='Use this as '
                      'Targets hostname/peername when checking it\'s'
                      'certificate CN. You can check the cert with:\nopenssl'
                      'x509 -in certificate.crt -text -noout', required=False)
  parser.add_argument('-v', '--verbose', type=str, help='Print verbose messages'
                      'to the terminal', required=False)
  parser.add_argument('-d', '--debug', help='Enable gRPC debugging',
                      required=False, action='store_true')
  return parser


def _path_names(xpath):
  """Parses the xpath names.

  This takes an input string and converts it to a list of gNMI Path names. Those
  are later turned into a gNMI Path Class object for use in the Get/SetRequests.
  Args:
    xpath: (str) xpath formatted path.

  Returns:
    list of gNMI path names.
  """
  if not xpath or xpath == '/':  # A blank xpath was provided at CLI.
    return []
  return xpath.strip().strip('/').split('/')  # Remove leading and trailing '/'.


def _parse_path(p_names):
  """Parses a list of path names for path keys.

  Args:
    p_names: (list) of path elements, which may include keys.

  Returns:
    a gnmi_pb2.Path object representing gNMI path elements.

  Raises:
    XpathError: Unabled to parse the xpath provided.
  """
  gnmi_elems = []
  for word in p_names:
    word_search = _RE_PATH_COMPONENT.search(word)
    if not word_search:  # Invalid path specified.
      raise XpathError('xpath component parse error: %s' % word)
    if word_search.group('key') is not None:  # A path key was provided.
      gnmi_elems.append(gnmi_pb2.PathElem(name=word_search.group(
          'pname'), key={word_search.group('key'): word_search.group('value')}))
    else:
      gnmi_elems.append(gnmi_pb2.PathElem(name=word, key={}))
  return gnmi_pb2.Path(elem=gnmi_elems)


def _create_stub(creds, target, port, host_override):
  """Creates a gNMI GetRequest.

  Args:
    creds: (object) of gNMI Credentials class used to build the secure channel.
    target: (str) gNMI Target.
    port: (str) gNMI Target IP port.
    host_override: (str) Hostname being overridden for Cert check.

  Returns:
    a gnmi_pb2_grpc object representing a gNMI Stub.
  """
  if host_override:
    channel = gnmi_pb2_grpc.grpc.secure_channel(target + ':' + port, creds, ((
        'grpc.ssl_target_name_override', host_override,),))
  else:
    channel = gnmi_pb2_grpc.grpc.secure_channel(target + ':' + port, creds)
  return gnmi_pb2_grpc.gNMIStub(channel)


def _get(stub, paths, username, password):
  """Create a gNMI GetRequest.

  Args:
    stub: (class) gNMI Stub used to build the secure channel.
    paths: gNMI Path
    username: (str) Username used when building the channel.
    password: (str) Password used when building the channel.

  Returns:
    a gnmi_pb2.GetResponse object representing a gNMI GetResponse.
  """
  if username:  # User/pass supplied for Authentication.
    return stub.Get(
        gnmi_pb2.GetRequest(path=[paths], encoding='JSON_IETF'),
        metadata=[('username', username), ('password', password)])
  return stub.Get(gnmi_pb2.GetRequest(path=[paths], encoding='JSON_IETF'))


def _build_creds(target, port, root_cert, cert_chain, private_key):
  """Define credentials used in gNMI Requests.

  Args:
    target: (str) gNMI Target.
    port: (str) gNMI Target IP port.
    root_cert: (str) Root certificate file to use in the gRPC secure channel.
    cert_chain: (str) Certificate chain file to use in the gRPC secure channel.
    private_key: (str) Private key file to use in the gRPC secure channel.

  Returns:
    a gRPC.ssl_channel_credentials object.
  """
  if not root_cert:
    logging.warning('No certificate supplied, obtaining from Target')
    root_cert = ssl.get_server_certificate((target, port)).encode('utf-8')
    return gnmi_pb2_grpc.grpc.ssl_channel_credentials(
        root_certificates=root_cert, private_key=None, certificate_chain=None)
  elif not cert_chain:
    logging.info('Only user/pass in use for Authentication')
    return gnmi_pb2_grpc.grpc.ssl_channel_credentials(
        root_certificates=six.moves.builtins.open(root_cert, 'rb').read(),
        private_key=None, certificate_chain=None)
  return gnmi_pb2_grpc.grpc.ssl_channel_credentials(
      root_certificates=six.moves.builtins.open(root_cert, 'rb').read(),
      private_key=six.moves.builtins.open(private_key, 'rb').read(),
      certificate_chain=six.moves.builtins.open(cert_chain, 'rb').read())


def main():
  argparser = _create_parser()
  args = vars(argparser.parse_args())
  mode = args['mode']
  target = args['target']
  port = args['port']
  root_cert = args['root_cert']
  cert_chain = args['cert_chain']
  private_key = args['private_key']
  xpath = args['xpath']
  host_override = args['host_override']
  user = args['username']
  password = args['password']
  if mode in ('set', 'subscribe'):
    print('Mode %s not available in this version' % mode)
    return
  if args['debug']:
    os.environ['GRPC_TRACE'] = 'all'
    os.environ['GRPC_VERBOSITY'] = 'DEBUG'
  paths = _parse_path(_path_names(xpath))
  creds = _build_creds(target, port, root_cert, cert_chain, private_key)
  stub = _create_stub(creds, target, port, host_override)
  if mode == 'get':
    print('Performing GetRequest, encoding=JSON_IETF', ' to ', target,
          ' with the following gNMI Path\n', '-'*25, '\n', paths)
    response = _get(stub, paths, user, password)
    print('The GetResponse is below\n' + '-'*25 + '\n', response)


if __name__ == '__main__':
  main()
