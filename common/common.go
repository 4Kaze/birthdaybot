package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var ErrorLogger = log.New(os.Stderr, "", log.Lshortfile)
var serviceStatusMaxRetries = 10
var serviceStatusBackoffMilis = 300

func init() {
	if maxRetries, present := os.LookupEnv("SERVICE_STATUS_MAX_RETRIES"); present {
		maxRetriesInt, err := strconv.Atoi(maxRetries)
		if err != nil {
			log.Fatalln("'SERVICE_STATUS_MAX_RETRIES' environment variable is not an int")
		}
		serviceStatusMaxRetries = maxRetriesInt
	}
	if backoffMilis, present := os.LookupEnv("SERVICE_STATUS_BACKOFF_MILIS"); present {
		backoffMilisInt, err := strconv.Atoi(backoffMilis)
		if err != nil {
			log.Fatalln("'SERVICE_STATUS_BACKOFF_MILIS' environment variable is not an int")
		}
		serviceStatusBackoffMilis = backoffMilisInt
	}
}

func GetServiceUrl(serviceLocation string, serviceName string) (string, error) {
	token, err := getToken()
	if err != nil {
		return "", err
	}
	return pollForServiceUrl(token, serviceLocation, serviceName)
}

func GetServiceLocationAndName() (serviceLocation string, serviceName string) {
	var location, name string
	var present bool
	if location, present = os.LookupEnv("LOCATION"); !present {
		log.Fatalln("'LOCATION' environment variable must be set")
	}
	if name, present = os.LookupEnv("K8S_SERVICE_NAME"); !present {
		log.Fatalln("'SERVICE_NAME' environment variable must be sent")
	}
	return location, name
}

func pollForServiceUrl(token string, serviceLocation string, serviceName string) (string, error) {
	for retry := range serviceStatusMaxRetries {
		log.Printf("Fetching service status. Attempt %v/%v...\n", retry+1, serviceStatusMaxRetries)
		status, err := fetchServiceStatus(token, serviceLocation, serviceName)
		if err != nil {
			ErrorLogger.Printf("Failed to fetch service status: %v\n", err)
		} else if isServiceReady(*status) {
			log.Printf("Service status reported Ready: %v\n", status)
			return status.Status.Url, nil
		} else {
			log.Printf("Service status not ready: %v\n", status)
		}
		time.Sleep(time.Duration(serviceStatusBackoffMilis) * time.Millisecond)
	}
	ErrorLogger.Println("Reached max number of retries when trying to fetch the service status")
	return "", errors.New("timed-out fetching service status")
}

func fetchServiceStatus(token string, serviceLocation string, serviceName string) (*Service, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("https://%v-run.googleapis.com/apis/serving.knative.dev/v1/%v", serviceLocation, serviceName), nil)
	if err != nil {
		ErrorLogger.Printf("Could not create service request: %v\n", err)
		return nil, err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		ErrorLogger.Printf("Could not get service info: %v\n", err)
		return nil, err
	}
	if response.StatusCode != 200 {
		responseBytes, _ := io.ReadAll(response.Body)
		ErrorLogger.Printf("Request to get service info returned tatus: %v, body: %v\n", response.Status, string(responseBytes))
		return nil, errors.New("failed request to get service info")
	}
	var service Service
	err = json.NewDecoder(response.Body).Decode(&service)
	if err != nil {
		ErrorLogger.Printf("Could not decode service info: %v\n", err)
		return nil, err
	}
	return &service, nil
}

func getToken() (string, error) {
	request, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token", nil)
	if err != nil {
		ErrorLogger.Printf("Could not create token request: %v\n", err)
		return "", err
	}
	request.Header.Set("Metadata-Flavor", "Google")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		ErrorLogger.Printf("Could not get token: %v\n", err)
		return "", err
	}
	var token Token
	err = json.NewDecoder(response.Body).Decode(&token)
	if err != nil {
		ErrorLogger.Printf("Could not decode token: %v\n", err)
		return "", err
	}
	return token.Value, nil
}

func isServiceReady(service Service) bool {
	for _, condition := range service.Status.Conditions {
		if condition.Type == "Ready" {
			return condition.Status == "True"
		}
	}
	ErrorLogger.Println("Could not find 'Ready' condition in conditions list.")
	return false
}

type Token struct {
	Value string `json:"access_token"`
}

type Service struct {
	Status ServiceStatus `json:"status"`
}

type ServiceStatus struct {
	Conditions []ServiceCondition `json:"conditions"`
	Url        string             `json:"url"`
}

type ServiceCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}
