package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, "HHello world! \n")
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
