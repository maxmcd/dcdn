package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func postRequest(t *testing.T) (resp *http.Response) {
	// emptyBytes := make([]byte, 10000)
	// _ = emptyBytes
	message := []byte(`adfasdfajsdlfkjasl;dfkj as;ldkf jas;ldfjk asdf`)

	now := time.Now()

	var err error
	resp, err = http.Post(
		"http://0.0.0.0:4040/",
		"text/plain",
		bytes.NewBuffer(message),
	)
	elapsed := time.Since(now)
	print(elapsed.Seconds())
	// resp, err := http.Get(
	//  "http://0.0.0.0:4040/",
	// )
	if err != nil {
		t.Error(err)
	}
	return resp
}

func TestSingleWithDebuggerRequest(t *testing.T) {

	srvA, srvD := fullyLaunchServers()
	defer srvA.Shutdown(nil)
	defer srvD.Shutdown(nil)

	debugger := launchBrowserAndDebugger("http://0.0.0.0:4041")
	defer debugger.ExitProcess()

	resp := postRequest(t)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	print(string(bytes))
	if string(bytes) != "Hello World" {
		t.Fatalf("not a hello")
	}
	time.Sleep(time.Minute * 1)
}

func TestSingleRequest(t *testing.T) {

	srvA, srvD := fullyLaunchServers()
	defer srvA.Shutdown(nil)
	defer srvD.Shutdown(nil)

	cmd, err := launchBrowser("http://0.0.0.0:4041")
	if err != nil {
		t.Error(err)
	}
	time.Sleep(1 * time.Second)
	defer cmd.Process.Kill()

	resp := postRequest(t)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	print(string(bytes))
	if string(bytes) != "Hello World" {
		t.Fatalf("not a hello")
	}
	time.Sleep(time.Minute * 1)
}
