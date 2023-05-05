package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	_ "embed"

	"github.com/diamondburned/arikawa/v3/api/webhook"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
)

var webhookRe = regexp.MustCompile(`/webhooks/(\d+)/(.+)`)

func NewWebhookClient(url string) (*webhook.Client, error) {
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

type Handler struct {
	*http.ServeMux
	webhook *webhook.Client
}

func NewHandler(webhook *webhook.Client) *Handler {
	h := &Handler{webhook: webhook}

	h.ServeMux = http.NewServeMux()
	h.HandleFunc("/", h.handle)

	return h
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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
}

func writeErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}
