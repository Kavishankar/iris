package main

import (
	"log"
	"net/http"
	"time"
)

// main function spins up a proxy(?) server at port:3001.
func main() {
	err := http.ListenAndServe(":3001", http.HandlerFunc(countHandler))
	if err != nil {
		log.Panic(err)
	}
}

// countHandler waits for 30sec to send an 'OK' response.
func countHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(30 * time.Second)
	w.Write([]byte("OK"))
}
