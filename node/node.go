package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/pavelverigo/natalie/internal/bimap"
	"github.com/pavelverigo/natalie/internal/set"
)

const directDestName = "DESTNAME_DIRECT_HANDSHAKE"
const keepAliveInterval = 5 * time.Second
const routingStatusInterval = 10 * time.Second

func New(name string, port int, log *log.Logger) (Node, error) {
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprint(":", port))
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	return &node{
		log: log,
		wg:  sync.WaitGroup{},
		mu:  sync.Mutex{},

		conn: conn,
		name: name,

		name2addr: bimap.New[string, string](0),
		knownAddr: set.New[string](0),

		keepAliveTime: make(map[string]time.Time),

		routingTable: map[string]string{
			name: name,
		},
		nodesNeighborState: map[string]neighborState{
			name: {
				Seq:       0,
				Neighbors: make([]string, 0),
			},
		},

		chatRecv: make([]ChatData, 0),
	}, nil
}

type node struct {
	log *log.Logger
	wg  sync.WaitGroup
	mu  sync.Mutex

	conn *net.UDPConn
	name string

	name2addr *bimap.BiMap[string, string]
	knownAddr *set.Set[string]

	keepAliveTime map[string]time.Time // previosly recorded keep alive message from neighbor

	routingTable       map[string]string
	nodesNeighborState map[string]neighborState

	chatRecv []ChatData
}

type neighborState struct {
	Seq       uint
	Neighbors []string
}

type ChatData struct {
	Source string    `json:"src"`
	Time   time.Time `json:"time"`
	Text   string    `json:"text"`
}

type Node interface {
	Start() error
	Stop() error

	LocalAddr() string
	Neighbors() map[string]string
	KnownAddr() []string
	RoutingTable() map[string]string
	Chat() []ChatData

	SendChat(dest, text string) error

	DirectHandshake(addr string)
	TraversalHandshake(name string)
}

func (n *node) Start() error {
	go n.readLoop()
	go n.keepAliveLoop()
	go n.routingLoop()

	return nil
}

func (n *node) Stop() error {
	return nil
}

func (n *node) LocalAddr() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.conn.LocalAddr().String()
}

func (n *node) Neighbors() map[string]string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.name2addr.CopyM1()
}

func (n *node) KnownAddr() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.knownAddr.Keys()
}

func (n *node) RoutingTable() map[string]string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return copyMap(n.routingTable)
}

func (n *node) Chat() []ChatData {
	n.mu.Lock()
	defer n.mu.Unlock()
	return copySlice(n.chatRecv)
}

func (n *node) DirectHandshake(addr string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.directHandshake(addr)
}

func (n *node) SendChat(dest, text string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.sendChat(dest, text)
}

func (n *node) TraversalHandshake(name string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.traversalHandshake(name)
}

func (n *node) removeNeighbor(name string) {
	delete(n.keepAliveTime, name)
	n.name2addr.DeleteByKey(name)
	n.routingNeighborUpdate()
}

func (n *node) readLoop() {
	buf := make([]byte, 8192)
	for {
		sz, netaddr, err := n.conn.ReadFrom(buf)
		if err != nil {
			log.Fatalln(err)
		}
		addr := netaddr.String()

		pkt := &packet{}
		err = json.Unmarshal(buf[:sz], pkt)
		if err != nil {
			log.Fatalln(err)
		}

		n.mu.Lock()

		neighbor, ok := n.name2addr.GetByValue(addr)
		onlyLocal := (pkt.Type == "routingupdate" || pkt.Type == "keepalive" || pkt.Type == "routingstatus")
		if !ok && onlyLocal {
			n.log.Println("logic failed, local pkt", pkt)
			n.mu.Unlock()
			continue
		}

		isHandshake := (pkt.Type == "handshakereq" || pkt.Type == "handshakeresp" || pkt.Type == "traversalreq" || pkt.Type == "traversalresp")
		if ok && isHandshake && pkt.Destination == n.name && pkt.Destination == directDestName {
			n.log.Println("handshake or traversal from neighbor", pkt)
			n.mu.Unlock()
			continue
		}

		if ok {
			n.keepAliveTime[neighbor] = time.Now() // update
		}

		if pkt.Destination != n.name && pkt.Destination != directDestName {
			relayAddr := n.resolveRelayAddr(pkt.Destination)
			if relayAddr == "" {
				err = errors.New("unknown addr to relay to")
			} else {
				err = n.sendPacket(relayAddr, pkt)
			}
		} else {
			switch pkt.Type {
			case "handshakereq":
				err = n.processHandshakeReq(pkt, addr)
			case "handshakeresp":
				err = n.processHandshakeResp(pkt, addr)
			case "keepalive":
				err = n.processKeepAlive(pkt, addr)
			case "routingstatus":
				err = n.processRoutingStatus(pkt, addr)
			case "routingupdate":
				err = n.processRoutingUpdate(pkt, addr)
			case "chat":
				err = n.processChat(pkt, addr)
			case "traversalreq":
				err = n.processTraversalReq(pkt, addr)
			case "traversalresp":
				err = n.processTraversalResp(pkt, addr)
			default:
				err = errors.New(fmt.Sprint("unknown packet type:", pkt.Type))
				log.Fatalln(err)
			}
		}

		if err != nil {
			n.log.Println(err)
		}

		n.mu.Unlock()
	}
}

type message interface {
	Type() string
}

func (n *node) newPacket(dest string, msg message) *packet {
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}
	return &packet{
		Id:          randomID(16),
		Source:      n.name,
		Destination: dest,
		Type:        msg.Type(),
		Payload:     payload,
	}
}

type packet struct {
	Id          string
	Source      string
	Destination string
	Type        string
	Payload     json.RawMessage
}

func (n *node) sendPacket(addr string, pkt *packet) error {
	netaddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return err
	}

	data, err := json.Marshal(pkt)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = n.conn.WriteTo(data, netaddr)
	return err
}
