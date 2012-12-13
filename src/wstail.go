package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
	"os"
)

func indexHandler(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, `<html>
<head>
  <style>
    body
      { color: #1a2c37;
        font-family: 'Helvetica', sans-serif; font-size: 86%;
        padding: 2em; }
    #info
      { font-size: 120%;
        font-weight: bold; }
    #tail
      { border: 1px solid #ccc;
        height: 300px;
        padding: 0.5em;
        overflow: hidden;
        position: relative;
        overflow-y: scroll; }
  </style>
  <script type="text/javascript">
var header = "[wstail]"
var ws;
function init() {
  console.log("init");
  if (ws != null) {
    ws.close();
    ws = null;
  }
  path = "/tail";
  console.log("path:" + path);
  var div = document.getElementById("tail");
  div.innerText = div.innerText + header + "path:" + path + "\n";
  ws = new WebSocket("ws://localhost:23456" + path);
  ws.onopen = function () {
    div.innerText = div.innerText + header + "opened\n";
  };
  ws.onmessage = function (e) {
    console.log(e.data);
    var data = JSON.parse(e.data);
    if (data.key == "filename") {
      document.getElementById("info").innerText = data.value;
    } else if (data.key == "msg") {
      div.innerText = div.innerText + data.value;
    }
  };
  ws.onclose = function (e) {
    div.innerText = div.innerText + header + "closed\n";
  };
  console.log("init");
  div.innerText = div.innerText + header + "init\n";
};
  </script>
  <body onLoad="init();">
    <pre id="info"></pre>
    <pre id="tail"></pre>
</html>
`)
}

func startTail(file string, ch chan string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()
	var bufSizeMax int64 = 1024
	var bufSize int64
	if fileSize > bufSizeMax {
		bufSize = bufSizeMax
	} else {
		bufSize = fileSize
	}
	go func() {
		fmt.Println("tail start")
		ch <- file
		buf := make([]byte, bufSize)
		var offset int64 = 0
		{
			n, err := f.ReadAt(buf, offset+bufSize)
			if err != nil && err != io.EOF {
				panic("reader.ReadString(): " + err.Error())
			}
			line := string(buf[0:n])
			fmt.Printf("read[%v:%v]\n", n, line)
			ch <- line
		}
		for {
			n, err := f.Read(buf)
			if err == io.EOF && n == 0 {
				continue
			}
			if err != nil {
				panic("reader.ReadString(): " + err.Error())
			}
			line := string(buf[0:n])
			fmt.Printf("read[%v:%v]\n", n, line)
			ch <- line
		}
		fmt.Println("tail end")
	}()
	return nil
}

func makeWebsocketHandlerWithChannel(ch chan string, f func(chan string, *websocket.Conn)) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		f(ch, ws)
	}
}

type Data struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func websocketTailHandler(ch chan string, ws *websocket.Conn) {
	fmt.Printf("tailHandler %v\n", ws)
	// send first line as file name
	fileName := <-ch
	if err := websocket.JSON.Send(ws, Data{"filename", fileName}); err != nil {
		fmt.Println("ERR:websoket.Message.Send(): " + err.Error())
	}
	for {
		line := <-ch
		if err := websocket.JSON.Send(ws, Data{"msg", line}); err != nil {
			fmt.Println("ERR:websoket.Message.Send(): " + err.Error())
		}
		fmt.Printf("tailHandler write[%v]\n", line)
	}
	fmt.Println("tailHandler finished")
}

// for debug
func pseudoSubscriber(ch chan string) {
	for {
		line := <-ch
		fmt.Println("[sub]: " + line)
	}
}

func main() {
	ch := make(chan string)
	http.Handle("/tail", websocket.Handler(makeWebsocketHandlerWithChannel(ch, websocketTailHandler)))
	http.HandleFunc("/", indexHandler)

	if err := startTail(os.Args[1], ch); err != nil {
		panic(err)
	}

	fmt.Println("start wstail...")
	err := http.ListenAndServe(":23456", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	//pseudoSubscriber(ch)
}
