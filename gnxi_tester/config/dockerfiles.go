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

package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	log "github.com/golang/glog"
)

const dockerfile = `
FROM golang:1.14-alpine AS build

RUN apk update && apk add git

RUN go get -v github.com/google/gnxi/%s

RUN go install github.com/google/gnxi/%s

FROM alpine

WORKDIR /bin

COPY --from=build /go/bin/%s ./%s

CMD [ "tail", "-f", "/dev/null" ]
`

func createDockerfiles(tests Tests) string {
	homeDir, err := userHome()
	if err != nil {
		log.Errorf("couldn't get home directory: %v", err)
	}
	filesDir := path.Join(homeDir, ".gnxi")
	for test := range tests {
		safeWriteDockerfile(test, path.Join(filesDir, fmt.Sprintf("%s.Dockerfile", test)))
	}
	return filesDir
}

var userHome = func() (string, error) {
	return os.UserHomeDir()
}

var safeWriteDockerfile = func(name, loc string) {
	_, existErr := os.Stat(loc)
	if errors.Is(existErr, os.ErrNotExist) {
		content := fmt.Sprintf(dockerfile, name, name, name, name)
		if err := os.MkdirAll(path.Dir(loc), 0755); err != nil {
			log.Errorf("couldn't create dockerfile directory: %v", err)
		}
		file, err := os.Create(loc)
		if err != nil {
			log.Errorf("couldn't create dockerfile: %v", err)
		}
		if _, err = file.WriteString(content); err != nil {
			log.Errorf("couldn't write dockerfile: %v", err)
		}
	}
}
