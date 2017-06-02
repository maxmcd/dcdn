package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestSingleRequest(t *testing.T) {
	go launchApplication()
	go launchDriver()

	for {
		resp, err := quickGet("http://localhost:4041/")
		if err != nil {
			fmt.Println(err)
		}
		if resp.StatusCode == 200 {
			break
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}

	debugger := launchBrowser()
	defer debugger.ExitProcess()

	resp, err := http.Get("http://localhost:4040/")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(time.Second * 20)
	bytes, err := ioutil.ReadAll(resp.Body)
	print(string(bytes))
	if string(bytes) != "Hello World" {
		t.Fatalf("not a hello")
	}
}
