package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Result struct {
	duration   int
	statusCode int
	failed     bool
}

type Report struct {
	max         int
	min         int
	mean        float64
	statusCodes map[int]int
	n_failed    int
}

func readPostData(filepath string) *bytes.Buffer {

	plan, err := ioutil.ReadFile(filepath)

	if err != nil {
		fmt.Println("Could not read data file")
		fmt.Println(err)
	}

	var data interface{}
	err = json.Unmarshal(plan, &data)

	if err != nil {
		fmt.Println("Could not read json")
	}

	return bytes.NewBuffer(plan)
}

func readHeadersFile(filepath string) map[string]string {

	plan, err := ioutil.ReadFile(filepath)

	if err != nil {
		fmt.Println("Could not read data file")
		fmt.Println(err)
	}

	var headersData = map[string]string{}
	err = json.Unmarshal(plan, &headersData)

	return headersData

}

func call(c chan Result,
	endpoint string,
	dataBuffer *bytes.Buffer,
	headers map[string]string) {

	start := time.Now()

	req, err := http.NewRequest("POST", endpoint, dataBuffer)
	if err != nil {
		fmt.Println(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Response Error")
		fmt.Println(err)
		fmt.Println("Failed Request")
		c <- Result{failed: true}
		return
	}

	fmt.Println(resp)
	fmt.Println(resp.StatusCode)

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	s := buf.String()

	fmt.Println(s)

	t := time.Now()
	timeTaken := int(t.Sub(start)) / 1000000 // Gives milliseconds

	c <- Result{duration: timeTaken, statusCode: resp.StatusCode}

}

func analytics(c chan Result, n int) Report {

	max := 0
	min := 1000000000
	sum := 0
	count := 0
	n_failed := 0

	status := map[int]int{}

	for i := 0; i < n; i++ {

		ele := <-c

		if ele.failed {
			n_failed += 1
			continue
		}

		if ele.duration > max {
			max = ele.duration
		}

		if ele.duration < min {
			min = ele.duration
		}

		sum += ele.duration
		count += 1

		status[ele.statusCode] += 1

	}

	mean := float64(sum) / float64(count)

	return Report{
		max:         max,
		min:         min,
		mean:        mean,
		statusCodes: status,
		n_failed:    n_failed,
	}

}

func main() {

	ngoPtr := flag.Int("ngo", 10, "number of goroutines")
	nPerGoPtr := flag.Int("npergo", 25, "number of calls per goroutine")
	endpointPtr := flag.String("endpoint", "http://localhost:8000", "endpoint we are targeting")
	dataFilePtr := flag.String("datafile", "my_data.json", "json file that contains data for post request")
	headersFilePtr := flag.String("headersfile", "", "json file that contains headers data")

	flag.Parse()

	n_goroutines := *ngoPtr
	calls_per_goroutine := *nPerGoPtr
	endpoint := *endpointPtr
	datafile := *dataFilePtr
	headersfile := *headersFilePtr

	dataBuffer := readPostData(datafile)

	headers := map[string]string{}
	if headersfile != "" {
		headers = readHeadersFile(headersfile)
	}

	c := make(chan Result)

	for client := 0; client < n_goroutines; client++ {
		go func() {
			for i := 0; i < calls_per_goroutine; i++ {
				// dataBuffer is being consumed, we need to copy it
				dataBufferCopy := &bytes.Buffer{}
				*dataBufferCopy = *dataBuffer
				call(c, endpoint, dataBufferCopy, headers)
			}
		}()
	}

	report := analytics(c, n_goroutines*calls_per_goroutine)

	fmt.Println("REPORT")
	fmt.Printf("%+v\n", report)
}
