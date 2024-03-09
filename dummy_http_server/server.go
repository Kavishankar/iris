package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	err := http.ListenAndServe(":3001", http.HandlerFunc(countHandler))
	if err != nil {
		log.Panic(err)
	}
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(30 * time.Second)
	w.Write([]byte("OK"))
}
