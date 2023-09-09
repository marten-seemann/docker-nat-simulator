FROM golang:1.21 AS builder

COPY node /node
WORKDIR /node
RUN go mod download
RUN go build -o node main.go


FROM ubuntu:jammy

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y iproute2
# install some tools useful for debugging
RUN apt-get install -y tcpdump ncat iputils-ping

COPY --from=builder /node/node node

COPY *.sh /scripts/
RUN chmod +x /scripts/*.sh

# should be overwritten by docker compose
ENTRYPOINT ["exit 137"]
