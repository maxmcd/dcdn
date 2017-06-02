package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

var upgrader = websocket.Upgrader{}

func main() {
	// launchBrowser()
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

func driverHandler(w http.ResponseWriter, r *http.Request) {
	headers := map[string]string{}

	for k, v := range r.Header {
		// assume there's one header for now
		headers[k] = v[0]
	}
    print(headers)
	fmt.Println(r.URL.String())
	w.Write([]byte(`hi`))
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
            let ws = new WebSocket("ws://" + document.location.host + "/ws");
            ws.onclose = function (evt) {
                var item = document.createElement("div");
                item.innerHTML = "<b>Connection closed.</b>";
                appendLog(item);
            };
            ws.onmessage = function (evt) {
                console.log(evt.data)
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
	conn, err := upgrader.Upgrade(w, r, nil)
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

func launchBrowser() (debugger *gcd.Gcd) {
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

	target.Subscribe("Console.messageAdded", func(target *gcd.ChromeTarget, v []byte) {

		msg := &gcdapi.ConsoleMessageAddedEvent{}
		err := json.Unmarshal(v, msg)
		if err != nil {
			log.Fatalf("error unmarshalling event data: %v\n", err)
		}
		log.Printf("Console log: %s\n", msg.Params.Message.Text)
	})
	console.Enable()

	// get the Page API and enable it
	if _, err := target.Page.Enable(); err != nil {
		log.Fatalf("error getting page: %s\n", err)
	}

	ret, err := target.Page.Navigate("http://0.0.0.0:4041", "", "") // navigate
	if err != nil {
		log.Fatalf("Error navigating: %s\n", err)
	}
	log.Printf("ret: %#v\n", ret)

	wg.Wait()
	print("loaded")
	return debugger
}
