#!/bin/bash

# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Code taken and adapted from 
# https://github.com/openconfig/gnmitest/blob/master/schemas/openconfig/update.sh,
# under Apache License, Version 2.0.

MODELS_FOLDER=$(dirname $0)/models

rm -rf "${MODELS_FOLDER}"

git clone --depth 1 https://github.com/openconfig/public.git "$MODELS_FOLDER/public"

PYBINDPLUGIN=$(/usr/bin/env python -c 'import pyangbind; import os; print ("{}/plugin".format(os.path.dirname(pyangbind.__file__)))')

MODELS=( "${MODELS_FOLDER}/public/release/models/interfaces/openconfig-interfaces.yang" \
          "${MODELS_FOLDER}/public/release/models/interfaces/openconfig-if-ip.yang" \
          "${MODELS_FOLDER}/public/release/models/wifi/openconfig-access-points.yang" \
          "${MODELS_FOLDER}/public/release/models/wifi/openconfig-ap-interfaces.yang" \
          "${MODELS_FOLDER}/public/release/models/wifi/openconfig-ap-manager.yang" \
          "${MODELS_FOLDER}/public/release/models/system/openconfig-system.yang" \
          "${MODELS_FOLDER}/public/release/models/system/openconfig-aaa.yang" \
          "${MODELS_FOLDER}/public/release/models/local-routing/openconfig-local-routing.yang" \
          "${MODELS_FOLDER}/public/release/models/bgp/openconfig-bgp.yang" \
          "${MODELS_FOLDER}/public/release/models/bgp/openconfig-bgp-global.yang" \
          "${MODELS_FOLDER}/public/release/models/bgp/openconfig-bgp-neighbor.yang" \
          "${MODELS_FOLDER}/public/release/models/lldp/openconfig-lldp.yang" )

pyang --plugindir $PYBINDPLUGIN -f pybind  --path "$MODELS_FOLDER/" \
  --split-class-dir "$MODELS_FOLDER/" ${MODELS[@]}

pyang --plugindir $PYBINDPLUGIN -f name --name-print-revision \
  --path "$MODELS_FOLDER/" --output "$MODELS_FOLDER/versions" ${MODELS[@]}

rm -rf "$MODELS_FOLDER/public"
