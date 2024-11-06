package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	log.Printf("Server Started on localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
