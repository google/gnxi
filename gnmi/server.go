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

// Package gnmi implements a gnmi server to mock a device with YANG models.
package gnmi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/openconfig/gnmi/coalesce"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/ygot/util"
	"github.com/openconfig/ygot/ygot"
	"github.com/openconfig/ygot/ytypes"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	pb "github.com/openconfig/gnmi/proto/gnmi"
)

// ConfigCallback is the signature of the function to apply a validated config to the physical device.
type ConfigCallback func(ygot.ValidatedGoStruct) error

var (
	pbRootPath         = &pb.Path{}
	supportedEncodings = []pb.Encoding{pb.Encoding_JSON, pb.Encoding_JSON_IETF}
	subscribeSync      = &pb.SubscribeResponse{Response: &pb.SubscribeResponse_SyncResponse{SyncResponse: true}}
)

// Server struct maintains the data structure for device config and implements the interface of gnmi server. It supports Capabilities, Get, and Set APIs.
// Typical usage:
//
//	g := grpc.NewServer()
//	s, err := Server.NewServer(model, config, callback)
//	pb.NewServer(g, s)
//	reflection.Register(g)
//	listen, err := net.Listen("tcp", ":8080")
//	g.Serve(listen)
//
// For a real device, apply the config changes to the hardware in the callback function.
// Arguments:
//
//	newConfig: new root config to be applied on the device.
//
//	func callback(newConfig ygot.ValidatedGoStruct) error {
//			// Apply the config to your device and return nil if success. return error if fails.
//			//
//			// Do something ...
//	}
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
	s := &Server{
		model:    model,
		config:   rootStruct,
		callback: callback,
	}
	if config != nil && s.callback != nil {
		if err := s.callback(rootStruct); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// checkEncodingAndModel checks whether encoding and models are supported by the server. Return error if anything is unsupported.
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

// doDelete deletes the path from the json tree if the path exists. If success,
// it calls the callback function to apply the change to the device hardware.
func (s *Server) doDelete(jsonTree map[string]interface{}, prefix, path *pb.Path) (*pb.UpdateResult, error) {
	// Update json tree of the device config
	var curNode interface{} = jsonTree
	pathDeleted := false
	fullPath := gnmiFullPath(prefix, path)
	schema := s.model.schemaTreeRoot
	for i, elem := range fullPath.Elem { // Delete sub-tree or leaf node.
		node, ok := curNode.(map[string]interface{})
		if !ok {
			break
		}

		// Delete node
		if i == len(fullPath.Elem)-1 {
			if elem.GetKey() == nil {
				delete(node, elem.Name)
				pathDeleted = true
				break
			}
			pathDeleted = deleteKeyedListEntry(node, elem)
			break
		}

		if curNode, schema = getChildNode(node, schema, elem, false); curNode == nil {
			break
		}
	}
	if reflect.DeepEqual(fullPath, pbRootPath) { // Delete root
		for k := range jsonTree {
			delete(jsonTree, k)
		}
	}

	// Apply the validated operation to the config tree and device.
	if pathDeleted {
		newConfig, err := s.toGoStruct(jsonTree)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if s.callback != nil {
			if applyErr := s.callback(newConfig); applyErr != nil {
				if rollbackErr := s.callback(s.config); rollbackErr != nil {
					return nil, status.Errorf(codes.Internal, "error in rollback the failed operation (%v): %v", applyErr, rollbackErr)
				}
				return nil, status.Errorf(codes.Aborted, "error in applying operation to device: %v", applyErr)
			}
		}
	}
	return &pb.UpdateResult{
		Path: path,
		Op:   pb.UpdateResult_DELETE,
	}, nil
}

// doReplaceOrUpdate validates the replace or update operation to be applied to
// the device, modifies the json tree of the config struct, then calls the
// callback function to apply the operation to the device hardware.
func (s *Server) doReplaceOrUpdate(jsonTree map[string]interface{}, op pb.UpdateResult_Operation, prefix, path *pb.Path, val *pb.TypedValue) (*pb.UpdateResult, error) {
	// Validate the operation.
	fullPath := gnmiFullPath(prefix, path)
	emptyNode, _, err := ytypes.GetOrCreateNode(s.model.schemaTreeRoot, s.model.newRootValue(), fullPath)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "path %v is not found in the config structure: %v", fullPath, err)
	}
	var nodeVal interface{}
	nodeStruct, ok := emptyNode.(ygot.ValidatedGoStruct)
	if ok {
		if err := s.model.jsonUnmarshaler(val.GetJsonIetfVal(), nodeStruct); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "unmarshaling json data to config struct fails: %v", err)
		}
		if err := nodeStruct.Validate(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "config data validation fails: %v", err)
		}
		var err error
		if nodeVal, err = ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{}); err != nil {
			msg := fmt.Sprintf("error in constructing IETF JSON tree from config struct: %v", err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}
	} else {
		var err error
		if nodeVal, err = value.ToScalar(val); err != nil {
			return nil, status.Errorf(codes.Internal, "cannot convert leaf node to scalar type: %v", err)
		}
	}

	// Update json tree of the device config.
	var curNode interface{} = jsonTree
	schema := s.model.schemaTreeRoot
	for i, elem := range fullPath.Elem {
		switch node := curNode.(type) {
		case map[string]interface{}:
			// Set node value.
			if i == len(fullPath.Elem)-1 {
				if elem.GetKey() == nil {
					if grpcStatusError := setPathWithoutAttribute(op, node, elem, nodeVal); grpcStatusError != nil {
						return nil, grpcStatusError
					}
					break
				}
				if grpcStatusError := setPathWithAttribute(op, node, elem, nodeVal); grpcStatusError != nil {
					return nil, grpcStatusError
				}
				break
			}

			if curNode, schema = getChildNode(node, schema, elem, true); curNode == nil {
				return nil, status.Errorf(codes.NotFound, "path elem not found: %v", elem)
			}
		case []interface{}:
			return nil, status.Errorf(codes.NotFound, "incompatible path elem: %v", elem)
		default:
			return nil, status.Errorf(codes.Internal, "wrong node type: %T", curNode)
		}
	}
	if reflect.DeepEqual(fullPath, pbRootPath) { // Replace/Update root.
		if op == pb.UpdateResult_UPDATE {
			return nil, status.Error(codes.Unimplemented, "update the root of config tree is unsupported")
		}
		nodeValAsTree, ok := nodeVal.(map[string]interface{})
		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "expect a tree to replace the root, got a scalar value: %T", nodeVal)
		}
		for k := range jsonTree {
			delete(jsonTree, k)
		}
		for k, v := range nodeValAsTree {
			jsonTree[k] = v
		}
	}
	newConfig, err := s.toGoStruct(jsonTree)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Apply the validated operation to the device.
	if s.callback != nil {
		if applyErr := s.callback(newConfig); applyErr != nil {
			if rollbackErr := s.callback(s.config); rollbackErr != nil {
				return nil, status.Errorf(codes.Internal, "error in rollback the failed operation (%v): %v", applyErr, rollbackErr)
			}
			return nil, status.Errorf(codes.Aborted, "error in applying operation to device: %v", applyErr)
		}
	}
	return &pb.UpdateResult{
		Path: path,
		Op:   op,
	}, nil
}

func (s *Server) toGoStruct(jsonTree map[string]interface{}) (ygot.ValidatedGoStruct, error) {
	jsonDump, err := json.Marshal(jsonTree)
	if err != nil {
		return nil, fmt.Errorf("error in marshaling IETF JSON tree to bytes: %v", err)
	}
	goStruct, err := s.model.NewConfigStruct(jsonDump)
	if err != nil {
		return nil, fmt.Errorf("error in creating config struct from IETF JSON data: %v", err)
	}
	return goStruct, nil
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

// deleteKeyedListEntry deletes the keyed list entry from node that matches the
// path elem. If the entry is the only one in keyed list, deletes the entire
// list. If the entry is found and deleted, the function returns true. If it is
// not found, the function returns false.
func deleteKeyedListEntry(node map[string]interface{}, elem *pb.PathElem) bool {
	curNode, ok := node[elem.Name]
	if !ok {
		return false
	}

	keyedList, ok := curNode.([]interface{})
	if !ok {
		return false
	}
	for i, n := range keyedList {
		m, ok := n.(map[string]interface{})
		if !ok {
			log.Errorf("expect map[string]interface{} for a keyed list entry, got %T", n)
			return false
		}
		keyMatching := true
		for k, v := range elem.Key {
			attrVal, ok := m[k]
			if !ok {
				return false
			}
			if v != fmt.Sprintf("%v", attrVal) {
				keyMatching = false
				break
			}
		}
		if keyMatching {
			listLen := len(keyedList)
			if listLen == 1 {
				delete(node, elem.Name)
				return true
			}
			keyedList[i] = keyedList[listLen-1]
			node[elem.Name] = keyedList[0 : listLen-1]
			return true
		}
	}
	return false
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

// isNIl checks if an interface is nil or its value is nil.
func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch kind := reflect.ValueOf(i).Kind(); kind {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	default:
		return false
	}
}

// setPathWithAttribute replaces or updates a child node of curNode in the IETF
// JSON config tree, where the child node is indexed by pathElem with attribute.
// The function returns grpc status error if unsuccessful.
func setPathWithAttribute(op pb.UpdateResult_Operation, curNode map[string]interface{}, pathElem *pb.PathElem, nodeVal interface{}) error {
	nodeValAsTree, ok := nodeVal.(map[string]interface{})
	if !ok {
		return status.Errorf(codes.InvalidArgument, "expect nodeVal is a json node of map[string]interface{}, received %T", nodeVal)
	}
	m := getKeyedListEntry(curNode, pathElem, true)
	if m == nil {
		return status.Errorf(codes.NotFound, "path elem not found: %v", pathElem)
	}
	if op == pb.UpdateResult_REPLACE {
		for k := range m {
			delete(m, k)
		}
	}
	for attrKey, attrVal := range pathElem.GetKey() {
		m[attrKey] = attrVal
		if asNum, err := strconv.ParseFloat(attrVal, 64); err == nil {
			m[attrKey] = asNum
		}
		for k, v := range nodeValAsTree {
			if k == attrKey && fmt.Sprintf("%v", v) != attrVal {
				return status.Errorf(codes.InvalidArgument, "invalid config data: %v is a path attribute", k)
			}
		}
	}
	for k, v := range nodeValAsTree {
		m[k] = v
	}
	return nil
}

// setPathWithoutAttribute replaces or updates a child node of curNode in the
// IETF config tree, where the child node is indexed by pathElem without
// attribute. The function returns grpc status error if unsuccessful.
func setPathWithoutAttribute(op pb.UpdateResult_Operation, curNode map[string]interface{}, pathElem *pb.PathElem, nodeVal interface{}) error {
	target, hasElem := curNode[pathElem.Name]
	nodeValAsTree, nodeValIsTree := nodeVal.(map[string]interface{})
	if op == pb.UpdateResult_REPLACE || !hasElem || !nodeValIsTree {
		curNode[pathElem.Name] = nodeVal
		return nil
	}
	targetAsTree, ok := target.(map[string]interface{})
	if !ok {
		return status.Errorf(codes.Internal, "error in setting path: expect map[string]interface{} to update, got %T", target)
	}
	for k, v := range nodeValAsTree {
		targetAsTree[k] = v
	}
	return nil
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
		// Get schema node for path from config struct.
		fullPath := path
		if prefix != nil {
			fullPath = gnmiFullPath(prefix, path)
		}
		if fullPath.GetElem() == nil && fullPath.GetElement() != nil {
			return nil, status.Error(codes.Unimplemented, "deprecated path element type is unsupported")
		}
		nodes, err := ytypes.GetNode(s.model.schemaTreeRoot, s.config, fullPath)
		if len(nodes) == 0 || err != nil || util.IsValueNil(nodes[0].Data) {
			return nil, status.Errorf(codes.NotFound, "path %v not found: %v", fullPath, err)
		}
		node := nodes[0].Data

		ts := time.Now().UnixNano()

		nodeStruct, ok := node.(ygot.GoStruct)
		// Return leaf node.
		if !ok {
			var val *pb.TypedValue
			switch kind := reflect.ValueOf(node).Kind(); kind {
			case reflect.Ptr, reflect.Interface:
				var err error
				val, err = value.FromScalar(reflect.ValueOf(node).Elem().Interface())
				if err != nil {
					msg := fmt.Sprintf("leaf node %v does not contain a scalar type value: %v", path, err)
					log.Error(msg)
					return nil, status.Error(codes.Internal, msg)
				}
			case reflect.Int64:
				enumMap, ok := s.model.enumData[reflect.TypeOf(node).Name()]
				if !ok {
					return nil, status.Error(codes.Internal, "not a GoStruct enumeration type")
				}
				val = &pb.TypedValue{
					Value: &pb.TypedValue_StringVal{
						StringVal: enumMap[reflect.ValueOf(node).Int()].Name,
					},
				}
			default:
				return nil, status.Errorf(codes.Internal, "unexpected kind of leaf node type: %v %v", node, kind)
			}

			update := &pb.Update{Path: path, Val: val}
			notifications[i] = &pb.Notification{
				Timestamp: ts,
				Prefix:    prefix,
				Update:    []*pb.Update{update},
			}
			continue
		}

		if req.GetUseModels() != nil {
			return nil, status.Errorf(codes.Unimplemented, "filtering Get using use_models is unsupported, got: %v", req.GetUseModels())
		}

		// Return IETF JSON by default.
		jsonEncoder := func() (map[string]interface{}, error) {
			return ygot.ConstructIETFJSON(nodeStruct, &ygot.RFC7951JSONConfig{AppendModuleName: true})
		}
		jsonType := "IETF"
		buildUpdate := func(b []byte) *pb.Update {
			return &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonIetfVal{JsonIetfVal: b}}}
		}

		if req.GetEncoding() == pb.Encoding_JSON {
			jsonEncoder = func() (map[string]interface{}, error) {
				return ygot.ConstructInternalJSON(nodeStruct)
			}
			jsonType = "Internal"
			buildUpdate = func(b []byte) *pb.Update {
				return &pb.Update{Path: path, Val: &pb.TypedValue{Value: &pb.TypedValue_JsonVal{JsonVal: b}}}
			}
		}

		jsonTree, err := jsonEncoder()
		if err != nil {
			msg := fmt.Sprintf("error in constructing %s JSON tree from requested node: %v", jsonType, err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		jsonDump, err := json.Marshal(jsonTree)
		if err != nil {
			msg := fmt.Sprintf("error in marshaling %s JSON tree to bytes: %v", jsonType, err)
			log.Error(msg)
			return nil, status.Error(codes.Internal, msg)
		}

		update := buildUpdate(jsonDump)
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
		res, grpcStatusError := s.doDelete(jsonTree, prefix, path)
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetReplace() {
		res, grpcStatusError := s.doReplaceOrUpdate(jsonTree, pb.UpdateResult_REPLACE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
		}
		results = append(results, res)
	}
	for _, upd := range req.GetUpdate() {
		res, grpcStatusError := s.doReplaceOrUpdate(jsonTree, pb.UpdateResult_UPDATE, prefix, upd.GetPath(), upd.GetVal())
		if grpcStatusError != nil {
			return nil, grpcStatusError
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

// InternalUpdate is an experimental feature to let the server update its
// internal states. Use it with your own risk.
func (s *Server) InternalUpdate(fp func(config ygot.ValidatedGoStruct) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fp(s.config)
}

// Set implements the Subscribe gNMI RPC.
func (s *Server) Subscribe(stream pb.GNMI_SubscribeServer) error {
	var err error

	c := streamClient{stream: stream}
	c.sr, err = stream.Recv()

	switch {
	case err == io.EOF:
		return nil
	case err != nil:
		return err
	}

	if c.sr.GetSubscribe() == nil {
		return status.Errorf(codes.InvalidArgument, "request must contain a subscription %#v", c.sr)
	}
	if c.sr.GetSubscribe().GetAllowAggregation() {
		return status.Error(codes.Unimplemented, "aggregation is not supported")
	}
	if c.sr.GetSubscribe().GetUseModels() != nil {
		return status.Errorf(codes.Unimplemented, "subscription using use_models is not supported")
	}

	if err = s.checkEncodingAndModel(c.sr.GetSubscribe().GetEncoding(), nil); err != nil {
		return status.Error(codes.Unimplemented, err.Error())
	}

	mode := c.sr.GetSubscribe().Mode

	// This error channel accepts errors from all goroutines spawned.
	errC := make(chan error, 3)
	c.errC = errC

	c.msgQ = coalesce.NewQueue()
	defer c.msgQ.Close()

	switch mode {
	case pb.SubscriptionList_ONCE:
		go func() {
			s.doOnceSubscription(&c)
		}()
	case pb.SubscriptionList_POLL:
		return status.Error(codes.Unimplemented, "mode POLL is not implemented.")
	case pb.SubscriptionList_STREAM:
		return status.Error(codes.Unimplemented, "mode SAMPLE is not implemented.")
	default:
		return status.Errorf(codes.InvalidArgument, "subscription mode %v not recognized", mode)
	}
	go s.doSendSubscriptionMsgs(&c)
	return <-errC
}

// streamClient represents a Streaming client.
type streamClient struct {
	sr     *pb.SubscribeRequest
	stream pb.GNMI_SubscribeServer
	errC   chan<- error
	msgQ   *coalesce.Queue
}

// subscribeSyncToken signals doSendSubscriptionMsgs to send subscribeSync.
type subscribeSyncToken struct{}

// doOnceSubscription processes a ONCE Subscription. It produces a single
// Notification message for all subscribed paths.
func (s *Server) doOnceSubscription(c *streamClient) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var updates []*pb.Update
	prefix := c.sr.GetSubscribe().GetPrefix()

	if !c.sr.GetSubscribe().GetUpdatesOnly() {
		for _, subscription := range c.sr.GetSubscribe().GetSubscription() {
			fullPath := subscription.GetPath()
			if prefix != nil {
				fullPath = gnmiFullPath(prefix, fullPath)
			}
			updts, err := s.updatesFromNode(fullPath)
			if err != nil {
				c.errC <- err
				return
			}
			updates = append(updates, updts...)
		}
		c.msgQ.Insert(
			&pb.Notification{
				Timestamp: time.Now().UnixNano(),
				Update:    updates,
			})
	}
	c.msgQ.Insert(subscribeSyncToken{})
	c.msgQ.Close()
}

// doSendSubscriptionMsgs monitors the message queue and sends
// Subscription Responses to the client.
func (s *Server) doSendSubscriptionMsgs(c *streamClient) {
	subscribeTimeout := time.NewTimer(time.Second * 5)
	subscribeTimeout.Stop()

	// If a send doesn't occur within the timeout.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-subscribeTimeout.C:
			c.errC <- errors.New("subscription timed out while sending")
		case <-done:
		}
	}()

	for {
		item, _, err := c.msgQ.Next(c.stream.Context())
		if err != nil {
			if coalesce.IsClosedQueue(err) {
				c.errC <- nil
			} else {
				c.errC <- err
			}
			return
		}
		if _, ok := item.(subscribeSyncToken); ok {
			if err = c.stream.Send(subscribeSync); err != nil {
				c.errC <- err
				return
			}
			continue
		}
		n, ok := item.(*pb.Notification)
		if !ok || n == nil {
			c.errC <- status.Errorf(codes.Internal, "invalid notification message: %v", item)
			return
		}
		response := &pb.SubscribeResponse{
			Response: &pb.SubscribeResponse_Update{
				Update: n,
			},
		}
		subscribeTimeout.Reset(time.Second * 5)
		defer subscribeTimeout.Stop()
		err = c.stream.Send(response)
		if err != nil {
			c.errC <- err
			return
		}
	}
}

// updatesFromNode returns a list of Update messages for the leaf nodes found,
// starting to walk the tree at the path.
func (s *Server) updatesFromNode(fullPath *pb.Path) ([]*pb.Update, error) {
	var updates []*pb.Update

	nodes, err := ytypes.GetNode(s.model.schemaTreeRoot, s.config, fullPath, &ytypes.GetHandleWildcards{})
	if len(nodes) == 0 || err != nil || util.IsValueNil(nodes[0].Data) {
		return nil, status.Errorf(codes.NotFound, "path %v not found: %v", fullPath, err)
	}
	for _, node := range nodes {
		data := node.Data
		nodeStruct, isContainer := data.(ygot.GoStruct)

		if isContainer {
			n, err := ygot.TogNMINotifications(nodeStruct, time.Now().UnixNano(),
				ygot.GNMINotificationsConfig{
					UsePathElem:    true,
					PathElemPrefix: fullPath.GetElem(),
				})
			if err != nil {
				return nil, err
			}
			for _, up := range n[0].Update {
				up.Path = &pb.Path{Elem: append(node.Path.GetElem(), up.Path.GetElem()...)}
			}
			updates = append(updates, n[0].Update...)
			continue
		}

		var val *pb.TypedValue
		switch kind := reflect.ValueOf(data).Kind(); kind {
		case reflect.Ptr, reflect.Interface:
			var err error
			val, err = value.FromScalar(reflect.ValueOf(data).Elem().Interface())
			if err != nil {
				return nil, status.Errorf(codes.Internal, "leaf node %v does not contain a scalar type value: %v", fullPath, err)
			}
		case reflect.Int64:
			enumMap, ok := s.model.enumData[reflect.TypeOf(data).Name()]
			if !ok {
				return nil, status.Error(codes.Internal, "not a GoStruct enumeration type")
			}
			val = &pb.TypedValue{
				Value: &pb.TypedValue_StringVal{
					StringVal: enumMap[reflect.ValueOf(data).Int()].Name,
				},
			}
		default:
			return nil, status.Errorf(codes.Internal, "unexpected kind of leaf node type: %v %v", node, kind)
		}
		updates = append(updates, &pb.Update{Path: node.Path, Val: val})
	}

	return updates, nil
}
