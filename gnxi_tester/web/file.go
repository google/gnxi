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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	maxFileMemory              = 2000000000 // Represents a value of 2GB, when a file is uploaded contents exceeding this are stored as a temporary file.
	defaultDirectoryPermission = 0777       // This value is the default for directory creation, gives rwx to owner, group and others
)

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxFileMemory); err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	tempFile, tempFileHeader, err := r.FormFile("file")
	if err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	filesPath, err := initDirectory()
	if err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fileName := uuid.New().String()
	if extension := extractExtension(tempFileHeader.Filename); extension != "" {
		fileName = fileName + "." + extension
	}
	fullPath := path.Join(filesPath, fileName)
	if _, err = os.Stat(fullPath); !os.IsNotExist(err) {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
		return
	}
	file, err := os.Create(fullPath)
	if _, err := io.Copy(file, tempFile); err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	filenameRes := map[string]string{"filename": fileName}
	response, err := json.Marshal(filenameRes)
	if err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func handleFileDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["file"]
	filesDir, err := filesDir()
	if err != nil {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	filePath := path.Join(filesDir, fileName)
	if err := os.Remove(filePath); os.IsNotExist(err) {
		logErr(r.Header, err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}

func initDirectory() (string, error) {
	filesDir, err := filesDir()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		os.MkdirAll(filesDir, defaultDirectoryPermission)
	}
	return filesDir, nil
}

func filesDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filesDir := path.Join(homeDir, ".gnxi/files")
	return filesDir, nil
}

func extractExtension(fileName string) string {
	temp := strings.Split(fileName, ".")
	if len(temp) >= 2 {
		return temp[len(temp)-1]
	}
	return ""
}
