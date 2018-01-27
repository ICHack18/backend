FROM golang:1.9-alpine

ADD . /go/src/safespace
WORKDIR /go/src
RUN go get safespace
RUN go install safespace

ENTRYPOINT /go/bin/safespace

EXPOSE 2048
