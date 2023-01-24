package node

import (
	"encoding/json"
	"log"
	"time"
)

type routingStatusMsg struct {
	SeqState map[string]uint
}

func (msg *routingStatusMsg) Type() string {
	return "routingstatus"
}

func (n *node) processRoutingStatus(pkt *packet, addr string) error {
	msg := &routingStatusMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	// log.Println(n.name, "status", msg.SeqState)

	neighbor := pkt.Source

	nodes := make(map[string]neighborState) // which should be send after
	newState := false
	for name, seq := range msg.SeqState {
		state, ok := n.nodesNeighborState[name]
		if !ok || state.Seq < seq {
			newState = true
			continue
		}
		if state.Seq > seq {
			nodes[name] = n.nodesNeighborState[name]
		}
	}
	for name, state := range n.nodesNeighborState {
		_, ok := msg.SeqState[name]
		if !ok {
			nodes[name] = state
		}
	}

	if newState {
		// neighbor will respond with update
		n.sendRoutingStatus(neighbor) // TODO: handle error
	}

	if len(nodes) > 0 {
		addr, _ := n.name2addr.GetByKey(neighbor)
		n.sendPacket(addr, n.newPacket(neighbor, &routingUpdateMsg{ // TODO: handle error
			Nodes: nodes,
		}))
	}

	return nil
}

type routingUpdateMsg struct {
	Nodes map[string]neighborState
}

func (msg *routingUpdateMsg) Type() string {
	return "routingupdate"
}

func (n *node) processRoutingUpdate(pkt *packet, addr string) error {
	msg := &routingUpdateMsg{}
	err := json.Unmarshal(pkt.Payload, msg)
	if err != nil {
		log.Fatalln(err)
	}

	// log.Println(n.name, "update", msg.Nodes)

	recvNew := false
	for name, state1 := range msg.Nodes {
		state2, ok := n.nodesNeighborState[name]
		if !ok || state2.Seq < state1.Seq {
			n.nodesNeighborState[name] = state1
			recvNew = true
		}
	}

	if recvNew {
		n.recalculateRoutingTable()
		n.broadcastRoutingStatusExcept(pkt.Source) // TODO: handle error
	}

	return nil
}

// BFS, calculate routing table from nodes neighbor state
func (n *node) recalculateRoutingTable() {
	layer := n.name2addr.Keys()
	bfs := make(map[string]string)
	// init state
	bfs[n.name] = n.name
	for _, name := range layer {
		bfs[name] = name
	}

	for len(layer) > 0 {
		newLayer := make([]string, 0)
		for _, from := range layer {
			state, ok := n.nodesNeighborState[from]
			if !ok {
				continue
			}
			for _, to := range state.Neighbors {
				_, ok := bfs[to]
				if !ok {
					newLayer = append(newLayer, to)
					bfs[to] = bfs[from]
				}
			}
		}
		layer = newLayer
	}

	n.routingTable = bfs
}

func (n *node) resolveRelayAddr(dest string) string {
	relay, ok := n.routingTable[dest]
	if !ok {
		return ""
	}
	addr, _ := n.name2addr.GetByKey(relay)
	return addr
}

func (n *node) sendRoutingStatus(dest string) error {
	seqState := make(map[string]uint, len(n.nodesNeighborState))
	for name, state := range n.nodesNeighborState {
		seqState[name] = state.Seq
	}

	addr, _ := n.name2addr.GetByKey(dest)
	return n.sendPacket(addr, n.newPacket(dest, &routingStatusMsg{ // TODO: handle error
		SeqState: seqState,
	}))
}

// send to all neighbors, except one
func (n *node) broadcastRoutingStatusExcept(ignore string) error {
	for _, name := range n.name2addr.Keys() {
		if name == ignore {
			continue
		}
		n.sendRoutingStatus(name) // TODO: handle error
	}
	return nil
}

// neighbor was deleted or added,
// note: calculated state may not be reconstructed from neighborstate, until new neighbor send his state
func (n *node) routingNeighborUpdate() error {
	prev := n.nodesNeighborState[n.name]
	n.nodesNeighborState[n.name] = neighborState{
		Seq:       prev.Seq + 1,
		Neighbors: n.name2addr.Keys(),
	}

	n.recalculateRoutingTable()

	return n.broadcastRoutingStatusExcept("") // TODO: really better to send everyone update!
}

func (n *node) routingLoop() { // may help when packets lost, or reordered (handshake response recieved, after routing distance msg)
	for {
		time.Sleep(routingStatusInterval)

		n.mu.Lock()

		n.broadcastRoutingStatusExcept("") // send to everyone

		n.mu.Unlock()
	}
}
