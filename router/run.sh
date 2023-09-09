#!/bin/sh

set -ex

$IPTABLES -t nat -A POSTROUTING -s $SUBNET_INTERNAL -o eth1 -j SNAT --to-source $ADDR_EXTERNAL

ulogd