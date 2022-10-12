package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var nameFlag = flag.String("name", "", "client name")
var serverFlag = flag.String("server", "localhost:3000", "server address")

func decodeTable(data string) map[string]net.Addr {
	pairs := strings.Split(data, "|")
	m := make(map[string]net.Addr)
	for _, p := range pairs {
		split := strings.Split(p, ",")
		name := split[0]
		addr := split[1]
		uaddr, _ := net.ResolveUDPAddr("udp4", addr)
		m[name] = uaddr
	}
	return m
}

func encodeTable(table map[string]net.Addr) string {
	var sb strings.Builder
	count := 0
	mapLen := len(table)
	for name, addr := range table {
		sb.WriteString(name)
		sb.WriteRune(',')
		sb.WriteString(addr.String())
		count++
		if count < mapLen {
			sb.WriteRune('|')
		}
	}
	return sb.String()
}

func main() {
	flag.Parse()

	if *nameFlag == "" {
		panic("name flag is not present")
	}

	serverAddr, _ := net.ResolveUDPAddr("udp4", *serverFlag)

	conn, _ := net.ListenUDP("udp4", &net.UDPAddr{})

	go func() {
		for {
			conn.WriteTo([]byte(*nameFlag), serverAddr)
			time.Sleep(time.Second * 5)
		}
	}()

	var mu sync.Mutex
	table := make(map[string]net.Addr)
	recv := make(map[string]net.Addr)

	go func() {
		for {
			mu.Lock()
			log.Println("from server:", encodeTable(table), "recv:", encodeTable(recv))
			mu.Unlock()
			time.Sleep(time.Second * 20)
		}
	}()

	go func() {
		var buf [1024]byte
		for {
			n, addr, _ := conn.ReadFrom(buf[:])
			data := string(buf[:n])

			mu.Lock()
			if strings.Contains(data, ",") {
				table = decodeTable(data)
			} else {
				// data == name
				recv[data] = addr
			}
			mu.Unlock()
		}
	}()

	go func() {
		for {
			for name, addr := range table {
				if name == *nameFlag {
					continue
				}
				conn.WriteTo([]byte(*nameFlag), addr)
			}

			time.Sleep(time.Second * 5)
		}
	}()

	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
