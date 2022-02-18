package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var outputs []string
var screenMUX sync.Mutex
var flatList bool = false
var timeout = 10
var numClients = 10
var duration = 10

func Cursor(format string) {
	if !flatList {
		fmt.Printf(format)
	}
}

func Status(index int, str string) {
	screenMUX.Lock()
	defer screenMUX.Unlock()
	Cursor(fmt.Sprintf("\033[%d;0H", 1+index))
	fmt.Print(str)
	if flatList {
		fmt.Print("\n")
	}
}

func generateLoad(index int, url string, dur time.Duration, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	end := start.Add(dur)
	for time.Now().Before(end) {
		// resp, err := http.Get(url)

		resp, err := (&http.Client{
			Timeout: time.Duration(time.Duration(timeout) * time.Second),
		}).Get(url)

		output := ""
		pause := false
		if resp != nil && resp.StatusCode > 299 {
			output = resp.Status
			pause = true
		} else if err != nil {
			if output = err.Error(); strings.HasPrefix(output, "Get") {
				output = output[4+len(url):]
			}
			pause = true
		} else if resp.Body != nil {
			out, _ := ioutil.ReadAll(resp.Body)
			output = strings.TrimSpace(string(out))
		}
		if i := strings.IndexAny(output, "\n\r"); i >= 0 {
			output = strings.TrimSpace(output[:i-1])
		}

		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		// go Status(index, fmt.Sprintf("%02d: %-76.76s", 1+index, output))
		Status(index, fmt.Sprintf("%02d: %-76.76s", 1+index, output))

		// Don't pause the 1st client if there's more than one client
		if pause && !(numClients > 1 && index == 0) {
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	flag.BoolVar(&flatList, "l", false, "List of output instead of fancy")
	flag.IntVar(&timeout, "t", timeout, "HTTP client timeout")
	flag.Parse()

	if flag.NArg() != 3 {
		fmt.Fprintf(os.Stderr, "Usage: load num_of_requests duration URL\n")
		os.Exit(1)
	}
	numClients, _ = strconv.Atoi(flag.Arg(0))
	duration, _ = strconv.Atoi(flag.Arg(1))
	url := flag.Arg(2)
	outputs = make([]string, numClients)
	wg := sync.WaitGroup{}
	Cursor("\033[2J")
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go generateLoad(i, url, time.Duration(duration)*time.Second, &wg)
	}
	wg.Wait()
	Cursor(fmt.Sprintf("\033[%d;0H", 1+numClients))
}
