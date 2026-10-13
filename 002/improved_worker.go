package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Factor XI (Logs): Logging zwingend auf Standard-Output (Stdout)
	logger := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)

	// 2. Factor III (Config): Credentials über Umgebungsvariablen injizieren
	dbHost := os.Getenv("DB_HOST")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbHost == "" || dbPassword == "" {
		logger.Fatal("required environment variables DB_HOST and DB_PASSWORD not set")
	}
	// Wichtig: Niemals Credentials ins Log schreiben!
	logger.Printf("connecting DB: %s...", dbHost)

	// 3. Factor IX (Disposability): Channel für OS-Signale (Kubernetes SIGTERM) einrichten
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Asynchron auf Abbruch-Signale warten
	go func() {
		sig := <-sigChan
		logger.Printf("Sig %v received. Shutdown initiated ...", sig)
		cancel() // Kontext abbrechen
	}()

	logger.Println("worker started...")

	// 4. Factor VI (Processes): Hauptschleife ohne lokalen Zustand (Stateless)
	for {
		select {
		case <-ctx.Done():
			// Wenn cancel() aufgerufen wurde, springt er hier rein
			logger.Println("canceling...")
			break
		default:
			taskID := fmt.Sprintf("task_%d", time.Now().Unix())
			logger.Printf("processing: %s", taskID)

			// Der State liegt extern! Wir verlassen uns auf eine externe DB / Queue.
			// simulate processing
			time.Sleep(2 * time.Second)
		}
	}

	logger.Println("worker completed")
}
