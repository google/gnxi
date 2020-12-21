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
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/golang/protobuf/proto"
	"github.com/google/gnxi/utils/mockos/pb"
)

const (
	cookie = "cookiestring"
)

type OS struct {
	pb.MockOS
}

// Hash calculates the hash of the MockOS and embeds it in the package.
func (os *OS) Hash() {
	os.MockOS.Hash = calcHash(os)
}

// CheckHash calculates the hash of the MockOS and checks against the embedded hash.
func (os *OS) CheckHash() bool {
	return bytes.Equal(os.MockOS.Hash, calcHash(os))
}

// GenerateOS creates a Mock OS file for gNOI target use.
func GenerateOS(filename, version, size, activationFailMessage string, incompatible bool) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return errors.New("File already exists")
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	out, err := packageOS(version, size, activationFailMessage, incompatible)
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

func packageOS(version, size, activationFailMessage string, incompatible bool) ([]byte, error) {
	bufferSize, err := humanize.ParseBytes(size)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, bufferSize)
	rand.Read(buf)
	mockOs := &OS{MockOS: pb.MockOS{
		Version:               version,
		Cookie:                cookie,
		Padding:               buf,
		Incompatible:          incompatible,
		ActivationFailMessage: activationFailMessage,
	}}
	mockOs.Hash()
	out, err := proto.Marshal(&mockOs.MockOS)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValidateOS unmarshals the serialized OS proto and verifies the OS package's integrity.
func ValidateOS(buf *bytes.Buffer) *OS {
	mockOs := &OS{MockOS: pb.MockOS{}}
	if err := proto.Unmarshal(buf.Bytes(), &mockOs.MockOS); err != nil {
		return nil
	}
	return mockOs
}

// calcHash returns the md5 hash of the OS.
func calcHash(os *OS) []byte {
	bb := []byte(os.MockOS.Version)
	bb = append(bb, []byte(os.MockOS.Cookie)...)
	bb = append(bb, []byte(os.MockOS.Padding)...)
	bb = append(bb, map[bool]byte{false: 0, true: 1}[os.MockOS.Incompatible])
	bb = append(bb, []byte(os.MockOS.ActivationFailMessage)...)
	hash := sha256.Sum256(bb)
	return hash[:]
}
