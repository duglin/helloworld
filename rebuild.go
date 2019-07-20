package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
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

func Run(cmd string, args ...string) (string, error) {
	out, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Failed running: %s %s\n%s\n",
			cmd, strings.Join(args, " "), err)
	}
	return string(out), err
}

func AnnotateService(svc string, key string, value string) error {
	// Which Knative Service are we rebuilding?
	if svc == "" {
		svc = os.Getenv("KSVC")
		if svc == "" {
			svc = "helloworld"
		}
	}

	stamp := time.Now().String()

	_, err := Kubectl("patch", "ksvc/"+svc, "--type=merge",
		"-p",
		fmt.Sprintf(`{"spec":{"template":{"metadata":{"name":"","annotations":{"newimage":"%s"}}}}}`, stamp))

	return err
}

func Kubectl(args ...string) (string, error) {
	buf, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return "", fmt.Errorf("Can't load API token:%s\n", err)
	}

	token := "--token=" + string(buf)
	args = append([]string{token}, args...)

	out, err := Run("kubectl", args...)
	if err != nil {
		return "", fmt.Errorf("%s\n%s\n%s\n", strings.Join(args, " "), out, err)
	}
	return out, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var err error
		msg := map[string]interface{}{}

		body, _ := ioutil.ReadAll(r.Body)
		if err = json.Unmarshal(body, &msg); err != nil {
			Print(w, "Error parsing: %s\n\n%s\n", err, string(body))
			return
		}

		body, _ = json.MarshalIndent(msg, "", "  ")
		Print(w, "HEADERS:\n%v\nBODY:\n%s\n\n", r.Header, body)

		envs := os.Environ()
		sort.Strings(envs)
		Print(w, "ENV:\n%s\n\n", strings.Join(envs, "\n"))

		ceType := r.Header.Get("ce-type")

		if ceType == "newimage" {
			Print(w, "Got newimage event\n")
			err := AnnotateService("", "newimage", time.Now().String())
			if err != nil {
				Print(w, "Error annotating: %s\n", err)
				return
			}
			Print(w, "  Done\n")
		} else if msg["action"] != nil {
			Print(w, "Got issue event\n")
		} else if msg["hook"] != nil {
			Print(w, "Got hook event\n")
		} else if msg["pusher"] != nil {
			Print(w, "Got push event\n")

			Print(w, "Rebuild: %s\n", os.Getenv("REBUILDURL"))

			out, err := Run("/kapply", "-c", "/task.yaml")
			fmt.Printf("'kapply -v -c /task.yaml' output:\n")
			if err != nil {
				Print(w, "Error: %s\n", err)
				return
			}
			fmt.Printf("%s\n", out)
			fmt.Printf("----\n")
			Print(w, "  Done\n")
		} else {
			Print(w, "Unknown event: %s\n", string(body))
			if !echo {
				fmt.Fprintf(w, "Unknown event: %s\n", string(body))
			}
		}
	})

	serviceURL, err := Kubectl("get", "ksvc", os.Getenv("K_SERVICE"),
		"-ocustom-columns=url:status.url", "--no-headers")
	if err != nil {
		fmt.Printf("Error getting service URL: %s\n", err)
		return
	}
	fmt.Printf("ServiceURL: %s\n", serviceURL)
	os.Setenv("REBUILDURL", serviceURL)

	fmt.Print("Listening on port 8080\n")
	http.ListenAndServe(":8080", nil)
}
