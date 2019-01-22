package main

import (
	"fmt"
	// "io/ioutil"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// body, _ := ioutil.ReadAll(r.Body)
		// fmt.Printf("URL: %s\nBody: %v\n", r.URL, string(body))
		time.Sleep(1 * time.Second)
		fmt.Fprintf(w, "Hello world!\n")
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
