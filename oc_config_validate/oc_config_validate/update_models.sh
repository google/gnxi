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

# Code taken and adapted from
# https://github.com/openconfig/gnmitest/blob/master/schemas/openconfig/update.sh,
# under Apache License, Version 2.0.

MODELS_FOLDER=$(dirname $0)/models
MODELS_PATH=${MODELS_FOLDER}/public/release/models

PYBINDPLUGIN=$(/usr/bin/env python -c 'import pyangbind; import os; print ("{}/plugin".format(os.path.dirname(pyangbind.__file__)))')

rm -fr ${MODELS_FOLDER}/*
touch ${MODELS_FOLDER}/__init__.py

# Note: Ensure the branch "v5.9.0" below is valid 
git clone --depth 1 https://github.com/openconfig/public.git --branch "v5.9.0" --config "advice.detachedHead=false" "$MODELS_FOLDER/public"

echo -n "__model_versions__ = \"\"\"" > ${MODELS_FOLDER}/__init__.py

# 1. Generate standard OpenConfig models
for m in $( ls -d ${MODELS_PATH}/*/ ); do
  MODELS=$(find $m -name "openconfig-*.yang")
  MODEL_NAME=$(basename $m | tr "-" "_")
  echo -e "Binding models in $m. Found the following:\n$MODELS"
  echo "Will create model: $MODEL_NAME"
  pyang --plugindir $PYBINDPLUGIN -f pybind \
    --path "$MODELS_PATH/" --output "$MODELS_FOLDER/${MODEL_NAME}.py" ${MODELS}
  pyang --plugindir $PYBINDPLUGIN -f name --name-print-revision \
    --path "$MODELS_PATH/" ${MODELS} >> ${MODELS_FOLDER}/__init__.py
done

# 2. Create custom wifi_arista that includes Arista vendor augments.
curl -o ${MODELS_PATH}/wifi/arista-wifi-augments.yang https://raw.githubusercontent.com/aristanetworks/yang/master/WIFI-16.1.0/release/openconfig/models/arista-wifi-augments.yang
# If Arista provides the file over drive/email, comment the line above and use
# the line below instead
# cp $(dirname $0)/arista-wifi-augments.yang ${MODELS_PATH}/wifi/arista-wifi-augments.yang
for m in $( ls -d ${MODELS_PATH}/wifi/ ); do
  MODELS=$(find $m -name "openconfig-*.yang" -o -name "arista-wifi-augments.yang")
  MODELS_ARISTA=$(find $m -name "arista-wifi-augments.yang")
  MODEL_NAME=$(basename $m | tr "-" "_")_arista
  echo -e "Binding models in $m. Found the following:\n$MODELS"
  echo "Will create model: $MODEL_NAME"
  pyang --plugindir $PYBINDPLUGIN -f pybind \
    --path "$MODELS_PATH/" --output "$MODELS_FOLDER/${MODEL_NAME}.py" ${MODELS}
done
pyang --plugindir $PYBINDPLUGIN -f name --name-print-revision \
  --path "$MODELS_PATH/" ${MODELS_ARISTA} >> ${MODELS_FOLDER}/__init__.py

# 3. Create custom wifi_aruba that includes Aruba vendor augments conditionally.
ARUBA_YANG_FILE="$(dirname $0)/aruba-wifi-augments.yang"

if [ -f "$ARUBA_YANG_FILE" ]; then
  echo "Aruba YANG file found. Generating Aruba bindings..."
  cp "$ARUBA_YANG_FILE" ${MODELS_PATH}/wifi/aruba-wifi-augments.yang
  for m in $( ls -d ${MODELS_PATH}/wifi/ ); do
    MODELS=$(find $m -name "openconfig-*.yang" -o -name "aruba-wifi-augments.yang")
    MODELS_ARUBA=$(find $m -name "aruba-wifi-augments.yang")
    MODEL_NAME=$(basename $m | tr "-" "_")_aruba
    echo -e "Binding models in $m. Found the following:\n$MODELS"
    echo "Will create model: $MODEL_NAME"
    pyang --plugindir $PYBINDPLUGIN -f pybind \
      --path "$MODELS_PATH/" --output "$MODELS_FOLDER/${MODEL_NAME}.py" ${MODELS}
  done
  pyang --plugindir $PYBINDPLUGIN -f name --name-print-revision \
    --path "$MODELS_PATH/" ${MODELS_ARUBA} >> ${MODELS_FOLDER}/__init__.py
else
  echo "Aruba YANG file not found at $ARUBA_YANG_FILE. Skipping Aruba bindings generation."
fi

echo -n "\"\"\"" >> ${MODELS_FOLDER}/__init__.py

rm -rf "$MODELS_FOLDER/public"
