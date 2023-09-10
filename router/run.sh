#!/bin/sh

set -ex

$IPTABLES -t nat -A POSTROUTING -s $SUBNET_INTERNAL -o eth1 -j SNAT --to-source $ADDR_EXTERNAL

tc qdisc add dev eth1 root netem delay 50ms

tcpdump -i eth0 -n -w /dump.pcap &

ulogd
