package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/wirepair/gcd"
)

func main() {
	launchBrowser()
}

func print(things ...interface{}) {
	fmt.Printf("%#v\n", things...)
}

func launchBrowser() {
	var wg sync.WaitGroup
	wg.Add(1)

	debugger := gcd.NewChromeDebugger()
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
	defer debugger.ExitProcess()
	targets, err := debugger.GetTargets()
	if err != nil {
		log.Fatalf("error getting targets: %s\n", err)
	}
	target := targets[0] // take the first one

	//subscribe to page load
	target.Subscribe("Page.loadEventFired", func(targ *gcd.ChromeTarget, v []byte) {
		wg.Done() // page loaded, we can exit now
		// if you wanted to inspect the full response data, you could do that here
	})
	// get the Page API and enable it
	if _, err := target.Page.Enable(); err != nil {
		log.Fatalf("error getting page: %s\n", err)
	}
	ret, err := target.Page.Navigate("http://www.apple.com", "", "") // navigate
	if err != nil {
		log.Fatalf("Error navigating: %s\n", err)
	}
	log.Printf("ret: %#v\n", ret)
	wg.Wait()
	print("loaded")
}
