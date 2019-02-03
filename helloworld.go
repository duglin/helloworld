package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	text := "Hello World!"

	rev := os.Getenv("K_REVISION")
	if i := strings.LastIndex(rev, "-"); i > 0 {
		rev = rev[i+1:]
	}
	msg := fmt.Sprintf("%s: %s\n", rev, text)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Got request\n")
		time.Sleep(1000 * time.Millisecond)
		fmt.Fprint(w, msg)
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
