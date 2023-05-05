package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"

	"github.com/apex/gateway"

	"etok.codes/discord_uploader/server"
)

//go:embed index.html
var indexHTML string

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatalln("$DISCORD_WEBHOOK_URL is not set")
	}

	webhook, err := server.NewWebhookClient(webhookURL)
	if err != nil {
		log.Fatalln("failed to create webhook client:", err)
	}

	server := server.NewHandler(webhook)

	h := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.ServeContent(w, r, "index.html", time.Time{}, strings.NewReader(indexHTML))
		case http.MethodPost:
			server.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
	return gateway.ListenAndServe("", http.StripPrefix("/upload", http.HandlerFunc(h)))
}
