FROM golang:1.8.3-alpine AS builder

COPY . /go/src/github.com/hirsim/hello-server/
WORKDIR /go/src/github.com/hirsim/hello-server/

RUN go build -o hello

FROM alpine:3.6

LABEL maintainer "Hiroshi Nomura <n.hirsim@gmail.com>"

COPY --from=builder /go/src/github.com/hirsim/hello-server/hello /usr/local/bin/hello

EXPOSE 8080

ENTRYPOINT ["hello"]