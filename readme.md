# DCDN

Web server requests handled within a chrome browser runtime.

A headless chrome web server runs in a docker container. A simple NodeJS server runs in a different container. The Node server handles all public requests.

![](request-diagram.png?1 "Optional title")

The request flow is as follows:

1. A request is received by the Node server
2. The request information is gathered and sent to an active chrome browser session through a websocket connection. 
3. In the browser session, the request information is passed to a user-defined function. The user may define any external libraries and custom code they would like to use to process the request.
4. The response body and headers are sent back to the node server in another websocket request.
5. The response is returned to the requesting client.

## Run

```bash
docker-compose up
```

Visit http://localhost:4040/request/ in your browser. You can dynamically alter the response body with the `body` url parameter: http://localhost:4040/request/?body=foo

Chrome debugging information is available at http://localhost:9222/


## Dev

### Notes

#### Performance In The Browser

Seems useful to have multiple requests feed to the same browser session. Concurrent requests lead to an expected slow-down when being fed to the browser. Requests that take 3ms without concurrent traffic take up to 100ms when 50 concurrent requests are fed to the same browser instance. I guess even WASM wouldn't help here as websockets seem to be the bottleneck. 

Might be best to use the debugging protocol for communication with the chrome instances. Need to look into how that works. Maybe there's a performance improvement over regular websockets. 

A single browser instance is helpful for simplicity and for allowing a user to share memory. So we can either keep the memory sharing model and have an upper performance threshold, or provide access to a simple shared datastore between instances and recommend no memory sharing. Likely a decision to be made later on. 

#### Websocket performance

Re-implementing `socket.io` with `ws` led to an incredibly small performance improvment, so we're sticking with `socket.io` for now.

#### Structure

**Request Driver:** Handles the inbound request and passes it off to the websocket. In the beginning we'll just use something very simple here, but could be expanded to support udp, tcp, and other kinds of network requests.

**Host/Browser Code:** Whatever the user wants. In the beginning might make sense to provide some helpful patterns and allow for quick and simple request-making. But, at the end of the day, the user should own the browser environment.

**Hosting Structure:** Very open-ended, could go with the Dynamic CDN route, could allow dynamic scaling, could hide it all away, could provide full control. 

**Datastore:** Sharing a single browser session allows for memory sharing. Should likely just provide a stubbed localstorage, or other tools to easy handle data persistence between sessions. Would allow for more flexibility with scaling and managing requests. 



### Resources

 - https://chromedevtools.github.io/devtools-protocol/
 - https://github.com/novnc/websockify
 - https://idea.popcount.org/2017-03-28-sandboxing-lanscape/
 - https://github.com/golang/go/issues/18892
 - https://github.com/google/lovefield  
 - https://github.com/yukinying/chrome-headless-browser-docker
 - https://github.com/websockets/ws