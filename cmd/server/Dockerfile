FROM golang:1.12 AS builder
COPY . /src/peerless
ENV GOPATH=/gopath
WORKDIR /src/peerless
RUN go get ./...
RUN go install ./cmd/server

FROM ubuntu:18.04
COPY --from=builder /gopath/bin/server /bin/server
ENTRYPOINT ["/bin/server"]
