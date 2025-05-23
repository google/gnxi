# -*- coding: utf-8 -*-
from operator import attrgetter
from pyangbind.lib.yangtypes import RestrictedPrecisionDecimalType
from pyangbind.lib.yangtypes import RestrictedClassType
from pyangbind.lib.yangtypes import TypedListType
from pyangbind.lib.yangtypes import YANGBool
from pyangbind.lib.yangtypes import YANGListType
from pyangbind.lib.yangtypes import YANGDynClass
from pyangbind.lib.yangtypes import ReferenceType
from pyangbind.lib.yangtypes import YANGBinary
from pyangbind.lib.yangtypes import YANGBitsType
from pyangbind.lib.base import PybindBase
from collections import OrderedDict
from decimal import Decimal

import builtins as __builtin__
long = int
class openconfig_gnsi_authz(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi-authz - based on the path /openconfig-gnsi-authz. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module provides a data model for the metadata of the gRPC
authorization policies installed on a networking device.
  """
  _pyangbind_elements = {}

  

class openconfig_gnsi_certz(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi-certz - based on the path /openconfig-gnsi-certz. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module provides a data model for the metadata of gRPC credentials
installed on a networking device.
  """
  _pyangbind_elements = {}

  

class openconfig_gnsi_acctz(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi-acctz - based on the path /openconfig-gnsi-acctz. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module provides counters of gNSI accountZ requests and responses and
the quantity of data transferred.
  """
  _pyangbind_elements = {}

  

class openconfig_gnsi_pathz(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi-pathz - based on the path /openconfig-gnsi-pathz. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module provides a data model for the metadata of
OpenConfig-path-based authorization policies installed on a networking
device.
  """
  _pyangbind_elements = {}

  

class openconfig_gnsi(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi - based on the path /openconfig-gnsi. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module defines a set of extensions that provide gNSI (the gRPC
Network Security Interface) specific extensions to the OpenConfig data models.
Specifically, the parameters for the configuration of the service, and
configuration and state are added.

The gNSI protobufs and documentation are published at
https://github.com/openconfig/gnsi.
  """
  _pyangbind_elements = {}

  

class openconfig_gnsi_credentialz(PybindBase):
  """
  This class was auto-generated by the PythonClass plugin for PYANG
  from YANG module openconfig-gnsi-credentialz - based on the path /openconfig-gnsi-credentialz. Each member element of
  the container is represented as a class variable - with a specific
  YANG type.

  YANG Description: This module provides a data model for the metadata of SSH and console
credentials installed on a networking device.

The following leaves MUST be treated as invalid when the gNSI server is
enabled and credentialz is supported by the implementation:
 /system/aaa/authentication/users/user/config/ssh-key
 /system/aaa/authentication/users/user/state/ssh-key
 /system/aaa/authentication/users/user/config/password
 /system/aaa/authentication/users/user/state/password
 /system/aaa/authentication/users/user/config/password-hashed
 /system/aaa/authentication/users/user/state/password-hashed
  """
  _pyangbind_elements = {}

  

