package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	endings := os.Getenv("ENDING")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		fmt.Fprintf(w, "Hello world!%s\n", ending)
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
