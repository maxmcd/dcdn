package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/lib/pq"
)

func postRequest(t *testing.T) (resp *http.Response) {
	// emptyBytes := make([]byte, 10000)
	// _ = emptyBytes
	message := []byte(`
		This is a request body. Let's making it
		longer than 1500 bytes. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
		Padding. Padding. Padding. Padding. Padding. Padding.
	`)

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
	t.Skip()

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
}

func TestSingleRequest(t *testing.T) {
	// t.Skip()
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
		t.Errorf("not a hello")
	}
	// time.Sleep(time.Minute * 1)
}

func TestDB(t *testing.T) {
	db := connectToDB()
	err := dropAppTable(db, appTableName)
	if err != nil {
		t.Error(err)
	}
	name, err := createAppTable(db, appTableName)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Routine != "NewRelationAlreadyExistsError" {
			t.Error(err)
		}
	}
	if name != appTableName {
		t.Errorf("incorrect table name")
	}
	err = writeKeyValue(db, appTableName, "foo", "bar")
	if err != nil {
		t.Error(err)
	}
	value, err := getKeyValue(db, appTableName, "foo")
	if value != "bar" {
		t.Errorf("incorrect value for key")
	}
	err = writeKeyValue(db, appTableName, "foo", "baz")
	if err != nil {
		t.Error(err)
	}
	value, err = getKeyValue(db, appTableName, "foo")
	if value != "baz" {
		t.Errorf("incorrect value for key")
	}
}

func BenchmarkDB(b *testing.B) {
	db := connectToDB()
	err := dropAppTable(db, appTableName)
	if err != nil {
		b.Error(err)
	}
	_, err = createAppTable(db, appTableName)
	if err != nil {
		pqErr := err.(*pq.Error)
		if pqErr.Routine != "NewRelationAlreadyExistsError" {
			b.Error(err)
		}
	}

	for i := 0; i < b.N; i++ {
		err := writeKeyValue(db, appTableName, "foo", "baz")
		if err != nil {
			b.Error(err)
		}
		_, err = getKeyValue(db, appTableName, "foo")
		if err != nil {
			b.Error(err)
		}
	}
}
