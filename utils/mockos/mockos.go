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
package mockos

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils/mockos/pb"
)

type OS struct {
	pb.MockOS
}

// Hash calculates the hash of the MockOS and embeds it in the package.
func (os *OS) Hash() {
	temp := calcHash(os)
	os.MockOS.Hash = temp
}

// CheckHash recalculates the hash of the MockOS and checks against the embedded hash.
func (os *OS) CheckHash() bool {
	temp := calcHash(os)
	return bytes.Compare(os.MockOS.Hash, temp) == 0
}

// GenerateOS creates a Mock OS file for gNOI client and target use.
func GenerateOS(filename, version string, size string, supported bool) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return errors.New("File already exists")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	bufferSize, err := humanize.ParseBytes(size)
	if err != nil {
		return err
	}
	buf := make([]byte, bufferSize)
	rand.Read(buf)
	cookieBuf := make([]byte, 16)
	rand.Read(cookieBuf)
	mockOs := &OS{MockOS: pb.MockOS{
		Version:   version,
		Cookie:    fmt.Sprintf("%x", cookieBuf),
		Padding:   buf,
		Supported: supported,
	}}
	mockOs.Hash()
	out, err := proto.Marshal(&mockOs.MockOS)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	if _, err = writer.Write(out); err != nil {
		return err
	}
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

// ValidateOS unmarshals the serialized OS proto and verifys the OS package's integrity.
func ValidateOS(filename string) (*OS, error) {
	mockOs := &OS{}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	proto.Unmarshal(buf.Bytes(), &mockOs.MockOS)
	if mockOs.CheckHash() {
		return mockOs, nil
	}
	return nil, errors.New("Hash check failed!")
}

func calcHash(os *OS) []byte {
	var supported byte
	if os.MockOS.Supported {
		supported = byte(1)
	} else {
		supported = byte(0)
	}
	temp := md5.Sum(append([]byte(os.MockOS.Version+os.MockOS.Cookie), append(os.MockOS.Padding, supported)...))
	return temp[:]
}
