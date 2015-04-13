/*
	Author: Stefan Nilsson
	Modified by Ivan Liljeqvist 2015-04-13

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

// FIXME
// I've found two insidious bugs in this function; both of them are unlikely
// to show up in testing. Please fix them right away and don't forget to
// write a doc comment this time.
func Read(url string, timeout time.Duration) (res *Response) {
	done := make(chan bool)
	go func() {
		res = Get(url)
		done <- true
	}()
	select {
	case <-done:
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
	return // TODO
}
