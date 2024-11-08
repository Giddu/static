package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/fsnotify/fsnotify"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "port to use for serving")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Path argument missing")
	}

	path := args[0]
	log.Printf("Serving path: %v\n", path)

	go watchPath(path)

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	log.Printf("Server Started on localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func watchPath(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

}
