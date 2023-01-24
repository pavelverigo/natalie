package node

import (
	"encoding/json"
	"log"
)

type handshakeReqMsg struct {
	ServerAddr string
}

func (msg *handshakeReqMsg) Type() string {
	return "handshakereq"
}

type handshakeRespMsg struct {
	ClientAddr string
}

func (msg *handshakeRespMsg) Type() string {
	return "handshakeresp"
}

func (n *node) processHandshakeReq(pkt *packet, addr string) error {
	msg := &handshakeReqMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	if pkt.Source == n.name {
		log.Println("handshake to myself WTF")
		return nil
	}

	n.name2addr.Set(pkt.Source, addr)
	n.knownAddr.Set(msg.ServerAddr)

	respPkt := n.newPacket(pkt.Source, &handshakeRespMsg{
		ClientAddr: addr,
	})
	n.sendPacket(addr, respPkt)

	return n.routingNeighborUpdate()
}

func (n *node) processHandshakeResp(pkt *packet, addr string) error {
	msg := &handshakeRespMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	if pkt.Source == n.name {
		log.Println("handshake to myself WTF")
		return nil
	}

	n.name2addr.Set(pkt.Source, addr)
	n.knownAddr.Set(msg.ClientAddr)

	n.routingNeighborUpdate()

	return nil
}

func (n *node) directHandshake(addr string) error {
	pkt := n.newPacket(directDestName, &handshakeReqMsg{
		ServerAddr: addr,
	})
	return n.sendPacket(addr, pkt)
}
