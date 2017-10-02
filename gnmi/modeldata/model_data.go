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

// Package modeldata contains the following model data in gnmi proto struct:
//	openconfig-interfaces 2.0.0,
//	openconfig-openflow 0.1.0,
//	openconfig-platform 0.5.0,
//	openconfig-system 0.2.0.
package modeldata

import (
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

const (
	// OpenconfigInterfacesModel is the openconfig YANG model for interfaces.
	OpenconfigInterfacesModel = "openconfig-interfaces"
	// OpenconfigOpenflowModel is the openconfig YANG model for openflow.
	OpenconfigOpenflowModel = "openconfig-openflow"
	// OpenconfigPlatformModel is the openconfig YANG model for platform.
	OpenconfigPlatformModel = "openconfig-platform"
	// OpenconfigSystemModel is the openconfig YANG model for system.
	OpenconfigSystemModel = "openconfig-system"
)

var (
	// ModelData is a list of supported models.
	ModelData = []*pb.ModelData{{
		Name:         OpenconfigInterfacesModel,
		Organization: "OpenConfig working group",
		Version:      "2.0.0",
	}, {
		Name:         OpenconfigOpenflowModel,
		Organization: "OpenConfig working group",
		Version:      "0.1.0",
	}, {
		Name:         OpenconfigPlatformModel,
		Organization: "OpenConfig working group",
		Version:      "0.5.0",
	}, {
		Name:         OpenconfigSystemModel,
		Organization: "OpenConfig working group",
		Version:      "0.2.0",
	}}
)
