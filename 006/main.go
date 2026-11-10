package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"
)

const (
	SubjectJobsNotify = "jobs.notify"
)

type JobData struct {
	JobID    string   `json:"job_id"`
	Channels []string `json:"channels"`
}

func main() {
	mode := os.Getenv("APP_MODE")
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Fehler bei NATS Verbindung: %v", err)
	}
	defer nc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if mode == "consumer" {
		startConsumer(ctx, nc)
	} else {
		startProducer(nc)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("Beende Service...")
}

func startProducer(nc *nats.Conn) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/v1/jobs:notify", func(w http.ResponseWriter, r *http.Request) {
		var jd JobData
		if err := json.NewDecoder(r.Body).Decode(&jd); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		event := cloudevents.NewEvent()
		event.SetID(fmt.Sprintf("%d", time.Now().UnixNano()))
		event.SetSource("com.example.jobservice")
		event.SetType("com.example.jobs.notify")
		event.SetTime(time.Now())
		if err := event.SetData(cloudevents.ApplicationJSON, jd); err != nil {
			http.Error(w, "Error setting event data", http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "Error marshaling event", http.StatusInternalServerError)
			return
		}

		if err := nc.Publish(SubjectJobsNotify, bytes); err != nil {
			http.Error(w, "Error publishing to NATS", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status": "accepted"}`))
	})

	port := ":8080"
	log.Printf("Starte PRODUCER: HTTP Router lauscht auf Port %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("HTTP Server Fehler: %v", err)
	}
}

func startConsumer(ctx context.Context, nc *nats.Conn) {
	log.Printf("Starte CONSUMER: Lausche auf NATS Subject '%s'", SubjectJobsNotify)

	sub, err := nc.Subscribe(SubjectJobsNotify, func(msg *nats.Msg) {
		var event cloudevents.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Ignoriere ungültiges CloudEvent: %v", err)
			return
		}

		var jd JobData
		if err := event.DataAs(&jd); err != nil {
			log.Printf("Konnte Event-Data nicht als JobData lesen: %v", err)
			return
		}

		if slices.Contains(jd.Channels, "email") {
			log.Printf("successfully send to channel: email (JobID: %s, EventID: %s)", jd.JobID, event.ID())
			return
		}

		log.Printf("Ignoriere Event %s - kein 'email' Channel (Channels: %v)", event.ID(), jd.Channels)
	})

	if err != nil {
		log.Fatalf("Fehler beim Abonnieren: %v", err)
	}

	<-ctx.Done()
	sub.Unsubscribe()
}
