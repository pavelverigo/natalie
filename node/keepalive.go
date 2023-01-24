package node

import (
	"encoding/json"
	"log"
	"time"
)

type keepAliveMsg struct{}

func (msg *keepAliveMsg) Type() string {
	return "keepalive"
}

func (n *node) processKeepAlive(pkt *packet, addr string) error {
	msg := &keepAliveMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	// nothing to do, time update done in recv loop

	return nil
}

func (n *node) keepAliveLoop() {
	for {
		time.Sleep(keepAliveInterval)

		n.mu.Lock()

		now := time.Now()
		for _, name := range n.name2addr.Keys() {
			// check if neighbor, should be deleted
			prev := n.keepAliveTime[name]
			if now.Sub(prev) > 3*keepAliveInterval {
				n.removeNeighbor(name)
				continue
			}

			// send keep alive to neighbor
			addr, _ := n.name2addr.GetByKey(name)
			n.sendPacket(addr, n.newPacket(name, &keepAliveMsg{}))
		}

		n.mu.Unlock()
	}
}
