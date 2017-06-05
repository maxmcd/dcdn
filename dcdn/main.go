package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

var upgrader = websocket.Upgrader{}
var conn *websocket.Conn

func main() {
	// launchBrowserAndDebugger("http://0.0.0.0:4041")
	go launchApplication()
	launchDriver()
}

func launchApplication() {
	r := mux.NewRouter()
	r.HandleFunc("/", userCodeHandler)
	r.HandleFunc("/ws", websocketHandler)
	port := "4041"
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("websocket comm listening on " + port)
	log.Fatal(srv.ListenAndServe())
}

func launchDriver() {
	handler := http.HandlerFunc(driverHandler)
	port := "4040"
	srv := &http.Server{
		Handler:      handler,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("driver listening on " + port)
	log.Fatal(srv.ListenAndServe())
}

type RequestInfo struct {
	Headers map[string]string `json:"headers"`
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Id      string            `json:"id"`
}

func driverHandler(w http.ResponseWriter, r *http.Request) {
	var requestInfo RequestInfo
	requestInfo.Url = r.URL.String()
	requestInfo.Method = r.Method

	requestInfo.Headers = make(map[string]string)

	for k, v := range r.Header {
		// assume there are only single header values
		// for now
		requestInfo.Headers[k] = v[0]
	}

	id := make([]byte, 20)
	_, err := rand.Read(id)
	if err != nil {
		fmt.Println("error:", err)
	}
	hexStrId := hex.EncodeToString(id)
	print(id)
	print(hexStrId)

	var total int
	for {
		chunk := make([]byte, 1500)
		n, err := r.Body.Read(chunk)
		total += n
		print(len(chunk))
		print(n)
		print("---")
		// TODO: limit chunk size by length
		toSend := append(id, chunk...)
		conn.WriteMessage(websocket.BinaryMessage, toSend)
		if err == io.EOF {
			break
		}
	}
	print(total)
	w.Write([]byte(`Hello World`))
}

func userCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
<!doctype html>
<html>
    <head>
        <title></title>
    </head>
    <body>
        <script>

            function getKey(array) { // buffer is an ArrayBuffer
                var array = array.slice(0,20)
                return Array.prototype.map.call(array, x => ('00' + x.toString(16)).slice(-2)).join('');
            }

            let ws = new WebSocket("ws://" + document.location.host + "/ws");
            ws.binaryType = "arraybuffer"
            ws.onclose = function (evt) {
                var item = document.createElement("div");
                item.innerHTML = "<b>Connection closed.</b>";
                appendLog(item);
            };
            ws.onmessage = function (evt) {
                console.log(evt)
                console.log(evt.data)
                var array = new Uint8Array(evt.data)
                console.log(array)
                var key = getKey(array)
                console.log(key)
                var string = new TextDecoder("utf-8").decode(array.slice(20));
                console.log(string)
                var dv = new DataView(evt.data, 0);
                console.log(dv.getInt8(0))
                // console.log(evt.data)
                // console.log(evt.data[0])
            };
            console.log("whatever yo")
            ws.onopen = function () {
                ws.send("hi")
            }
        </script>
    </body>
</html>
    `))
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	conn, err = upgrader.Upgrade(w, r, nil)
	print("websocket connected")
	if err != nil {
		log.Println("error upgrading:", err)
		return
	}
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		print(mt)
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func print(things ...interface{}) {
	fmt.Printf("%#v\n", things...)
}

func quickGet(url string) (resp *http.Response, err error) {
	timeout := time.Duration(100 * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
	}
	return client.Get(url)
}

func launchBrowser(location string) {
	// this might be helpful in the future to avoid the overhead
	// of the debugger, but for now headless chrome exits the
	// moment the dom is ready.
	cmd := exec.Command(
		"/usr/bin/google-chrome-unstable",
		"--headless",
		"--disable-gpu",
		"--dump-dom",
		location,
	)
	cmd.Run()
}

func launchBrowserAndDebugger(location string) (debugger *gcd.Gcd) {
	var wg sync.WaitGroup
	wg.Add(1)

	debugger = gcd.NewChromeDebugger()
	debugger.AddFlags([]string{
		"--disable-gpu", // required currently for headless
		"--headless",
		"--disable-web-security",             // disables CORS
		"--remote-debugging-address=0.0.0.0", // redundant?
		"--remote-debugging-port=9222",       // redundant?
		"--user-data-dir=/data",              // redundant?
	})
	debugger.StartProcess(
		"/usr/bin/google-chrome-unstable",
		"/data",
		"9222",
	)
	// defer debugger.ExitProcess()

	targets, err := debugger.GetTargets()
	if err != nil {
		log.Fatalf("error getting targets: %s\n", err)
	}
	if err != nil {
		log.Fatalf("error getting targets: %s\n", err)
	}
	target := targets[0]

	target.Subscribe("Page.loadEventFired", func(targ *gcd.ChromeTarget, v []byte) {
		wg.Done()
		// if you wanted to inspect the full response
		// data, you could do that here
	})

	console := target.Console
	runtime := target.Runtime

	target.Subscribe("Console.messageAdded", func(target *gcd.ChromeTarget, v []byte) {

		msg := &gcdapi.ConsoleMessageAddedEvent{}
		err := json.Unmarshal(v, msg)
		if err != nil {
			log.Fatalf("error unmarshalling event data: %v\n", err)
		}
		log.Printf("Console log: %s\n", msg.Params.Message)
	})

	target.Subscribe("Runtime.exceptionThrown", func(target *gcd.ChromeTarget, v []byte) {

		msg := &gcdapi.RuntimeExceptionThrownEvent{}
		err := json.Unmarshal(v, msg)
		if err != nil {
			log.Fatalf("error unmarshalling event data: %v\n", err)
		}
		log.Printf("Console log: %#v\n", msg.Params.ExceptionDetails)
	})

	runtime.Enable()
	console.Enable()

	// get the Page API and enable it
	if _, err := target.Page.Enable(); err != nil {
		log.Fatalf("error getting page: %s\n", err)
	}

	ret, err := target.Page.Navigate(location, "", "") // navigate
	if err != nil {
		log.Fatalf("Error navigating: %s\n", err)
	}
	log.Printf("ret: %#v\n", ret)

	wg.Wait()
	print("loaded")
	return debugger
}
