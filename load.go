package main

import (
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

func Status(index int, str string) {
	screenMUX.Lock()
	defer screenMUX.Unlock()
	fmt.Printf("\033[%d;0H%s", 1+index, str)
}

func generateLoad(index int, url string, dur time.Duration, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	start := time.Now()
	end := start.Add(dur)
	for time.Now().Before(end) {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error(%d): %s\n", index, err)
			return
		}
		if resp.Body != nil {
			out, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			// outputs[index] = strings.TrimSpace(string(out))
			output := strings.TrimSpace(string(out))
			// go Status(index, fmt.Sprintf("%02d: %-76.76s", 1+index, output))
			Status(index, fmt.Sprintf("%02d: %-76.76s", 1+index, output))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error(%d): %s\n", index, err)
			return
		}
	}
}

func displayOutput() {
	for {
		fmt.Printf("\033[H")
		for index, output := range outputs {
			fmt.Printf("%02d: %-76s\n", 1+index, output)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	num, _ := strconv.Atoi(os.Args[1])
	dur, _ := strconv.Atoi(os.Args[2])
	url := os.Args[3]
	outputs = make([]string, num)
	wg := sync.WaitGroup{}
	fmt.Printf("\033[2J")
	for i := 0; i < num; i++ {
		wg.Add(1)
		go generateLoad(i, url, time.Duration(dur)*time.Second, &wg)
	}
	// go displayOutput()
	wg.Wait()
	fmt.Printf("\033[%d;0H", 1+10)
}
