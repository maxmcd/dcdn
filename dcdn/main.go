package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

type websocketMessage struct {
	messageType int
	data        []byte
}

type RequestInfo struct {
	Headers map[string]string `json:"headers"`
	Url     string            `json:"url"`
	Method  string            `json:"method"`
	Key     string            `json:"key"`
	HasBody bool              `json:"hasBody"`
}

type ResponseInfo struct {
	Headers map[string]string `json:"headers"`
	Status  int               `json:"status"`
	Body    string            `json:"body"`
}
type Response struct {
	Key  string       `json:"key"`
	Info ResponseInfo `json:"info"`
}

var upgrader = websocket.Upgrader{}
var websocketChannel chan websocketMessage
var requests = map[string](chan ResponseInfo){}
var lock = sync.RWMutex{}

func init() {
	websocketChannel = make(chan websocketMessage, 100)
}

func main() {

	// launchBrowserAndDebugger("http://0.0.0.0:4041")
	go launchApplication()
	launchDriver()
}

func writeRequest(key string) (reqChannel chan ResponseInfo) {
	reqChannel = make(chan ResponseInfo, 1)
	lock.Lock()
	defer lock.Unlock()
	requests[key] = reqChannel
	return reqChannel
}

func getRequestChannel(key string) (reqChannel chan ResponseInfo) {
	lock.RLock()
	defer lock.RUnlock()
	return requests[key]
}

func deleteRequest(key string) {
	lock.Lock()
	defer lock.Unlock()
	delete(requests, key)
}

func fullyLaunchServers() (srvA *http.Server, srvD *http.Server) {
	now := time.Now()
	srvA = launchApplication()
	srvD = launchDriver()

	for {
		// TODO: add endpoint for testing driver?
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
	elapsed := time.Since(now)
	fmt.Printf("Launching servers took %f seconds.\n", elapsed.Seconds())

	return
}

func launchApplication() (srv *http.Server) {
	r := mux.NewRouter()
	r.HandleFunc("/", userCodeHandler)
	r.HandleFunc("/ws", websocketHandler)
	port := "4041"
	srv = &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("websocket comm listening on " + port)
	go func() {
		log.Println(srv.ListenAndServe())
	}()
	return srv
}

func launchDriver() (srv *http.Server) {
	handler := http.HandlerFunc(driverHandler)
	port := "4040"
	srv = &http.Server{
		Handler:      handler,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	fmt.Println("driver listening on " + port)
	go func() {
		log.Println(srv.ListenAndServe())
	}()
	return srv
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

	if requestInfo.Headers["Content-Length"] != "" {
		requestInfo.HasBody = true
	}

	print(requestInfo.Headers)
	id := make([]byte, 20)
	_, err := rand.Read(id)
	if err != nil {
		fmt.Println("error:", err)
	}
	hexStrId := hex.EncodeToString(id)

	requestInfo.Key = hexStrId
	messageBytes, err := json.Marshal(requestInfo)
	if err != nil {
		fmt.Println("error:", err)
	}
	message := websocketMessage{
		messageType: websocket.TextMessage,
		data:        messageBytes,
	}
	websocketChannel <- message

	if requestInfo.HasBody {
		go func() {
			var total int
			for {
				// decide what the optimal size is here
				chunk := make([]byte, 1500)
				n, err := r.Body.Read(chunk)
				total += n

				if n > 0 {
					toSend := append(id, chunk[:n]...)
					message := websocketMessage{
						messageType: websocket.BinaryMessage,
						data:        toSend,
					}
					websocketChannel <- message
				}
				if err == io.EOF {
					// TODO:
					// this seems very poor
					// lets figure out a better way to do
					// this
					message := websocketMessage{
						messageType: websocket.BinaryMessage,
						data:        append(id, []byte("EOF")...),
					}
					websocketChannel <- message
					break
				}
			}
			print(total)
		}()
	}

	reqChannel := writeRequest(hexStrId)
	defer deleteRequest(hexStrId)

	resp := <-reqChannel
	for k, v := range resp.Headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(resp.Status)
	w.Write([]byte(resp.Body))
}

func userCodeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	print("websocket connected")
	if err != nil {
		log.Println("error upgrading:", err)
		return
	}
	defer conn.Close()

	go func() {
		for {
			mt, message, err := conn.ReadMessage()
			_, _ = mt, message
			fmt.Println(string(message))
			var resp Response
			err = json.Unmarshal(message, &resp)
			if err != nil {
				fmt.Println(err)
				continue
			}
			reqChannel := getRequestChannel(resp.Key)
			reqChannel <- resp.Info
			if err != nil {
				log.Println("read:", err)
				break
			}
		}
	}()

	for {
		toSend := <-websocketChannel
		err = conn.WriteMessage(toSend.messageType, toSend.data)
		if err != nil {
			log.Println("write:", err)
			continue
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

func launchBrowser(location string) (cmd *exec.Cmd, err error) {
	// unclear if this is better
	// both open 4 chrome processes
	// if you run the process headless without a debugger port it opens 3.
	// so tab, browser view, and something else
	// debugger is one chrome process
	// hmm

	// does not occupy debugging port, gcd does
	cmd = exec.Command(
		"/usr/bin/google-chrome-unstable",
		"--headless",
		"--disable-gpu",
		"--disable-web-security",
		"--remote-debugging-address=0.0.0.0",
		"--remote-debugging-port=9222",
		"--user-data-dir=/data",
		location,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	return cmd, err
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
		// user-data-dir will tie chrome processes together, consider making this
		// dyanmic
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
		log.Printf("Console log: %s\n", msg.Params.Message.Text)
	})

	target.Subscribe("Runtime.exceptionThrown", func(target *gcd.ChromeTarget, v []byte) {

		msg := &gcdapi.RuntimeExceptionThrownEvent{}
		err := json.Unmarshal(v, msg)
		if err != nil {
			log.Fatalf("error unmarshalling event data: %v\n", err)
		}
		log.Printf(
			"Error: %s %s\n",
			msg.Params.ExceptionDetails.Text,
			msg.Params.ExceptionDetails.Exception.Description,
		)
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
