package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	ending := os.Getenv("ENDING")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// time.Sleep(1 * time.Second)
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, "Hello world! Dogs rule! %s\n", ending)
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
