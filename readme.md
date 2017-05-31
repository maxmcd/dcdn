# DCDN

Web server requests handled within a chrome browser runtime.

A headless chrome web server runs in a docker container. A simple node server runs in a different container. The node server handles all public requests.

![](request-diagram.png?1 "Optional title")

The request flow is as follows:

1. A request is received by the Node server
2. The request information is gathered and sent to an active chrome browser session through a websocket connection. 
3. In the browser session, the request information is passed to a user-defined function. The user may define any external libraries and custom code they would like to use to process the request. As long as it runs in the browser.
4. The response body and headers are sent back to the node server in another websocket request.
5. The response is returned to the requesting client.

## Run

```bash
docker-compose up
```

Visit http://localhost:4040/request/ in your browser. You can dynamically alter the response body with the `body` url parameter: http://localhost:4040/request/?body=foo

Chrome debugging information is available at http://localhost:9222/



