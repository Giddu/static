package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/fsnotify/fsnotify"
)

var (
	servingPath  string
	reloadMutex  sync.Mutex
	shouldReload bool
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "port to use for serving")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Path argument missing")
	}

	servingPath = args[0]
	log.Printf("Serving servingPath: %v\n", servingPath)

	go watchPath(servingPath)

	http.HandleFunc("/", fsHandler)
	http.HandleFunc("/reload", reloadHandler)

	log.Printf("Server Started on localhost:%d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func watchPath(servingPath string) {
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
					reloadMutex.Lock()
					shouldReload = true
					reloadMutex.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(servingPath)
	if err != nil {
		log.Fatal(err)
	}

}

func fsHandler(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path[1:]
	if requestPath == "" {
		requestPath = "."
	}

	stat, err := fs.Stat(os.DirFS(servingPath), requestPath)
	if err != nil {
		fmt.Println(err)
	}

	filePath := filepath.Join(servingPath, requestPath)
	if stat.IsDir() {
		filePath = filepath.Join(servingPath, "index.html")
	}

	fmt.Println("filePath", filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if strings.HasSuffix(filePath, ".html") {
		reloadScript := `<script>
            const es = new EventSource('/reload');
            es.onmessage = () => window.location.reload();
        </script>`
		tmpl := template.Must(template.New("html").Parse(string(content) + reloadScript))
		tmpl.Execute(w, nil)
	} else {
		w.Write(content)
	}
}

func reloadHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		reloadMutex.Lock()
		if shouldReload {
			fmt.Fprintf(w, "data: reload\n\n")
			w.(http.Flusher).Flush()
			shouldReload = false
		}
		reloadMutex.Unlock()
	}
}
