package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	args := os.Args
	if len(args) == 1 {
		log.Fatal("Path argument missing")
	}

	path := args[1]
	log.Printf("Serving path: %v\n", path)

	fs := http.FileServer(http.Dir(path))
	http.Handle("/", fs)

	log.Println("Server Started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
