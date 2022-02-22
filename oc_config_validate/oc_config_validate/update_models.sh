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

rm -fr ${MODELS_FOLDER}
git clone --depth 1 https://github.com/openconfig/public.git "$MODELS_FOLDER/public"

# Skip openconfig-bdf model until https://github.com/robshakir/pyangbind/issues/286 is fixed
for model in $(ls -d ${MODELS_PATH}/* | grep -v bfd); do
  MODELS=$(find $model -name openconfig-*.yang)
  echo "Binding models in $model"
  # Skipping Warnings due to https://github.com/openconfig/public/issues/571
  pyang -W none --plugindir $PYBINDPLUGIN -f pybind \
    --path "$MODELS_PATH/" --split-class-dir "$MODELS_FOLDER/" ${MODELS}
  pyang -W none --plugindir $PYBINDPLUGIN -f name --name-print-revision \
  --path "$MODELS_PATH/" ${MODELS} >> ${MODELS_FOLDER}/versions
done

rm -rf "$MODELS_FOLDER/public"
