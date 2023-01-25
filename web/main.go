package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"regexp"
	"sort"
	"sync"

	"github.com/pavelverigo/natalie/node"
)

//go:embed static/*
var static embed.FS

type server struct {
	fsys fs.FS // TODO: custom fileserver with 404 page.

	mu    sync.Mutex
	nodes map[string]node.Node
}

func main() {
	fsys, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatalln(err)
	}

	s := server{
		fsys: fsys,

		mu:    sync.Mutex{},
		nodes: make(map[string]node.Node),
	}

	http.HandleFunc("/api/nodes/", s.handleNodes)
	http.Handle("/", http.FileServer(http.FS(fsys)))

	log.Println("Listening on http://localhost:80")

	log.Fatalln(http.ListenAndServe(":80", nil))
}

var re = regexp.MustCompile("^[A-Za-z0-9]+$")

func (s *server) handleNodes(w http.ResponseWriter, r *http.Request) {
	const prefixLen = len("/api/nodes/")

	// log.Println("handle node", r.URL.Path, r.Method)

	name := r.URL.Path[prefixLen:]
	if name == "" {
		switch r.Method {
		case "GET":
			s.handleNodesList(w, r)
		case "POST":
			s.handleNodesAdd(w, r)
		default:
			http.NotFound(w, r)
		}
		return
	}

	if re.MatchString(name) {
		switch r.Method {
		case "GET":
			s.handleNodeData(w, r, name)
		case "POST":
			s.handleNodeOp(w, r, name)
		default:
			http.NotFound(w, r)
		}
		return
	}

	http.NotFound(w, r)
}

func (s *server) handleNodesList(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	names := make([]string, len(s.nodes))
	i := 0
	for name := range s.nodes {
		names[i] = name
		i++
	}
	s.mu.Unlock()

	sort.Strings(names)

	data, err := json.Marshal(names)
	if err != nil {
		panic(err)
	}
	w.Write(data)
}

func (s *server) handleNodesAdd(w http.ResponseWriter, r *http.Request) {
	type addData struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}

	var add addData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&add)
	if err != nil {
		panic(err)
	}

	node, err := node.New(add.Name, add.Port, log.Default())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = node.Start()
	if err != nil {
		panic(err)
	}

	s.mu.Lock()
	s.nodes[add.Name] = node
	s.mu.Unlock()
}

func (s *server) handleNodeData(w http.ResponseWriter, r *http.Request, name string) {
	s.mu.Lock()
	n, ok := s.nodes[name]
	s.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	type nodeData struct {
		LocalAddr    string            `json:"local"`
		Neighbors    map[string]string `json:"neigh"`
		Addresses    []string          `json:"addr"`
		RoutingTable map[string]string `json:"routing"`
		Chat         []node.ChatData   `json:"chat"`
	}

	var data nodeData

	data.LocalAddr = n.LocalAddr()
	data.Neighbors = n.Neighbors()
	data.Addresses = n.KnownAddr()
	data.RoutingTable = n.RoutingTable()
	data.Chat = n.Chat()

	json, err := json.Marshal(&data)
	if err != nil {
		log.Fatalln(err)
	}
	w.Write(json)
}

func (s *server) handleNodeOp(w http.ResponseWriter, r *http.Request, name string) {
	s.mu.Lock()
	node, ok := s.nodes[name]
	s.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}

	type opData struct {
		Op   string          `json:"op"`
		Data json.RawMessage `json:"data"`
	}

	var op opData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&op)
	if err != nil {
		panic(err)
	}

	switch op.Op {
	case "nat":
		type natData struct {
			Dest  string `json:"dest"`
			Local bool   `json:"local"`
		}
		var nat natData
		err := json.Unmarshal(op.Data, &nat)
		if err != nil {
			panic(err)
		}
		node.TraversalHandshake(nat.Dest, nat.Local)
	case "direct":
		type directData struct {
			Addr string `json:"addr"`
		}
		var direct directData
		err := json.Unmarshal(op.Data, &direct)
		if err != nil {
			panic(err)
		}
		node.DirectHandshake(direct.Addr)
	case "chat":
		type chatData struct {
			Dest string `json:"dest"`
			Text string `json:"text"`
		}
		var chat chatData
		err := json.Unmarshal(op.Data, &chat)
		if err != nil {
			panic(err)
		}
		node.SendChat(chat.Dest, chat.Text)
	default:
		log.Fatalln("unknown", op.Op)
	}
}
