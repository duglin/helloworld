package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		msg := map[string]interface{}{}

		body, _ := ioutil.ReadAll(r.Body)
		err := json.Unmarshal(body, &msg)
		if err != nil {
			fmt.Printf("Error parsing: %s\n\n%s\n", err, string(body))
			return
		}

		body, _ = json.MarshalIndent(msg, "", "  ")
		fmt.Printf("JSON:\n%s\n\n", body)

		if msg["action"] != nil {
			fmt.Printf("Got issue event\n")
		} else if msg["hook"] != nil {
			fmt.Printf("Got hook event\n")
		} else if msg["pusher"] != nil {
			fmt.Printf("Got push event\n")
			out, err := exec.Command("/rebuild.sh").CombinedOutput()
			fmt.Printf("%s\n%s\n", out, err)
			fmt.Fprintf(w, "%s\n%s\n", out, err)
		} else {
			fmt.Printf("Unknown event\n")
		}
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
