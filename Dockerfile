##
## -- runtime environment
##

FROM    golang:1.21-alpine AS build-env

ARG     MTR_BIN=mtr-exporter-0.1.0.linux.amd64

ADD     . /src/mtr-exporter
WORKDIR /src/mtr-exporter
RUN     apk add -U --no-cache make git
RUN     go build -o $MTR_BIN ./cmd/mtr-exporter

##
## -- runtime environment
##

FROM    alpine:latest AS rt-env

ARG     MTR_BIN=mtr-exporter-0.1.0.linux.amd64
RUN     apk add -U --no-cache mtr
COPY    --from=build-env /src/mtr-exporter/$MTR_BIN /usr/local/bin/mtr-exporter

EXPOSE  8080
ENTRYPOINT ["/usr/local/bin/mtr-exporter", "8.8.8.8"]
