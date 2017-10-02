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

// Package server implements a gnmi server to config a YANG modeled device.
package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/experimental/ygotutils"
	"github.com/openconfig/ygot/ygot"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	cpb "google.golang.org/genproto/googleapis/rpc/code"
)

// ConfigCallback is the signature of the function to apply a validated config
// to the physical device.
type ConfigCallback func(ygot.ValidatedGoStruct) error

var (
	pbRootPath         = &pb.Path{}
	supportedEncodings = []pb.Encoding{pb.Encoding_JSON, pb.Encoding_JSON_IETF}
)

// Server struct maintains the data structure for device config and implements
// the interface of gnmi server. It supports Capabilities, Get, and Set APIs.
// Typical usage:
//	g := grpc.NewServer()
//	s, err := server.NewServer(model, config, callback)
//	pb.RegisterGNMIServer(g, s)
//	reflection.Register(g)
//	listen, err := net.Listen("tcp", ":8080")
//	g.Serve(listen)
//
// For a real device, apply the config changes to the hardware in the callback
// function:
// func callback(config ygot.ValidatedGoStruct) error {
//		// Apply the config to your device and return nil if success. If fail,
//		// rollback all changes and return error.
//		//
//		// Do something ...
// }
type Server struct {
	model    *Model
	callback ConfigCallback

	config ygot.ValidatedGoStruct
	mu     sync.RWMutex // mu is the RW lock to protect the access to config
}

// NewServer creates an instance of Server with given json config.
func NewServer(model *Model, config []byte, callback ConfigCallback) (*Server, error) {
	rootStruct, err := model.NewConfigStruct(config)
	if err != nil {
		return nil, err
	}
	return &Server{
		model:    model,
		config:   rootStruct,
		callback: callback,
	}, nil
}

// checkEncodingAndModel checks whether encoding and models are supported by the
// server. Return error if anything is unsupported.
func (s *Server) checkEncodingAndModel(encoding pb.Encoding, models []*pb.ModelData) error {
	hasSupportedEncoding := false
	for _, supportedEncoding := range supportedEncodings {
		if encoding == supportedEncoding {
			hasSupportedEncoding = true
			break
		}
	}
	if !hasSupportedEncoding {
		return fmt.Errorf("unsupported encoding: %s", pb.Encoding_name[int32(encoding)])
	}
	for _, m := range models {
		isSupported := false
		for _, supportedModel := range s.model.modelData {
			if reflect.DeepEqual(m, supportedModel) {
				isSupported = true
				break
			}
		}
		if !isSupported {
			return fmt.Errorf("unsupported model: %v", m)
		}
	}
	return nil
}

// getGNMIServiceVersion returns a pointer to the gNMI service version string.
// The method is non-trivial because of the way it is defined in the proto file.
func getGNMIServiceVersion() (*string, error) {
	gzB, _ := (&pb.Update{}).Descriptor()
	r, err := gzip.NewReader(bytes.NewReader(gzB))
	if err != nil {
		return nil, fmt.Errorf("error in initializing gzip reader: %v", err)
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error in reading gzip data: %v", err)
	}
	desc := &dpb.FileDescriptorProto{}
	if err := proto.Unmarshal(b, desc); err != nil {
		return nil, fmt.Errorf("error in unmarshaling proto: %v", err)
	}
	ver, err := proto.GetExtension(desc.Options, pb.E_GnmiService)
	if err != nil {
		return nil, fmt.Errorf("error in getting version from proto extension: %v", err)
	}
	return ver.(*string), nil
}

// gnmiFullPath builds the full path from the prefix and path.
func gnmiFullPath(prefix, path *pb.Path) *pb.Path {
	fullPath := &pb.Path{Origin: path.Origin}
	if path.GetElement() != nil {
		fullPath.Element = append(prefix.GetElement(), path.GetElement()...)
	}
	if path.GetElem() != nil {
		fullPath.Elem = append(prefix.GetElem(), path.GetElem()...)
	}
	return fullPath
}

// processOperation validates the operation to be applied to the device, updates
// the json tree of the config struct, and then calls the callback function to
// apply the operation to the device hardware.
func (s *Server) processOperation(jsonTree *map[string]interface{}, op pb.UpdateResult_Operation, prefix, path *pb.Path, val *pb.TypedValue) (*pb.UpdateResult, error) {
	if !reflect.DeepEqual(path, pbRootPath) || op != pb.UpdateResult_REPLACE {
		return nil, status.Error(codes.Unimplemented, "only support setting the entire config tree with one REPLACE")
	}

	node, stat := ygotutils.NewNode(s.model.structRootType, path)
	if stat.GetCode() != int32(cpb.Code_OK) {
		return nil, status.Errorf(codes.NotFound, "path %v is not found in the config structure: %v", path, stat)
	}
	nodeStruct, ok := node.(ygot.ValidatedGoStruct)
	if !ok {
		msg := "the created node structure is not compatible to ygot.ValidatedGoStruct interface"
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	if err := s.model.jsonUnmarshaler(val.GetJsonIetfVal(), nodeStruct); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unmarshaling json data to config struct fails: %v", err)
	}
	if err := nodeStruct.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "config data validation fails: %v", err)
	}
	if s.callback != nil {
		if err := s.callback(nodeStruct); err != nil {
			return nil, status.Errorf(codes.Aborted, "error in applying operation to device: %v", err)
		}
	}

	tree, err := ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{})
	if err != nil {
		msg := fmt.Sprintf("error in constructing IETF JSON tree from config struct: %v", err)
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	*jsonTree = tree
	return &pb.UpdateResult{Path: path, Op: op}, nil
}

// Capabilities returns supported encodings and supported models.
func (s *Server) Capabilities(ctx context.Context, req *pb.CapabilityRequest) (*pb.CapabilityResponse, error) {
	ver, err := getGNMIServiceVersion()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error in getting gnmi service version: %v", err)
	}
	return &pb.CapabilityResponse{
		SupportedModels:    s.model.modelData,
		SupportedEncodings: supportedEncodings,
		GNMIVersion:        *ver,
	}, nil
}

// Get implements the Get RPC in gNMI spec.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.GetType() != pb.GetRequest_ALL {
		return nil, status.Errorf(codes.Unimplemented, "unsupported request type: %s", pb.GetRequest_DataType_name[int32(req.GetType())])
	}
	if err := s.checkEncodingAndModel(req.GetEncoding(), req.GetUseModels()); err != nil {
		return nil, status.Error(codes.Unimplemented, err.Error())
	}

	prefix := req.GetPrefix()
	paths := req.GetPath()
	notifications := make([]*pb.Notification, len(paths))

	s.mu.RLock()
	defer s.mu.RUnlock()

	for i, path := range paths {
		// Get schema node for path from config struct
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}
		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}
		node, stat := ygotutils.GetNode(s.model.schemaTreeRoot, s.config, fullPath)
		if stat.GetCode() != int32(cpb.Code_OK) {
			return nil, status.Errorf(codes.NotFound, "path %v not found", fullPath)
		}

		ts := time.Now().UnixNano()

		nodeStruct, ok := node.(ygot.GoStruct)
		// Leaf node case
		if !ok {
			if reflect.ValueOf(node).Kind() == reflect.Ptr {
				node = reflect.ValueOf(node).Elem().Interface()
			}
			val, err := value.FromScalar(node)
			if err != nil {
				msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
				log.Error(msg)
				return nil, status.Error(codes.Internal, msg)
			}
			update := &pb.Update{Path: path, Val: val}
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}

		// Subtree case
		if len(req.GetUseModels()) != len(s.model.modelData) || req.GetEncoding() != pb.Encoding_JSON_IETF {
			return nil, status.Error(codes.Unimplemented, "serializing config subtree into multiple leaf updates is unsupported")
		}
		jsonTree, err := ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{})
		if err != nil {
			msg := fmt.Sprintf("error in constructing IETF JSON tree from requested node: %v", err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		if len(path.GetElem()) == 1 {
			jsonTree = jsonTree[path.GetElem()[0].Name].(map[string]interface{})
		}
		jsonDump, err := json.Marshal(jsonTree)
		if err != nil {
			msg := fmt.Sprintf("error in marshaling IETF JSON tree to bytes: %v", err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
		update := &pb.Update{
			Path: path,
			Val: &pb.TypedValue{
				Value: &pb.TypedValue_JsonIetfVal{
					JsonIetfVal: jsonDump,
				},
			},
		}
		notifications[i] = &pb.Notification{
			Timestamp: ts,
			Prefix:    prefix,
			Update:    []*pb.Update{update},
		}
	}
	return &pb.GetResponse{Notification: notifications}, nil
}

// Set implements the Set RPC in gNMI spec.
func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonTree, err := ygot.ConstructIETFJSON(s.config, &ygot.RFC7951JSONConfig{})
	if err != nil {
		msg := fmt.Sprintf("error in constructing IETF JSON tree from config struct: %v", err)
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}

	prefix := req.GetPrefix()
	var results []*pb.UpdateResult

	for _, path := range req.GetDelete() {
		res, err := s.processOperation(&jsonTree, pb.UpdateResult_DELETE, prefix, path, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	for _, upd := range req.GetReplace() {
		res, err := s.processOperation(&jsonTree, pb.UpdateResult_REPLACE, prefix, upd.GetPath(), upd.GetVal())
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	for _, upd := range req.GetUpdate() {
		res, err := s.processOperation(&jsonTree, pb.UpdateResult_UPDATE, prefix, upd.GetPath(), upd.GetVal())
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	jsonDump, err := json.Marshal(jsonTree)
	if err != nil {
		msg := fmt.Sprintf("error in marshaling IETF JSON tree to bytes: %v", err)
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	rootStruct, err := s.model.NewConfigStruct(jsonDump)
	if err != nil {
		msg := fmt.Sprintf("error in creating config struct from IETF JSON data: %v", err)
		log.Error(msg)
		return nil, status.Error(codes.Internal, msg)
	}
	s.config = rootStruct
	return &pb.SetResponse{
		Prefix:   req.GetPrefix(),
		Response: results,
	}, nil
}

// Subscribe method is not implemented.
func (s *Server) Subscribe(stream pb.GNMI_SubscribeServer) error {
	return status.Error(codes.Unimplemented, "Subscribe is unsupported by the config server")
}
