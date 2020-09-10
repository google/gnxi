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

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/gorilla/mux"
	"github.com/spf13/afero"
)

type fileUploadResponse struct {
	Filename string `json:"filename"`
}

func TestHandleFileUpload(t *testing.T) {
	tests := []struct {
		testName string
		filename string
	}{
		{
			testName: "Test upload small file",
			filename: "ca.crt",
		},
	}
	for _, test := range tests {
		f := FileHandler{FileSys: afero.NewMemMapFs()}
		t.Run(test.testName, func(t *testing.T) {
			rr := httptest.NewRecorder()
			buf := new(bytes.Buffer)
			formWriter := multipart.NewWriter(buf)
			fileWriter, err := formWriter.CreateFormFile("file", test.filename)
			if err != nil {
				t.Errorf("Error creating FormFileWriter: %v", err)
			}
			fileWriter.Write([]byte("These are some test bytes for the test file"))
			req, err := http.NewRequest("POST", "/file", buf)
			if err != nil {
				t.Errorf("Error creating request: %v", err)
			}
			formWriter.Close()
			req.Header.Add("Content-Type", formWriter.FormDataContentType())
			router := mux.NewRouter()
			router.HandleFunc("/file", f.handleFileUpload).Methods("POST")
			router.ServeHTTP(rr, req)
			res := fileUploadResponse{}
			json.Unmarshal(rr.Body.Bytes(), &res)
			if _, err := f.FileSys.Stat(path.Join(filesDir(), res.Filename)); os.IsNotExist(err) {
				t.Errorf("File upload failed: %v", err)
			}
		})
	}
}

func TestHandleFileDelete(t *testing.T) {
	tests := []struct {
		testName string
		filename string
		exists   bool
	}{
		{
			testName: "Delete a file that exists",
			filename: "file.exists",
			exists:   true,
		},
		{
			testName: "Delete a non existant file",
			filename: "nonexistant-file",
		},
	}
	for _, test := range tests {
		f := FileHandler{FileSys: afero.NewMemMapFs()}
		t.Run(test.testName, func(t *testing.T) {
			fPath := path.Join(filesDir(), test.filename)
			if test.exists {
				f.FileSys.Create(fPath)
			}
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("DELETE", fmt.Sprintf("/file/%s", test.filename), nil)
			if err != nil {
				t.Errorf("Error creating request: %v", err)
			}
			router := mux.NewRouter()
			router.HandleFunc("/file/{file}", f.handleFileDelete).Methods("DELETE")
			router.ServeHTTP(rr, req)
			if _, err := f.FileSys.Stat(fPath); !os.IsNotExist(err) {
				t.Errorf("File deletion failed: %v", err)
			}
		})
	}
}
