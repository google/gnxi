FROM golang:1.14-alpine AS dev

WORKDIR /workspace

RUN apk update && apk add git

RUN GO111MODULE=on go get github.com/cortesi/modd/cmd/modd

COPY go.mod .
COPY go.sum .

RUN go mod download 

COPY gnxi_tester gnxi_tester

WORKDIR /workspace/gnxi_tester

RUN GO111MODULE=on go install github.com/google/gnxi/gnxi_tester

CMD [ "go", "run", "gnxi_tester.go" ]


FROM alpine

RUN apk add docker

WORKDIR /bin

COPY --from=dev /go/bin/gnxi_tester ./gnxi_tester

CMD [ "gnxi_tester", "web" ]
