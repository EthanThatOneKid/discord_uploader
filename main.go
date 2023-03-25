package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/joho/godotenv"
)

var port int

var webhookRe = regexp.MustCompile(`/webhooks/(\d+)/(.+)`)

func main() {
	flag.IntVar(&port, "port", 8000, "port to listen on")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Println("warning: failed to load .env:", err)
	}

	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		log.Fatalln("$DISCORD_WEBHOOK_URL is not set")
	}

	webhook, err := newWebhookClient(webhookURL)
	if err != nil {
		log.Fatalln("failed to create webhook client:", err)
	}

	h := newHandler(webhook)

	log.Printf("listening on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), h))
}

func newWebhookClient(url string) (*webhook.Client, error) {
	matches := webhookRe.FindStringSubmatch(url)
	if matches == nil {
		return nil, errors.New("invalid webhook URL")
	}

	webhookID, err := discord.ParseSnowflake(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid webhook ID: %w", err)
	}

	return webhook.New(discord.WebhookID(webhookID), matches[2]), nil
}

type handler struct {
	*http.ServeMux
	webhook *webhook.Client
}

func newHandler(webhook *webhook.Client) *handler {
	h := &handler{webhook: webhook}

	h.ServeMux = http.NewServeMux()
	h.HandleFunc("/", h.handle)

	return h
}

func (h *handler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "index.html")
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		defer file.Close()

		msg, err := h.webhook.ExecuteAndWait(webhook.ExecuteData{
			Files: []sendpart.File{
				{
					Name:   fileHeader.Filename,
					Reader: file,
				},
			},
		})
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(msg)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}
