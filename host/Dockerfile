FROM ubuntu:jammy

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y iproute2 
# install some tools useful for debugging
RUN apt-get install -y tcpdump ncat iputils-ping

COPY *.sh /scripts/
RUN chmod +x /scripts/*.sh

ENTRYPOINT ["./scripts/run.sh"]
