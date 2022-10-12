package main

import (
	"flag"
	"net"
	"strings"
)

var portFlag = flag.Int("port", 3000, "initial port")

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

	conn, _ := net.ListenUDP("udp4", &net.UDPAddr{Port: *portFlag})
	table := make(map[string]net.Addr)
	var buf [1024]byte

	for {
		n, addr, _ := conn.ReadFrom(buf[:])
		name := string(buf[:n])

		table[name] = addr

		resp := encodeTable(table)
		conn.WriteTo([]byte(resp), addr)
	}
}
