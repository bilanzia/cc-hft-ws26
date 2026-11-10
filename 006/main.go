package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

// Response struct; used by `swag` setup for API documentation
type Response struct {
	Status string `json:"status" example:"accepted"`
}

// @title NATS Job Service API
// @version 1.0
// @description CloudEvents Producer via NATS.
// @host localhost:8080
// @BasePath /v1
func main() {
	// Initialize structured logger (JSON format is best practice for production environments)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mode := os.Getenv("APP_MODE")
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		slog.Error("Failed to connect to NATS", "error", err, "url", natsURL)
		os.Exit(1)
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
	slog.Info("Shutting down service...")
}

func startProducer(nc *nats.Conn) {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/v1/jobs:notify", handleJobNotify(nc))
	r.Post("/v1/jobs%3Anotify", handleJobNotify(nc)) // Dirty bugfix covering URL encoded paths

	port := ":8080"
	slog.Info("Starting PRODUCER: HTTP router listening", "port", port)

	if err := http.ListenAndServe(port, r); err != nil {
		slog.Error("HTTP server error", "error", err)
		os.Exit(1)
	}
}

func startConsumer(ctx context.Context, nc *nats.Conn) {
	slog.Info("Starting CONSUMER: Listening on NATS subject", "subject", SubjectJobsNotify)

	sub, err := nc.Subscribe(SubjectJobsNotify, func(msg *nats.Msg) {
		var event cloudevents.Event
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			slog.Warn("Ignoring invalid CloudEvent", "error", err)
			return
		}

		var jd JobData
		if err := event.DataAs(&jd); err != nil {
			slog.Warn("Failed to decode event data as JobData", "error", err, "event_id", event.ID())
			return
		}

		if slices.Contains(jd.Channels, "email") {
			slog.Info("Successfully sent to channel: email", "job_id", jd.JobID, "event_id", event.ID())
			return
		}

		slog.Info("Ignoring event - missing 'email' channel", "event_id", event.ID(), "channels", jd.Channels)
	})

	if err != nil {
		slog.Error("Failed to subscribe to NATS", "error", err)
		os.Exit(1)
	}

	<-ctx.Done()
	sub.Unsubscribe()
}

// handleJobNotify generates and publishes an event
// @Summary      Notify new job
// @Description  Accepts a job and publishes it as a CloudEvent to NATS
// @Tags         jobs
// @Accept       json
// @Produce      json
// @Param        payload  body      JobData   true  "Job Information"
// @Success      202      {object}  Response  "Successfully sent to NATS"
// @Failure      400      {string}  string    "Invalid JSON body"
// @Failure      500      {string}  string    "Internal Server Error"
// @Router       /jobs:notify [post]
func handleJobNotify(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var jd JobData
		if err := json.NewDecoder(r.Body).Decode(&jd); err != nil {
			slog.Warn("Invalid JSON body received", "error", err)
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		event := cloudevents.NewEvent()
		event.SetID(fmt.Sprintf("%d", time.Now().UnixNano()))
		event.SetSource("com.example.jobservice")
		event.SetType("com.example.jobs.notify")
		event.SetTime(time.Now())

		if err := event.SetData(cloudevents.ApplicationJSON, jd); err != nil {
			slog.Error("Error setting event data", "error", err)
			http.Error(w, "Error setting event data", http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(event)
		if err != nil {
			slog.Error("Error marshaling event", "error", err)
			http.Error(w, "Error marshaling event", http.StatusInternalServerError)
			return
		}

		if err := nc.Publish(SubjectJobsNotify, bytes); err != nil {
			slog.Error("Error publishing to NATS", "error", err)
			http.Error(w, "Error publishing to NATS", http.StatusInternalServerError)
			return
		}

		slog.Info("Job successfully published to NATS", "job_id", jd.JobID, "event_id", event.ID())

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status": "accepted"}`))
	}
}
