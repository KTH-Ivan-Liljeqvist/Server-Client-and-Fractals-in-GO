/*
	Code skeleton written by: Stefan Nilsson
	Modified and completed by: Ivan Liljeqvist 2015-04-13

	****** MY TASK ******
	My task was to write a client that requests data from all of the servers at the same time.
	When one of the servers answered I should cancel the other requests.
	My task is also to make the request time out after a certain time.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	//all the URLs
	server := []string{
		"http://localhost:8080",
		"http://localhost:8081",
		"http://localhost:8082",
	}

	//loop forever and read the temperature
	for {
		//save time before the request
		before := time.Now()

		//request the temperature from all of the URLs at once
		res := MultiRead(server, time.Second)

		//save the time
		after := time.Now()

		//print the response and time to the console.
		fmt.Println("Response:", *res)
		fmt.Println("Time:", after.Sub(before))
		fmt.Println()

		//wait 500 ms before doing another request
		time.Sleep(500 * time.Millisecond)
	}

}

type Response struct {
	Body       string
	StatusCode int
}

// Get makes an HTTP Get request and returns an abbreviated response.
// Status code 200 means that the request was successful.
// The function returns &Response{"", 0} if the request fails
// and it blocks forever if the server doesn't respond.
func Get(url string) *Response {
	res, err := http.Get(url)
	if err != nil {
		return &Response{}
	}
	// res.Body != nil when err == nil
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	return &Response{string(body), res.StatusCode}
}

// I've found two insidious bugs in this function; both of them are unlikely
// to show up in testing. Please fix them right away and don't forget to
// write a doc comment this time.
/*
	Bug 1: If the request hits timeout this Read function will still run. It should not continue running.
		   I fixed this bug by making a buffered channel so that the function can exit if we hit a timeout.
		   By using a buffered channel we can place something in the channel and continue executing before someone takes
		   out something from the channel.

	Bug 2: The second bug is that the go routine inside Read wrote directly to the res channel when it recieved a response.
		   This means that the channel could be overridden by old requests that already timed out.
		   I fixed this by placing the assignment to the channel in the select statement. This way the channel
		   wont be written by old, timed out routines.
*/
func Read(url string, timeout time.Duration) (res *Response) {

	//make a buffered channel so that the routine can exit even if no one reads
	done := make(chan *Response, 1)

	//request - wait for Get(url) to place an answer in done channel
	go func() {
		done <- Get(url)
	}()

	select {

	//if we have an answer before timeout - go outside the select statement
	case res = <-done:

	//if we hit timeout before Get(url) places an answer in done - generate timeout response
	case <-time.After(timeout):
		res = &Response{"Gateway timeout\n", 504}

	}

	return

}

// MultiRead makes an HTTP Get request to each url and returns
// the response of the first server to answer with status code 200.
// If none of the servers answer before timeout, the response is
// 503 Service unavailable.
func MultiRead(urls []string, timeout time.Duration) (res *Response) {

	const REQUEST_SUCCEEDED_CODE int = 200

	answer_channel := make(chan *Response)

	//go through the urls
	for _, url := range urls {
		//start a routine for each one
		go func() {
			//get the response from this URL
			r := Read(url, timeout)
			//if the request succeeded - write to the answer channel
			if r.StatusCode == REQUEST_SUCCEEDED_CODE {
				answer_channel <- r
			}

		}()
	}

	//in this select statement we'll build the res that will be returned
	select {

	case res = <-answer_channel:
		//if we have a successfull answer, save it to res and continue outside this select statement
	case <-time.After(timeout):
		//if the request timed out - save a response with that message and code
		res = &Response{"Service unavailable\n", 503}
	}

	//return the res we built inside the select statement
	return
}
