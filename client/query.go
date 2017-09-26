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

// Package client provides helper functions for client gNMI binaries.
package client

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/santhosh-tekuri/xpathparser"
)

var (
	querySeparator = flag.String("query_separator", ",", "Query separator character.")
)

// ParseQuery parses command line queries.
func ParseQuery(query string) []string {
	var queries []string
	for _, q := range strings.Split(query, *querySeparator) {
		queries = append(queries, q)
	}
	return queries
}

// ToGetRequest generates a gnmi GetRequest out of a list of xPaths.
func ToGetRequest(xpaths []string) (*gnmi.GetRequest, error) {
	getRequest := gnmi.GetRequest{Path: []*gnmi.Path{}}
	for _, xpath := range xpaths {

		expr, err := xpathparser.Parse(xpath)
		if err != nil {
			return nil, err
		}

		locationPath, ok := expr.(*xpathparser.LocationPath)
		if !ok {
			return nil, fmt.Errorf("error parsing LocationPath in xpath")
		}

		path := gnmi.Path{}
		for _, step := range locationPath.Steps {

			nameTest, ok := step.NodeTest.(*xpathparser.NameTest)
			if !ok {
				return nil, fmt.Errorf("error parsing NameTest in xpath")
			}

			pathElem := gnmi.PathElem{Name: nameTest.Local, Key: make(map[string]string)}
			for _, predicate := range step.Predicates {
				binaryExpression, ok := predicate.(*xpathparser.BinaryExpr)
				if !ok {
					return nil, fmt.Errorf("error parsing BinaryExpr in xpath")
				}

				var key string
				switch lhs := binaryExpression.LHS.(type) {
				case *xpathparser.LocationPath:
					if len(lhs.Steps) != 1 {
						return nil, fmt.Errorf("error in LHS length in xpath")
					}
					valNameTest, ok := lhs.Steps[0].NodeTest.(*xpathparser.NameTest)
					if !ok {
						return nil, fmt.Errorf("error parsing LHS NameTest in xpath")
					}
					key = valNameTest.Local
				case xpathparser.Number:
					key = strconv.FormatFloat(float64(lhs), 'f', -1, 64)
				case xpathparser.String:
					key = string(lhs)
				default:
					return nil, fmt.Errorf("error parsing LHS in xpath")
				}

				switch rhs := binaryExpression.RHS.(type) {
				case *xpathparser.LocationPath:
					if len(rhs.Steps) != 1 {
						return nil, fmt.Errorf("error in RHS length in xpath")
					}
					valNameTest, ok := rhs.Steps[0].NodeTest.(*xpathparser.NameTest)
					if !ok {
						return nil, fmt.Errorf("error parsing RHS NameTest in xpath")
					}
					pathElem.Key[key] = valNameTest.Local
				case xpathparser.Number:
					pathElem.Key[key] = strconv.FormatFloat(float64(rhs), 'f', -1, 64)
				case xpathparser.String:
					pathElem.Key[key] = string(rhs)
				default:
					return nil, fmt.Errorf("error parsing RHS in xpath")
				}
			}
			path.Elem = append(path.Elem, &pathElem)
		}
		getRequest.Path = append(getRequest.Path, &path)
	}
	return &getRequest, nil
}
