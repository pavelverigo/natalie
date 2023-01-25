package node

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackpal/gateway"
)

type traversalReqMsg struct {
	KnownAddr []string
	UseLocal  bool
}

func (msg *traversalReqMsg) Type() string {
	return "traversalreq"
}

type traversalRespMsg struct {
	KnownAddr []string
	UseLocal  bool
}

func (msg *traversalRespMsg) Type() string {
	return "traversalresp"
}

func (n *node) processTraversalReq(pkt *packet, addr string) error {
	msg := &traversalReqMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	if pkt.Source == n.name {
		log.Println("traversal req to myself WTF")
		return nil
	}

	go n.traversalLoop(pkt.Source, msg.KnownAddr)

	respPkt := n.newPacket(pkt.Source, &traversalRespMsg{
		KnownAddr: n.getPossibleAddresses(msg.UseLocal),
		UseLocal:  msg.UseLocal,
	})
	relayAddr := n.resolveRelayAddr(pkt.Source)
	if relayAddr == "" {
		log.Println("traversal req unable to contact who requested")
		return nil
	}
	return n.sendPacket(relayAddr, respPkt)
}

func (n *node) processTraversalResp(pkt *packet, addr string) error {
	msg := &traversalRespMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	if pkt.Source == n.name {
		log.Println("traversal resp to myself WTF")
		return nil
	}

	go n.traversalLoop(pkt.Source, msg.KnownAddr)

	return nil
}

func (n *node) traversalLoop(dest string, known []string) {
	i := 0
	for i < 3 {
		n.mu.Lock()
		_, ok := n.name2addr.GetByKey(dest)
		if ok {
			log.Println("already traversed, finish")
			n.mu.Unlock()
			return
		}

		for _, addr := range known {
			pkt := n.newPacket(dest, &handshakeReqMsg{
				ServerAddr: addr,
			})
			n.sendPacket(addr, pkt)
		}
		n.mu.Unlock()

		time.Sleep(time.Second * 2)
		i++
	}
}

func (n *node) getPossibleAddresses(local bool) []string {
	r := n.knownAddr.Keys()
	if local {
		ip, err := gateway.DiscoverInterface()
		if err != nil {
			log.Fatalln(err, "getPossible")
		}
		r = append(r, fmt.Sprintf("%s:%d", ip, n.port))
	}
	return r
}

func (n *node) traversalHandshake(dest string, local bool) error {
	pkt := n.newPacket(dest, &traversalReqMsg{
		KnownAddr: n.getPossibleAddresses(local),
		UseLocal:  local,
	})
	relayAddr := n.resolveRelayAddr(dest)
	if relayAddr == "" {
		log.Println("traversal req unable to contact who requested")
		return nil
	}
	return n.sendPacket(relayAddr, pkt)
}
