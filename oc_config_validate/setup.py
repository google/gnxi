"""A setuptools-based setup module.

From: https://github.com/pypa/sampleproject

"""

from setuptools import setup, find_packages
import pathlib

here = pathlib.Path(__file__).parent.resolve()

# Get the long description from the README file
long_description = (here / 'README.md').read_text(encoding='utf-8')

setup(
  name = "oc_config_validate",
  version = "1.0.0",
  description = "Validate OpenConfig-based configuration of devices.",
  long_description = long_description,
  long_description_content_type='text/markdown',
  url='https://github.com/google/gnxi',

  classifiers = [
    "Programming Language :: Python :: 3"
    "License :: OSI Approved :: Apache Software License"
    "Operating System :: OS Independent"
  ],

  package_dir = { "" : "oc_config_validate" },
  packages= find_packages( where="oc_config_validate" ),

  python_requires = ">=3.6",
  install_requires = [ "grpcio", "grpcio-tools", "pyang", "pyangbind", "PyYAML", "retry" ]

)
