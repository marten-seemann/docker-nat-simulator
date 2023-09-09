FROM ubuntu:jammy

ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get -y install iproute2 iptables ulogd2 
RUN apt-get install -y tcpdump iputils-ping

COPY *.sh /scripts/
RUN chmod +x /scripts/*.sh
COPY ulogd.conf /etc/ulogd.conf

ENTRYPOINT ["./scripts/run.sh"]
