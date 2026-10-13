package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

var processedNotifications = make(map[string]bool)

func main() {
	dbUser := "admin"
	dbPass := "superSecret123!"
	dbHost := "localhost:5432"

	connectionString := fmt.Sprintf("postgres://%s:%s@%s/mydb", dbUser, dbPass, dbHost)
	fmt.Println("DB connection:", connectionString)

	logFile, err := os.OpenFile("/var/log/notification-worker.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	logger := log.New(logFile, "INFO: ", log.Ldate|log.Ltime)

	logger.Println("worker started...")

	for {
		taskID := fmt.Sprintf("task_%d", time.Now().Unix())

		if !processedNotifications[taskID] {
			logger.Printf("processing task: %s\n", taskID)
			processedNotifications[taskID] = true
		}

		// simulate processing
		time.Sleep(2 * time.Second)
	}
}
