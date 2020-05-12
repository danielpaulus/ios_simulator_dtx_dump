package proxy

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

type FileAndSocketWriter struct {
	File string
	Dst  io.Writer
}

func NewFileAndSocketWriter(path string, dst net.Conn) *FileAndSocketWriter {
	return &FileAndSocketWriter{path, dst}
}

func (f *FileAndSocketWriter) Write(data []byte) (n int, err error) {
	file, err := os.OpenFile(f.File,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	file.Write(data)
	file.Close()
	//ioutil.WriteFile(f.File, data, 0644)
	return f.Dst.Write(data)
}

type ProxyHandle struct {
	fakeSocket string
	realSocket string
	stop       chan interface{}
	conns      *list.List
}

func Launch(fakeSocket string, realSocket string) *ProxyHandle {
	handle := &ProxyHandle{fakeSocket, realSocket, make(chan interface{}), list.New()}
	handle.launch()
	return handle
}

func (h *ProxyHandle) launch() {

	l, err := net.Listen("unix", h.fakeSocket)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go startAccepting(h, l)

}

func startAccepting(h *ProxyHandle, l net.Listener) {
	defer l.Close()
	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		simulatorConnection, err := net.Dial("unix", h.realSocket)
		if err != nil {
			log.Fatal("Could not connect to usbmuxd socket, is it running?", err)
		}
		h.conns.PushBack(conn)
		h.conns.PushBack(simulatorConnection)
		log.Printf("Connection: %d", h.conns.Len())
		pathIn := fmt.Sprintf("conn-in%d.dump", h.conns.Len())
		pathOut := fmt.Sprintf("conn-out%d.dump", h.conns.Len())
		go forwardingConnection(NewFileAndSocketWriter(pathIn, conn), simulatorConnection)
		go forwardingConnection(NewFileAndSocketWriter(pathOut, simulatorConnection), conn)

	}
}

func forwardingConnection(c io.Writer, simulatorConnection net.Conn) {
	// /log.Printf("Client connected [%s]", c.RemoteAddr().Network())
	io.Copy(c, simulatorConnection)
}

func (h *ProxyHandle) Stop() {
	log.Println("closing conns..")
	for temp := h.conns.Front(); temp != nil; temp = temp.Next() {
		conn := temp.Value.(net.Conn)
		conn.Close()
	}
}
