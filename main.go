package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"etok.codes/discord_uploader/server"
	"github.com/joho/godotenv"
)

var port int

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

	webhook, err := server.NewWebhookClient(webhookURL)
	if err != nil {
		log.Fatalln("failed to create webhook client:", err)
	}

	h := server.NewHandler(webhook)

	log.Printf("listening on :%d", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), h))
}
