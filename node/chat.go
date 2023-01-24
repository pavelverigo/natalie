package node

import (
	"encoding/json"
	"errors"
	"log"
	"time"
)

type chatMsg struct {
	Text string
}

func (msg *chatMsg) Type() string {
	return "chat"
}

func (n *node) processChat(pkt *packet, addr string) error {
	msg := &chatMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	n.chatRecv = append(n.chatRecv, ChatData{
		Source: pkt.Source,
		Time:   time.Now(),
		Text:   msg.Text,
	})

	return nil
}

var errUnknown = errors.New("unknown destination")

func (n *node) sendChat(dest string, text string) error {
	addr := n.resolveRelayAddr(dest)
	if addr == "" {
		return errUnknown
	}
	pkt := n.newPacket(dest, &chatMsg{
		Text: text,
	})
	return n.sendPacket(addr, pkt)
}
