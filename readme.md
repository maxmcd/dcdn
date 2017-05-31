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

Might be best to use the debugging protocal for communication with the chrome instances. Need to look into how that works. Maybe there's a performance improvement over regular websockets. 

### Resources

 - https://chromedevtools.github.io/devtools-protocol/
 - https://github.com/novnc/websockify
 - https://idea.popcount.org/2017-03-28-sandboxing-lanscape/
 - https://github.com/golang/go/issues/18892
 - https://github.com/google/lovefield  
 - https://github.com/yukinying/chrome-headless-browser-docker
 - https://github.com/websockets/ws