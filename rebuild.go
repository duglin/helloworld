package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var echo = (os.Getenv("ECHO") != "")

func Print(w http.ResponseWriter, format string, args ...interface{}) {
	re := regexp.MustCompile(`token=[^ ]* `)
	str := re.ReplaceAllString(fmt.Sprintf(format, args...), "token=TOKEN  ")

	fmt.Printf(str)
	if echo {
		fmt.Fprintf(w, str)
	}
}

func Run(w http.ResponseWriter, cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		Print(w, "Failed running: %s %s\n%s\n",
			cmd, strings.Join(args, " "), err)
	}
	return string(out), err
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		var out string
		msg := map[string]interface{}{}

		body, _ := ioutil.ReadAll(r.Body)
		if err = json.Unmarshal(body, &msg); err != nil {
			Print(w, "Error parsing: %s\n\n%s\n", err, string(body))
			return
		}

		body, _ = json.MarshalIndent(msg, "", "  ")
		Print(w, "HEADERS:\n%v\nBODY:\n%s\n\n", r.Header, body)

		if msg["action"] != nil {
			Print(w, "Got issue event\n")
		} else if msg["hook"] != nil {
			Print(w, "Got hook event\n")
		} else if msg["pusher"] != nil {
			Print(w, "Got push event\n")

			token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
			if err != nil {
				Print(w, "Can't load API token:%s\n", err)
				return
			}

			// Which Knative Service are we rebuilding?
			ksvc := os.Getenv("KSVC")
			if ksvc == "" {
				ksvc = "helloworld"
			}
			Print(w, "ksvc: %s\n", ksvc)

			// Get the YAML of the KnService so we can edit it
			args := []string{"kubectl", "--token=" + string(token),
				"get", "ksvc/" + ksvc, "-oyaml"}
			if out, err = Run(w, args[0], args[1:]...); err != nil {
				Print(w, "Error getting ksvc %q: %s\n%s\n", ksvc, out, err)
				return
			}

			// Modify the:  trigger: ...   line
			lines := strings.Split(out, "\n")
			foundIt := false
			for i, line := range lines {
				if strings.Contains(line, "   trigger:") {
					j := strings.Index(line, ":")
					lines[i] = line[:j+2] + `"` + time.Now().String() + `"`
					foundIt = true
					break
				}
			}

			if !foundIt {
				Print(w, "Can't find trigger:\n%s\n", string(out))
				return
			}

			// Save edited files to disk
			tmpFile, _ := ioutil.TempFile("", "")
			defer os.Remove(tmpFile.Name())
			buf := []byte(strings.Join(lines, "\n"))
			if err = ioutil.WriteFile(tmpFile.Name(), buf, 0700); err != nil {
				Print(w, "Error saving yaml: %s\n", err)
				return
			}

			// Apple the edits
			Print(w, "  Applying edits...\n")
			args = []string{"kubectl", "--token=" + string(token),
				"apply", "-f", tmpFile.Name()}
			if _, err = Run(w, args[0], args[1:]...); err != nil {
				Print(w, "%s\n%s\n%s\n", strings.Join(args, " "), out, err)
				return
			}

			Print(w, "  Done\n")
		} else {
			Print(w, "Unknown event: %s\n", string(body))
			if !echo {
				fmt.Fprintf(w, "Unknown event: %s\n", string(body))
			}
		}
	})

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
