package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
)

var relayPeerID, clientPeerID peer.ID

func init() {
	identify.ActivationThresh = 1

	var err error
	relayPeerID, err = peer.Decode("12D3KooWFy8BjPcNCDW5uEPyzrqWqjycHgv1FWC7Kx8jQ1Jbunp5")
	if err != nil {
		log.Fatal(err)
	}

	clientPeerID, err = peer.Decode("12D3KooWM82bDYYgzgXaayHDdVciFe3bGvJ69qHnbSztNUJ933VQ")
	if err != nil {
		log.Fatal(err)
	}
}

type keyReader struct {
	secret byte
}

func (r *keyReader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = r.secret
	}
	return len(b), nil
}

func main() {
	var err error
	switch os.Args[1] {
	case "relay":
		err = runRelay()
	case "client": // the node behind the NAT
		time.Sleep(time.Second)
		err = runClient(os.Args[2])
	case "server":
		time.Sleep(2 * time.Second)
		err = runServer(os.Args[2])
	default:
		log.Fatalf("unknown role: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func runClient(relayIP string) error {
	priv, _, err := crypto.GenerateEd25519Key(&keyReader{'c'})
	if err != nil {
		return err
	}
	relayAddrInfo := peer.AddrInfo{
		ID:    relayPeerID,
		Addrs: []ma.Multiaddr{ma.StringCast(fmt.Sprintf("/ip4/%s/tcp/1234", relayIP))},
	}
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/10000"),
		libp2p.ForceReachabilityPrivate(),
		libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{relayAddrInfo}, autorelay.WithBootDelay(0)),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return err
	}
	if h.ID() != clientPeerID {
		return fmt.Errorf("got unexpected peer ID: %s", h.ID())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.Connect(ctx, relayAddrInfo); err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}
	log.Println("Connected to relay:", h.Network().ConnsToPeer(relayPeerID))
	time.Sleep(2 * time.Second)
	log.Println("Client listening on", addP2PComponents(h.Addrs(), h.ID()))
	select {}
}

func runServer(relayIP string) error {
	priv, _, err := crypto.GenerateEd25519Key(&keyReader{'s'})
	if err != nil {
		return err
	}
	clientAddrInfo := peer.AddrInfo{
		ID:    clientPeerID,
		Addrs: []ma.Multiaddr{ma.StringCast(fmt.Sprintf("/ip4/%s/tcp/1234/p2p/%s/p2p-circuit", relayIP, relayPeerID))},
	}
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/20000"),
		libp2p.ForceReachabilityPrivate(),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
	)
	if err != nil {
		return err
	}
	// TODO: use a different node for that
	// connect to the relay first in order to learn our public address
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Connect(ctx, peer.AddrInfo{
		ID:    relayPeerID,
		Addrs: []ma.Multiaddr{ma.StringCast(fmt.Sprintf("/ip4/%s/tcp/1234", relayIP))},
	}); err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}

	t := time.NewTicker(100 * time.Millisecond)
publicAddrLoop:
	for {
		select {
		case <-t.C:
			if hasPublicAddr(h.Addrs()) {
				break publicAddrLoop
			}
		case <-ctx.Done():
			return errors.New("failed to discover public address")
		}
	}
	for _, c := range h.Network().ConnsToPeer(relayPeerID) {
		c.Close()
	}
	time.Sleep(time.Second) // give the hole punch service some time to discover that we now have a public address

	// now connect to the client via the relay
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Connect(ctx, clientAddrInfo); err != nil {
		return fmt.Errorf("failed to connect to relay: %w", err)
	}
	for range time.NewTicker(12 * time.Second).C {
		log.Println("Connected to client:", h.Network().ConnsToPeer(clientPeerID))
	}
	return nil
}

// runRelay runs the relay
// The peer ID will always be 12D3KooWFy8BjPcNCDW5uEPyzrqWqjycHgv1FWC7Kx8jQ1Jbunp5.
func runRelay() error {
	priv, _, err := crypto.GenerateEd25519Key(&keyReader{'r'})
	if err != nil {
		return err
	}
	h, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/1234"),
		libp2p.ForceReachabilityPublic(),
		libp2p.EnableRelayService(),
		libp2p.DisableRelay(),
	)
	if err != nil {
		return err
	}
	if h.ID() != relayPeerID {
		return fmt.Errorf("got unexpected peer ID: %s", h.ID())
	}

	log.Println("Relay listening on", addP2PComponents(h.Addrs(), h.ID()))
	select {}
}

func addP2PComponents(addrs []ma.Multiaddr, id peer.ID) []ma.Multiaddr {
	p2ppart, err := ma.NewComponent("p2p", id.String())
	if err != nil {
		log.Fatal(err)
	}
	result := make([]ma.Multiaddr, 0, len(addrs))
	for _, addr := range addrs {
		result = append(result, addr.Encapsulate(p2ppart))
	}
	return result
}

func hasPublicAddr(addrs []ma.Multiaddr) bool {
	for _, addr := range addrs {
		if manet.IsPublicAddr(addr) {
			return true
		}
	}
	return false
}
