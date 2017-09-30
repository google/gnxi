/* Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gnmi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/value"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/google/gnxi/utils/target/gnmi/model_data"
	"github.com/google/gnxi/utils/target/gnmi/model_data/oc_struct"
)

var (
	// model is the model for test config target.
	model = &Model{
		modelData:       model_data.ModelData,
		structRootType:  reflect.TypeOf((*oc_struct.Device)(nil)),
		schemaTreeRoot:  oc_struct.SchemaTree["Device"],
		jsonUnmarshaler: oc_struct.Unmarshal,
	}
)

func TestCapabilities(t *testing.T) {
	s, err := NewTarget(model, nil, nil)
	if err != nil {
		t.Fatalf("error in creating target: %v", err)
	}
	resp, err := s.Capabilities(nil, &pb.CapabilityRequest{})
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}
	if !reflect.DeepEqual(resp.GetSupportedModels(), model.modelData) {
		t.Errorf("got supported models %v\nare not the same as\nmodel supported by the target %v", resp.GetSupportedModels(), model.modelData)
	}
	if !reflect.DeepEqual(resp.GetSupportedEncodings(), supportedEncodings) {
		t.Errorf("got supported encodings %v\nare not the same as\nencodings supported by the target %v", resp.GetSupportedEncodings(), supportedEncodings)
	}
}

func TestGet(t *testing.T) {
	jsonConfigRoot := `{
	  "components": {
	    "component": [
	      {
	        "config": {
	          "name": "gateway"
	        },
	        "name": "swpri1-1-1"
	      }
	    ]
	  }
	}`

	s, err := NewTarget(model, []byte(jsonConfigRoot), nil)
	if err != nil {
		t.Fatalf("error in creating target: %v", err)
	}

	tds := []struct {
		desc        string
		textPbPath  string
		wantRetCode codes.Code
		wantRespVal interface{}
	}{{
		desc:        "root node",
		wantRetCode: codes.OK,
		wantRespVal: jsonConfigRoot,
	}, {
		desc:        "root child node",
		textPbPath:  `elem: <name: "components" >`,
		wantRetCode: codes.OK,
		wantRespVal: `{
			"component": [{
				"config": {
		        	"name": "gateway"
				},
		        "name": "swpri1-1-1"
			}]}`,
	}, {
		desc: "node with attribute",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "name" value: "swpri1-1-1" >
				>`,
		wantRetCode: codes.OK,
		wantRespVal: `{
				"config": {"name": "gateway"},
				"name": "swpri1-1-1"
			}`,
	}, {
		desc: "node with attribute in its parent",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "name" value: "swpri1-1-1" >
				>
				elem: <name: "config" >`,
		wantRetCode: codes.OK,
		wantRespVal: `{"name": "gateway"}`,
	}, {
		desc: "ref leaf node",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "name" value: "swpri1-1-1" >
				>
				elem: <name: "name" >`,
		wantRetCode: codes.OK,
		wantRespVal: "swpri1-1-1",
	}, {
		desc: "regular leaf node",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "name" value: "swpri1-1-1" >
				>
				elem: <name: "config" >
				elem: <name: "name" >`,
		wantRetCode: codes.OK,
		wantRespVal: "gateway",
	}, {
		desc: "non-existing node: wrong path name",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "foo" value: "swpri1-1-1" >
				>
				elem: <name: "bar" >`,
		wantRetCode: codes.NotFound,
	}, {
		desc: "non-existing node: wrong path attribute",
		textPbPath: `
				elem: <name: "components" >
				elem: <
					name: "component"
					key: <key: "foo" value: "swpri2-2-2" >
				>
				elem: <name: "name" >`,
		wantRetCode: codes.NotFound,
	}}

	for _, td := range tds {
		t.Run(td.desc, func(t *testing.T) {
			runTestGet(t, s, td.textPbPath, td.wantRetCode, td.wantRespVal)
		})
	}
}

// runTestGet requests a path from the target by Get grpc call, and compares if the return code and response value are expected.
func runTestGet(t *testing.T, s *Target, textPbPath string, wantRetCode codes.Code, wantRespVal interface{}) {
	// Send request
	var pbPath pb.Path
	if err := proto.UnmarshalText(textPbPath, &pbPath); err != nil {
		t.Fatalf("error in unmarshaling path: %v", err)
	}
	req := &pb.GetRequest{
		Path:      []*pb.Path{&pbPath},
		Encoding:  pb.Encoding_JSON_IETF,
		UseModels: s.model.modelData,
	}
	resp, err := s.Get(nil, req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != wantRetCode {
		t.Fatalf("got return code %v, want %v", gotRetStatus.Code(), wantRetCode)
	}

	// Check response value
	var gotVal interface{}
	if resp != nil {
		notifs := resp.GetNotification()
		if len(notifs) != 1 {
			t.Fatalf("got %d notifications, want 1", len(notifs))
		}
		updates := notifs[0].GetUpdate()
		if len(updates) != 1 {
			t.Fatalf("got %d updates in the notification, want 1", len(updates))
		}
		val := updates[0].GetVal()
		if val.GetJsonIetfVal() == nil {
			gotVal, err = value.ToScalar(val)
			if err != nil {
				t.Errorf("got: %v, want a scalar value", gotVal)
			}
		} else {
			// Unmarshal json data to gotVal container for comparison
			if err := json.Unmarshal(val.GetJsonIetfVal(), &gotVal); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
			}
			var wantJSONStruct interface{}
			if err := json.Unmarshal([]byte(wantRespVal.(string)), &wantJSONStruct); err != nil {
				t.Fatalf("error in unmarshaling IETF JSON data to json container: %v", err)
			}
			wantRespVal = wantJSONStruct
		}
	}
	if !reflect.DeepEqual(gotVal, wantRespVal) {
		t.Errorf("got: %v,\nwant %v", gotVal, wantRespVal)
	}
}

func TestSet(t *testing.T) {
	interfacesValidConfig := `{
		"interfaces": {
			"interface": [
				{
					"config": {
						"name": "gateway"
					},
					"name": "eth1"
				}
			]
		}
	}`
	interfacesNonExistingField := `{
		"interfaces": {
			"interface": [
				{
					"config": {
						"name": "gateway"
					},
					"hostname": "eth1"
				}
			]
		}
	}`
	interfacesWrongDataTypeStruct4List := `{
		"interfaces": {
			"interface": {
				"config": {
					"name": "gateway"
				},
				"name": "eth1"
			}
		}
	}`
	interfacesWrongDataTypeString4Struct := `{
		"interfaces": {
			"interface": [
				{
					"config": "foo",
					"name": "eth1"
				}
			]
		}
	}`

	openflowValidConfig := `{
		"openconfig-openflow:system": {
			"openflow": {
				"agent": {
					"config": {
						"backoff-interval": 5,
						"datapath-id": "00:16:3e:00:00:00:00:00",
						"failure-mode": "SECURE",
						"inactivity-probe": 10,
						"max-backoff": 10
					}
				}
			}
		},
		"system": {
			"config": {
				"domain-name": "google.com",
				"hostname": "switch_a",
				"login-banner": "Hello!",
				"motd-banner": "Hi There!"
			}
		}
	}`
	openflowWrongDataFormatDpid := `{
		"openconfig-openflow:system": {
			"openflow": {
				"agent": {
					"config": {
						"backoff-interval": 5,
						"datapath-id": "123456",
						"failure-mode": "SECURE",
						"inactivity-probe": 10,
						"max-backoff": 10
					}
				}
			}
		},
	}`
	openflowWrongDataTypeFloat4Int := `{
		"openconfig-openflow:system": {
			"openflow": {
				"agent": {
					"config": {
						"backoff-interval": 5.0,
						"datapath-id": "00:16:3e:00:00:00:00:00",
						"failure-mode": "SECURE",
						"inactivity-probe": 10,
						"max-backoff": 10
					}
				}
			}
		},
	}`
	openflowWrongDataTypeString4Int := `{
		"openconfig-openflow:system": {
			"openflow": {
				"agent": {
					"config": {
						"backoff-interval": "5",
						"datapath-id": "00:16:3e:00:00:00:00:00",
						"failure-mode": "SECURE",
						"inactivity-probe": 10,
						"max-backoff": 10
					}
				}
			}
		},
	}`

	tds := []struct {
		desc        string
		config      string
		wantRetCode codes.Code
	}{{
		desc:        "interfaces valid config",
		config:      interfacesValidConfig,
		wantRetCode: codes.OK,
	}, {
		desc:        "interfaces config with non-existing field",
		config:      interfacesNonExistingField,
		wantRetCode: codes.InvalidArgument,
	}, {
		desc:        "interfaces config with wrong data type: struct for list",
		config:      interfacesWrongDataTypeStruct4List,
		wantRetCode: codes.InvalidArgument,
	}, {
		desc:        "interfaces config with wrong data type: string for struct",
		config:      interfacesWrongDataTypeString4Struct,
		wantRetCode: codes.InvalidArgument,
	}, {
		desc:        "openflow valid config",
		config:      openflowValidConfig,
		wantRetCode: codes.OK,
	}, {
		desc:        "openflow config with wrong data format: DPID",
		config:      openflowWrongDataFormatDpid,
		wantRetCode: codes.InvalidArgument,
	}, {
		desc:        "openflow config with wrong data type: float for int",
		config:      openflowWrongDataTypeFloat4Int,
		wantRetCode: codes.InvalidArgument,
	}, {
		desc:        "openflow config wrong data type: string for int",
		config:      openflowWrongDataTypeString4Int,
		wantRetCode: codes.InvalidArgument,
	}}

	for _, td := range tds {
		t.Run(td.desc, func(t *testing.T) {
			runTestSet(t, td.config, td.wantRetCode)
		})
	}
}

// runTestSet sets the json config to an empty target, then checks if the return code is expected.
func runTestSet(t *testing.T, config string, wantRetCode codes.Code) {
	// Create a new target with empty config
	s, err := NewTarget(model, nil, nil)
	if err != nil {
		t.Fatalf("error in creating config target: %v", err)
	}

	// Send request
	upd := &pb.Update{
		Path: pbRootPath,
		Val: &pb.TypedValue{
			Value: &pb.TypedValue_JsonIetfVal{
				JsonIetfVal: []byte(config),
			},
		},
	}
	req := &pb.SetRequest{Replace: []*pb.Update{upd}}
	_, err = s.Set(nil, req)

	// Check return code
	gotRetStatus, ok := status.FromError(err)
	if !ok {
		t.Fatal("got a non-grpc error from grpc call")
	}
	if gotRetStatus.Code() != wantRetCode {
		t.Fatalf("got return code %v, want %v", gotRetStatus.Code(), wantRetCode)
	}
}
