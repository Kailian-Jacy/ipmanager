FROM golang:1.19 AS BUILDER

WORKDIR /go/src/app
COPY . .

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.io,direct \
    GOOS=linux \
    GOARCH=amd64

RUN go mod init ipmanager \
    && go mod tidy \
    && go build -o /go/bin/app/ipmannager

FROM alpine:latest
ARG CONFIG
COPY --from=BUILDER /go/bin/app/ipmannager /ipmannager
COPY ./Config/config.json /config.json
CMD ["/ipmannager --config ${CONFIG}"]