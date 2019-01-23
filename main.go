package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Got request\n")
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, "Hello World! 12503\n")
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
