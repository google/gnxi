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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("couldn't get home directory: %v", err)
	}
	filesDir := path.Join(homeDir, ".gnxi")
	for test := range tests {
		safeWriteDockerfile(test, path.Join(filesDir, fmt.Sprintf("%s.Dockerfile", test)))
	}
	return filesDir
}

func safeWriteDockerfile(name, loc string) {
	_, existErr := os.Stat(loc)
	if errors.Is(existErr, os.ErrNotExist) {
		content := fmt.Sprintf(dockerfile, name, name, name, name)
		if err := os.MkdirAll(path.Dir(loc), 0755); err != nil {
			log.Errorf("couldn't create dockerfile directory", err)
		}
		file, err := os.Create(loc)
		if err != nil {
			log.Errorf("couldn't create dockerfile", err)
		}
		if _, err = file.WriteString(content); err != nil {
			log.Errorf("couldn't write dockerfile", err)
		}
	}
}
