package main

import (
	"log"
	"os"

	"github.com/apex/gateway"

	"etok.codes/discord_uploader/server"
)

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

	h := server.NewHandler(webhook)
	return gateway.ListenAndServe("", h)
}
