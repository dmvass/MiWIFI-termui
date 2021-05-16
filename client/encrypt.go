package client

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"
)

const key = "a2ffa5c9be07488bbb04a3a47d3c5f6a"

func generateNonce(macAddress string) string {
	t := time.Now().Unix()
	r := rand.Float64() * 10e3
	// type_MAC_timestamp_rand
	return fmt.Sprintf("0_%s_%d_%d", macAddress, t, int(r))
}

func hashPassword(password, nonce string) string {
	h := sha1.New()

	h.Write([]byte(password))
	h.Write([]byte(key))
	hash := hex.EncodeToString(h.Sum(nil))

	h.Reset()

	h.Write([]byte(nonce))
	h.Write([]byte(hash))
	hash = hex.EncodeToString(h.Sum(nil))

	return hash
}
