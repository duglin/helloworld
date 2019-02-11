package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Run(cmd string, args ...string) string {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed running: %s %s\n%s\n",
			cmd, strings.Join(args, " "), err)
		os.Exit(1)
	}
	return string(out)
}

func main() {
	ready := false
	apikey := os.Getenv("IC_KEY")
	cluster := os.Getenv("CLUSTER")

	fmt.Printf("Cluster: %s\nAPIKey: %s\n", cluster, apikey[:3])

	// Put this in the background so we don't slow down the
	// creation of the Knative rebuild service
	go func() {
		Run("bx", "login", "--apikey", apikey, "-r", "us-south")
		Run("bx", "config", "--check-version", "false")
		export := Run("bx", "ks", "cluster-config", "-s", "--export", cluster)
		export = strings.SplitN(export, "=", 2)[1]
		export = strings.TrimSpace(export)

		fmt.Printf("export KUBCONFIG=%s\n", export)
		os.Setenv("KUBECONFIG", export)
		ready = true
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Wait for our IBM Cloud setup to finish
		for !ready {
			time.Sleep(200 * time.Millisecond)
		}

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
