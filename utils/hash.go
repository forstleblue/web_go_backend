package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/OneOfOne/xxhash" //this is the fastest hash algorithm
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

//Hash returns a uint64 hash
func Hash(data []byte) uint64 {
	return xxhash.Checksum64(data)
}

//HashStr returns a string hash
func HashStr(data []byte) string {
	return fmt.Sprintf("%x", Hash(data))[2:]
}

//RandomString generates a random string
func RandomString(strlen int) string {
	return string(RandomStringBytes(strlen))
}

//RandomStringBytes generates a random string as []bytes
func RandomStringBytes(strlen int) []byte {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return result
}
