# Docker NAT Simulator

The setup is based on the simple Docker NAT Simulator in [the master branch](https://github.com/marten-seemann/docker-nat-simulator/tree/master).

It sets up a network with two go-libp2p nodes behind their respective NATs, starts a public relay node and achieves a direct connection between the two NAT-ed nodes by coordinating a hole-punching using DCUtR.

## Network Setup

<img title="Network Setup" src="network.png">

* The first host ("the client") (192.168.0.100) is behind a NAT that has the public IP 17.0.0.42.
* The second host ("the client") (10.0.0.100) is behind a NAT that has the public IP 17.0.10.42.
* Both routers apply an RTT of 50ms, thus the end-to-end RTT is 100ms.
* There's a (public) relay at 17.0.13.37.

## Running

```bash
docker compose build && docker compose up
```

Future Work:
* This example currently only works on TCP. This is most likely because the `iptables` command only applies to TCP, we'll need to figure out how to use `iptables` for a UDP NAT.
* The (go-libp2p) hole punching service requires nodes to discover their public-facing IP address. This is currently done by connecting to the relay. This is fine for the client, since the client requires a reservation with the relay anyway. For the server, it would be nicer if it didn't have to contact the relay before starting the hole punch attempt. This could be achieved by connecting to another public node.
