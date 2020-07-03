/* Copyright 2020 Google Inc.
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
package osgenerator

import (
	"bufio"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils/osgenerator/pb"
)

// GenerateOS creates a Mock OS file for gNOI client and target use.
func GenerateOS(filename, version string, unit rune, size int, supported bool) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return errors.New("File already exists")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferSize, err := anyToBytes(size, unit)
	if err != nil {
		return err
	}
	buf := make([]byte, bufferSize)
	rand.Read(buf)
	hash := md5.Sum(buf)
	cookieBuf := make([]byte, 16)
	rand.Read(cookieBuf)
	mockOs := &pb.MockOS{
		Version:   version,
		Cookie:    fmt.Sprintf("%x", cookieBuf),
		Data:      buf,
		Hash:      hash[:],
		Supported: supported,
	}
	out, err := proto.Marshal(mockOs)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	_, err = writer.Write(out)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

// anyToBytes() converts the inputted filesize from Giga/Mega/Kilobytes to bytes.
func anyToBytes(size int, unit rune) (int, error) {
	var multiplier int
	switch unit {
	case 'B':
		multiplier = 1
	case 'K':
		multiplier = 1024
	case 'M':
		multiplier = int(math.Pow(1024, 2))
	case 'G':
		multiplier = int(math.Pow(1024, 3))
	default:
		return 0, errors.New("Unknown filesize unit specified")
	}
	return multiplier * size, nil
}
