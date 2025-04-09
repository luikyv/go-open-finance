package main

import (
	"encoding/hex"
	"fmt"
	"math/rand"
)

func generateUUID(randSource rand.Source) string {
	r := rand.New(randSource)
	b := make([]byte, 16)
	r.Read(b)
	return hex.EncodeToString(b[:4]) + "-" + hex.EncodeToString(b[4:6]) + "-" +
		hex.EncodeToString(b[6:8]) + "-" + hex.EncodeToString(b[8:10]) + "-" +
		hex.EncodeToString(b[10:])
}

func main() {
	randSouce := rand.NewSource(42)
	fmt.Println(generateUUID(randSouce))
	fmt.Println(generateUUID(randSouce))
	fmt.Println(generateUUID(randSouce))
}
