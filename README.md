# Docker NAT Simulator

The logic is loosely based on https://github.com/zzJinux/docker-nat-simulate, but it replaces all bash scripts used for setup with a Docker compose setup.

The setup only uses iptables to achieve NAT-ing.

## Network Setup

<img title="Network Setup" src="network.png">

* The clients (192.168.0.0/16) are in a network that's assumed to be separate from the rest of the network by a NAT.
* The server (10.0.0.100) is on the other side of the NAT.
* The router (192.168.0.42 and 10.0.0.42, respectively) acts as a NAT between these two networks.

## Running

```bash
docker compose build && docker compose up
```

## Validating the Setup

### Using ping

Open a shell on one of the clients:
```bash
docker exec -it client /bin/bash
```

Then ping the server:
```bash
ping server
```

This works since the NAT is translating addresses from the internal network to the outside world.

Conversely, trying to ping the client from the server does not work, as we'd expect.

Open a shell on one of the server:
```bash
docker exec -it server /bin/bash
```

And try to ping the client:
```bash
ping 192.168.0.100
```
