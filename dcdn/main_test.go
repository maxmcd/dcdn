package main

import (
	"bytes"
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
			continue
		}
		if resp.StatusCode == 200 {
			break
		} else {
			time.Sleep(time.Millisecond * 100)
		}
	}

	debugger := launchBrowserAndDebugger("http://0.0.0.0:4041")
	defer debugger.ExitProcess()

	// emptyBytes := make([]byte, 10000)
	// _ = emptyBytes
	message := []byte(`adfasdfajsdlfkjasl;dfkj as;ldkf jas;ldfjk asdf`)
	resp, err := http.Post(
		"http://0.0.0.0:4040/",
		"text/plain",
		bytes.NewBuffer(message),
	)
	// resp, err := http.Get(
	// 	"http://0.0.0.0:4040/",
	// )
	if err != nil {
		t.Error(err)
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	print(string(bytes))
	if string(bytes) != "Hello World" {
		t.Fatalf("not a hello")
	}
	time.Sleep(time.Second * 300)
}
