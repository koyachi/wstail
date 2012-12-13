package main

import (
	"code.google.com/p/go.net/websocket"
	"fmt"
	"io"
	"net/http"
	"os"
)

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
	http.Handle("/", http.FileServer(http.Dir("../view")))

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
