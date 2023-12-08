package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var responseTime time.Duration

func main() {
	// Get values from environment variables with error handling
	name, ok := os.LookupEnv("NAME")
	if !ok {
		log.Fatal("Environment variable NAME is not set")
	}

	uri, ok := os.LookupEnv("URI")
	if !ok {
		log.Fatal("Environment variable URI is not set")
	}

	isPausedStr, ok := os.LookupEnv("IS_PAUSED")
	if !ok {
		log.Fatal("Environment variable IS_PAUSED is not set")
	}
	isPaused, err := strconv.ParseBool(isPausedStr)
	if err != nil {
		log.Fatalf("Error converting IS_PAUSED to boolean: %v", err)
	}

	numRetriesStr, ok := os.LookupEnv("NUM_RETRIES")
	if !ok {
		log.Fatal("Environment variable NUM_RETRIES is not set")
	}
	numRetries, err := strconv.Atoi(numRetriesStr)
	if err != nil {
		log.Fatalf("Error converting NUM_RETRIES to integer: %v", err)
	}

	SLAStr, ok := os.LookupEnv("RESPONSE_TIME_SLA")
	if !ok {
		log.Fatal("Environment variable RESPONSE_TIME_SLA is not set")
	}
	SLA, err := strconv.Atoi(SLAStr)
	if err != nil {
		log.Fatalf("Error converting RESPONSE_TIME_SLA to duration: %v", err)
	}
	var responseTimeSLA time.Duration = time.Duration(SLA) * time.Millisecond

	useSSLStr, ok := os.LookupEnv("USE_SSL")
	if !ok {
		log.Fatal("Environment variable USE_SSL is not set")
	}
	useSSL, err := strconv.ParseBool(useSSLStr)
	if err != nil {
		log.Fatalf("Error converting USE_SSL to boolean: %v", err)
	}

	responseStatusCode := "0"
	checkIntervalInSeconds := "86400"
	var checkstatus bool = false

	// Retry HTTP GET attempts
	for retry := 0; retry <= numRetries; retry++ {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: !useSSL},
			},
			Timeout: time.Second * 5,
		}

		startTime := time.Now()
		resp, err := client.Get(uri)
		endTime := time.Now()

		responseTime = endTime.Sub(startTime)

		// Check if the HTTP GET was successful and within the response time SLA
		if err == nil && resp.StatusCode == http.StatusOK && responseTime <= responseTimeSLA {
			log.Println("HTTP GET successful within SLA")
			responseStatusCode = "200"
			checkstatus = true
			break
		}

		// Log unsuccessful attempt
		log.Printf("Attempt %d: HTTP GET failed or exceeded response time SLA\n", retry+1, responseTime, resp.StatusCode, responseTimeSLA)

		// Sleep before the next retry
		if retry < numRetries {
			sleepDuration := time.Second * 5 // Adjust as needed
			log.Printf("Sleeping for %s before next retry\n", sleepDuration)
			time.Sleep(sleepDuration)
		}
		responseStatusCode = "404"
	}
	
	uptimeSLA := "100"

	// Create message payload without Kafka
	message := fmt.Sprintf(`{
		"name": "%s",
		"uri": "%s",
		"is_paused": %t,
		"num_retries": %d,
		"uptime_sla": %s,
		"response_time_sla": %d,
		"use_ssl": %t,
		"response_status_code": %s,
		"check_interval_in_seconds": %s,
		"http_get_success": %t,
		"response_time": %.2f
	}`, name, uri, isPaused, numRetries, uptimeSLA, responseTimeSLA.Milliseconds(), useSSL, responseStatusCode, checkIntervalInSeconds, checkstatus, responseTime.Seconds())

	// Log the message without Kafka
	log.Println("Message without Kafka:", message)

	// Kafka configuration with error handling
	kafkaBootstrapServers, ok := os.LookupEnv("KAFKA_BOOTSTRAP_SERVERS")
	fmt.Println("Kafka server:", kafkaBootstrapServers)
	if !ok {
		log.Fatal("Environment variable KAFKA_BOOTSTRAP_SERVERS is not set")
	}

	kafkaTopic := "httpcheck"

	if !ok {
		log.Fatal("topic not found")
	}

	// Create Kafka producer configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers": kafkaBootstrapServers,
		"security.protocol": "SASL_PLAINTEXT",
		"sasl.mechanism":    "PLAIN",
		"sasl.username":     "user1",
		"sasl.password":     "user1",
	}

	// Create Kafka producer
	producer, err := kafka.NewProducer(config)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer producer.Close()

	// Produce message to Kafka topic
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaTopic, Partition: kafka.PartitionAny},
		Value:          []byte(message),
	}, nil)

	if err != nil {
		log.Fatalf("Error producing message to Kafka: %v", err)
	}
	producer.Flush(5 * 1000) // 5 seconds timeout

	log.Println("Message sent to Kafka successfully")
}
