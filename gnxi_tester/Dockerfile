FROM golang:1.14-alpine AS dev

WORKDIR /workspace

RUN GO111MODULE=on go get github.com/cortesi/modd/cmd/modd

COPY ../go.mod .
COPY ../go.sum .

RUN go mod download

COPY . .

RUN go install github.com/google/gnxi/gnxi_tester

CMD [ "go", "run", "gnxi_tester.go" ]


FROM alpine

WORKDIR /bin

COPY --from=dev /go/bin/gnxi_tester ./gnxi_tester

CMD [ "gnxi_tester", "web" ]