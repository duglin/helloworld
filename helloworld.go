package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	text := "Hello World!"
	rev := os.Getenv("K_REVISION")
	if rev != "" {
		rev = rev[11:]
	}
	msg := fmt.Sprintf("%s: %s\n", rev, text)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Got request\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Fprint(w, msg)
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
