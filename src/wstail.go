package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	//"fmt"
	//"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	viewDir = flag.String("view-dir", "", "path to view directory")
)

/*
var templates *template.Template

func loadTemplate() error {
	var err error
	t := template.New("wstail")
	templates, err = t.ParseGlob(fmt.Sprintf("%s/*.html", *viewDir))
	if err != nil {
		return err
	}
	return nil
}
*/
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
		log.Println("tail start")
		ch <- file
		buf := make([]byte, bufSize)
		var offset int64 = 0
		{
			n, err := f.ReadAt(buf, offset+bufSize)
			if err != nil && err != io.EOF {
				panic("reader.ReadString(): " + err.Error())
			}
			line := string(buf[0:n])
			log.Printf("read[%v:%v]\n", n, line)
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
			log.Printf("read[%v:%v]\n", n, line)
			ch <- line
		}
		log.Println("tail end")
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
	log.Printf("tailHandler %v\n", ws)
	// send first line as file name
	fileName := <-ch
	if err := websocket.JSON.Send(ws, Data{"filename", fileName}); err != nil {
		log.Println("ERR:websoket.Message.Send(): " + err.Error())
	}
	for {
		line := <-ch
		if err := websocket.JSON.Send(ws, Data{"msg", line}); err != nil {
			log.Println("ERR:websoket.Message.Send(): " + err.Error())
		}
		log.Printf("tailHandler write[%v]\n", line)
	}
	log.Println("tailHandler finished")
}

// for debug
func pseudoSubscriber(ch chan string) {
	for {
		line := <-ch
		log.Println("[sub]: " + line)
	}
}

func main() {
	flag.Parse()
	if *viewDir == "" {
		for _, defaultPath := range []string{"../view", "view", "/usr/local/share/wstail/view"} {
			if info, err := os.Stat(defaultPath); err == nil && info.IsDir() {
				*viewDir = defaultPath
				break
			}
		}
	}
	if *viewDir == "" {
		log.Fatalf("view dir not found")
	}
	//loadTemplate()

	ch := make(chan string)
	http.Handle("/tail", websocket.Handler(makeWebsocketHandlerWithChannel(ch, websocketTailHandler)))
	http.Handle("/", http.FileServer(http.Dir(*viewDir)))

	file := flag.Args()[0]
	if err := startTail(file, ch); err != nil {
		panic(err)
	}

	log.Println("start wstail...")
	err := http.ListenAndServe(":23456", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
	//pseudoSubscriber(ch)
}
