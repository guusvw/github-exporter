FROM golang:1.12 as build
LABEL maintainer "Guus van Weelden"

RUN apk --no-cache add ca-certificates \
     && apk --no-cache add --virtual build-deps git

COPY ./ /go/src/github.com/guusvw/github-exporter
WORKDIR /go/src/github.com/guusvw/github-exporter

RUN go get \
 && go test ./... \
 && go build -o /bin/main

FROM alpine:3.6

RUN apk --no-cache add ca-certificates \
     && addgroup exporter \
     && adduser -S -G exporter exporter
USER exporter
COPY --from=build /bin/main /bin/main
ENV LISTEN_PORT=9171
EXPOSE 9171
ENTRYPOINT [ "/bin/main" ]
