#!/bin/bash

GNMI_FOLDER=$(dirname $0)

rm -f -R "${GNMI_FOLDER}/proto"

mkdir -p "${GNMI_FOLDER}/proto/gnmi_ext" "${GNMI_FOLDER}/proto/gnmi"

wget -q  https://raw.githubusercontent.com/openconfig/gnmi/master/proto/gnmi_ext/gnmi_ext.proto -P "${GNMI_FOLDER}/proto/gnmi_ext"
wget -q  https://raw.githubusercontent.com/openconfig/gnmi/master/proto/gnmi/gnmi.proto         -P "${GNMI_FOLDER}/proto/gnmi"

sed -i 's/import "github.com\/openconfig\/gnmi\/proto\/gnmi_ext\/gnmi_ext.proto";/import "gnmi_ext.proto";/' "${GNMI_FOLDER}/proto/gnmi/gnmi.proto"

python3 -m grpc.tools.protoc --proto_path "${GNMI_FOLDER}/proto/gnmi_ext/"                                           --python_out "${GNMI_FOLDER}/" --grpc_python_out "${GNMI_FOLDER}/" "${GNMI_FOLDER}/proto/gnmi_ext/gnmi_ext.proto"
python3 -m grpc.tools.protoc --proto_path "${GNMI_FOLDER}/proto/gnmi_ext/" --proto_path "${GNMI_FOLDER}/proto/gnmi/" --python_out "${GNMI_FOLDER}/" --grpc_python_out "${GNMI_FOLDER}/" "${GNMI_FOLDER}/proto/gnmi/gnmi.proto"

sed -i 's/import gnmi_ext_pb2/from oc_config_validate.gnmi import gnmi_ext_pb2/' "${GNMI_FOLDER}/gnmi_pb2.py"
sed -i 's/import gnmi_pb2 /from oc_config_validate.gnmi import gnmi_pb2 /' "${GNMI_FOLDER}/gnmi_pb2_grpc.py"

rm -f -R "${GNMI_FOLDER}/proto"
