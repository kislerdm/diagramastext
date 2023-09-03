package main

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/kislerdm/diagramastext/server/core/ciam"
)

func main() {
	_, key, _ := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().Unix())))

	o, err := ciam.MarshalKey(key)
	if err != nil {
		log.Fatalf("cannot marshal generated key: %v", err)
	}

	fmt.Print(string(o))
}
