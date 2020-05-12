package main

import (
	"log"
	"os"
	"os/signal"

	fu "github.com/danielpaulus/go-simulator-dump/fileutil"
	"github.com/danielpaulus/go-simulator-dump/proxy"
	"github.com/docopt/docopt-go"
)

func main() {
	Main()
}

const version = "v 0.01"

// Main Exports main for testing
func Main() {
	usage := `iOS client v 0.01

Usage:
  sim listen [<sock>]
  sim ls

  The commands work as following:
  sim ls  will dump a list of currently active testmanagerd simulator sockets. Copy paste a path out of there to use with listen
  sim listen  will either take the first available simulator that is running or you can pass it a socket path for a specific sim if you want. once it is running, start a xcuitest in xcode and watch the files with DTX dump being created
`
	arguments, err := docopt.ParseDoc(usage)
	if err != nil {
		log.Fatal(err)
	}

	ls, _ := arguments.Bool("ls")
	if ls {
		list, err := fu.ListSockets()
		if err != nil {
			log.Printf("Could not get sockets because: %s", err)
			return
		}
		log.Println(list)
		return
	}

	sock, _ := arguments.String("<sock>")
	if sock == "" {
		log.Print("No socket specified, trying to find active sockets..")
		sock, err = fu.FirstSocket()
		if err != nil {
			log.Fatal("could not find socket")
		}
		log.Printf("Using socket:%s", sock)
	}
	newSocket, _ := fu.MoveSock(sock)
	handle := proxy.Launch(sock, newSocket)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Print("CTRL+C detected, shutting down")
	handle.Stop()
	fu.MoveBack(sock)
}
