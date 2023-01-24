package node

import (
	"crypto/rand"
	"encoding/base64"
	"log"
)

func randomID(size int) string {
	data := make([]byte, size)
	_, err := rand.Read(data)
	if err != nil {
		log.Fatalln(err)
	}
	return base64.RawStdEncoding.EncodeToString(data)
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	r := make(map[K]V, len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

func copySlice[T any](s []T) []T {
	r := make([]T, len(s))
	copy(r, s)
	return r
}
