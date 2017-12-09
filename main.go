package main

import (
	"bytes"
	"flag"
	"github.com/flosch/pongo2"
	"github.com/fsnotify/fsnotify"
	"github.com/jimlawless/cfg"
	"log"
	"goji.io"
	"goji.io/pat"
	"net/http"
	"time"
)

const MESSAGES_FILE = "resources/messages.txt"

var (
	listenAddr     = flag.String("addr", ":8081", "Address to listen on")
	templateReload = flag.Bool("reload", false, "Reload templates and messages when changes are detected")
	secretKey      = flag.String("key", "key", "Key of secret for redacting text")
	secret         = flag.String("secret", "", "Secret for redacting text")
	messages       = loadMessages()
	publicPath     = "resources/static/"
	fs             http.Handler
	indexTemplate  *pongo2.Template
)

func init() {
	// Redact filter replaces each character with the specified newChar and wraps the word with a span class for styling
	// unless the param is a positive boolean in which case the inputted value is returned as is.
	pongo2.RegisterFilter("redact", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		if *secret == "" || param.Bool() {
			return in, nil
		}
		var buf bytes.Buffer
		spaceIndex := 0
		for idx, char := range in.String() {
			newChar := '#'
			if char == ' ' {
				spaceIndex = idx
				newChar = ' '
			}
			if idx == 0 || (idx == spaceIndex+1 && spaceIndex > 0) {
				buf.WriteString("<span class=\"redacted\">")
			} else if idx == len(in.String())-1 {
				buf.WriteRune(newChar)
				buf.WriteString("</span>")
				continue
			} else if idx == spaceIndex || idx == len(in.String())-1 {
				buf.WriteString("</span>")
			}
			buf.WriteRune(newChar)
		}
		return pongo2.AsValue(buf.String()), nil
	})
}

func main() {
	flag.Parse()

	go func() {
		watchMessages()
	}()

	mux := goji.NewMux()
	fs = http.StripPrefix("/static/", http.FileServer(http.Dir(publicPath)))

	log.Println("Mapping routes...")
	mux.HandleFunc(pat.Get("/"), index)
	mux.HandleFunc(pat.Get("/static/*"), serveStatic)

	log.Println("Binding to port", *listenAddr)
	http.ListenAndServe(*listenAddr, mux)
}

func index(writer http.ResponseWriter, request *http.Request) {
	if *templateReload || indexTemplate == nil {
		indexTemplate = pongo2.Must(pongo2.FromFile("resources/templates/index.html"))
	}

	if err := indexTemplate.ExecuteWriter(getContext(request), writer); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

func serveStatic(writer http.ResponseWriter, request *http.Request) {
	fs.ServeHTTP(writer, request)
}

func getContext(request *http.Request) pongo2.Context {
	key := request.URL.Query().Get(*secretKey)
	year := time.Now().Year()
	return pongo2.Context{
		"messages": messages,
		"unlock":   key == *secret,
		"year":     year,
	}
}

func watchMessages() {
	if !*templateReload {
		return
	}
	log.Println("Watching messages...")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Reloading messages")
					messages = loadMessages()
				}
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}()

	err = watcher.Add(MESSAGES_FILE)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func loadMessages() map[string]string {
	messages := make(map[string]string)
	cfg.Load(MESSAGES_FILE, messages)
	log.Println("Loaded", len(messages), "messages")
	return messages
}
