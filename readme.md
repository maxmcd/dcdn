# DCDN


## Overview

Web server requests handled within a chrome browser runtime.

A headless chrome browser is launched. Two web servers are launched. One to serve an html page to the browser, another to proxy external requests.

External requests are proxied from the web server to the browser environment through a websocket connection. Client code is executed in the browser environment and the response is returned via the websocket connection.

![](request-diagram.png?1 "")

The request flow is as follows:

1. A request is received by the proxy server
2. The request information is gathered and sent to an active chrome browser session through a websocket connection. 
3. In the browser session, the request information is passed to a user-defined function. The user may define any external libraries and custom code they would like to use to process the request.
4. The response body and headers are sent back to the node server in another websocket request.
5. The response is returned to the requesting client.

## Run

```bash
docker-compose up
```

## Dev

### Performance In The Browser

It seems that at the moment the single threaded nature of the browser is the main blocker. Need to explore using multiple tabs. 

Memory consumption might be a bigger issue, each chrome instance launches 3 or 4 instance of chrome. Using multiple tabs would help alleviate this issue. Although the security of hosting code from different users in the same browser session is potentially problematic. 

If the browser is the real target and the sandbox is not extracted from the chromium project it might be worth exploring WebRTC.

### Structure

**Request Driver:** Handles the inbound request and passes it off to the websocket. In the beginning we'll just use something very simple here, but could be expanded to support udp, tcp, and other kinds of network requests.

**Host/Browser Code:** Whatever the user wants. In the beginning might make sense to provide some helpful patterns and allow for quick and simple request-making. But, at the end of the day, the user should own the browser environment.

**Hosting Structure:** Very open-ended, could go with the Dynamic CDN route, could allow dynamic scaling, could hide it all away, could provide full control. 

**Datastore:** Sharing a single browser session allows for memory sharing. Might not be the best to advertise this feature. A globally available key>value store seems like a good option to provide to users. Beyond that a globally available SQL database like cockroach might be a good call. Hosting a DB instance for each install is problematic. A global dictionary store or KV store allows for simpler horizontal scaling between users.

### Suggested Architecture

Writing code for this application involved adopting an entire new stack. I don't think it's enough that these environments are lightweight. Lambda wins because you can re-use code, not just because of scale ease. With that in mind it might be worth it to just focus on the use-case of a dynamic CDN.

With a dynamic CDN we would want:

- Quick deploys to many nodes around the world
- Globally available datastore for some data needs
- Low latency is very important

Container orchestration could be handled in each region by Kubernetes. The datastore could be K/V at first. 

Cloud datastore seems appealing, but they require chosing a single geographic region for data. Voldemort is great, but needs a large initial instance size. Might just use Cockroach because it's "cool"™ and would potentially scale to providing SQL to users. Could create a new table for every application.

A deploy would:

 - Accept user code and write it to some persistent datastore
 - Allocate an endpoint to that code
 - Application running in all regions would consume the new endpoint and details
 - A node would be spun up for the application in all regions.

### Resources

 - https://chromedevtools.github.io/devtools-protocol/
 - https://github.com/novnc/websockify
 - https://idea.popcount.org/2017-03-28-sandboxing-lanscape/
 - https://github.com/golang/go/issues/18892
 - https://github.com/google/lovefield  
 - https://github.com/yukinying/chrome-headless-browser-docker
 - https://github.com/websockets/ws
 - https://github.com/wirepair/gcd
 - http://peter.sh/experiments/chromium-command-line-switches/?date=2015-05-14#disable-web-security
 - https://cs.chromium.org/chromium/src/headless/app/headless_shell_switches.cc