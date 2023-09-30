FROM alpine:3.18.4

WORKDIR /opt/infping
RUN apk add --no-cache fping
COPY infping /opt/infping/infping
COPY config.toml /opt/infping/config.toml
ENTRYPOINT ["/opt/infping/infping"]
