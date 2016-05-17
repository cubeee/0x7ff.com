package main

import (
	"bytes"
	"flag"
	"github.com/echo-contrib/pongor"
	"github.com/flosch/pongo2"
	"github.com/fsnotify/fsnotify"
	"github.com/jimlawless/cfg"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"log"
)

const MESSAGES_FILE = "resources/messages.txt"

var (
	listenAddr     = flag.String("addr", ":8081", "Address to listen on")
	templateReload = flag.Bool("reload", false, "Reload templates and messages when changes are detected")
	secretKey      = flag.String("key", "key", "Key of secret for redacting text")
	secret         = flag.String("secret", "", "Secret for redacting text")
	messages       = loadMessages()
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
	log.Println("Welcome to 0x7ff site")

	e := echo.New()
	renderer := pongor.GetRenderer()
	renderer.Directory = "resources/templates/"
	renderer.Reload = *templateReload
	e.SetRenderer(renderer)

	go func() {
		watchMessages()
	}()

	e.Pre(middleware.AddTrailingSlash())
	e.GET("/", func(ctx echo.Context) error {
		err := renderer.Render(ctx.Response().Writer(), "index.html", getContext(ctx), ctx)
		if err != nil {
			log.Println(err.Error())
		}
		return err
	})
	e.Static("/static", "resources/static")
	e.Run(standard.New(*listenAddr))
}

func getContext(ctx echo.Context) map[string]interface{} {
	key := ctx.QueryParam(*secretKey)
	return map[string]interface{}{
		"messages": messages,
		"unlock":   key == *secret,
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
