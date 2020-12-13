package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func hash(s string) string {
	bv := []byte(s)
	hasher := sha1.New()
	hasher.Write(bv)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func handler(w http.ResponseWriter, r *http.Request) {
	str := randStringBytes(750)
	hashStr := hash(str)
	fmt.Fprintf(w, "Hi there, %s! your hash: %s", r.URL.Path[1:], hashStr)
}

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("server is listen on port 8888")
	log.Fatal(http.ListenAndServe(":8888", nil))
}
