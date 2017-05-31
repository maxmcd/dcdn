const app = require("express")();
const http = require("http").Server(app);
const io = require('socket.io')(http);
const bodyParser = require('body-parser')
const uuidV4 = require('uuid/v4');
const port = 4040;

let requests = {}

app.use(bodyParser.text())

app.get("/", (req, res) => {
    res.sendFile(__dirname + "/index.html");
});

app.all("/request/", (req, res, next) => {
    let key = uuidV4();
    requests[key] = res
    io.emit('request', {
        headers: req.headers,
        key: key,
        query: req.query,
    })
});

http.listen(port, () => {
    console.log(`listening on *:${port}`);
});

io.on('connection', function(socket){
  socket.on('response', function(msg){
    requests[msg.key].send(msg.body)
  });
});