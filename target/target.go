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

// Package target provides helper functions for target gNMI binaries.
package target

import "github.com/openconfig/gnmi/proto/gnmi"

// ReflectGetRequest generates a gNMI GetResponse out of a gnmi GetRequest.
func ReflectGetRequest(request *gnmi.GetRequest) *gnmi.GetResponse {
	response := gnmi.GetResponse{Notification: []*gnmi.Notification{}}
	notification := gnmi.Notification{Update: []*gnmi.Update{}}
	for _, path := range request.Path {
		typedValue := gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "TEST STRING"}}
		update := gnmi.Update{Path: path, Val: &typedValue}
		notification.Update = append(notification.Update, &update)
	}
	response.Notification = append(response.Notification, &notification)
	return &response
}
