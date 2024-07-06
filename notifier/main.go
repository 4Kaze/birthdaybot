package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/4Kaze/birthdaybot/common"
	"github.com/4Kaze/birthdaybot/notifier/adapters"
	"github.com/4Kaze/birthdaybot/notifier/core"
	telegram "github.com/go-telegram/bot"
)

var birthdayNotifier *core.BirthdayNotifier

func main() {
	token := os.Getenv("BOT_TOKEN")
	managerUrl := os.Getenv("MANAGER_URL")
	port := os.Getenv("PORT")
	cloudTasksQueueId := os.Getenv("QUEUE_ID")
	cloudTasksDeadlineInSecondsStr := os.Getenv("TASK_DEADLINE_S")
	cloudTasksDeadlineInSeconds, err := strconv.Atoi(cloudTasksDeadlineInSecondsStr)
	if err != nil {
		log.Fatalf("Incorrect cloud task deadline format - must be a number, is: %v\n", cloudTasksDeadlineInSecondsStr)
	}
	cloudTasksDelayInSecondsStr := os.Getenv("TASK_DELAY_S")
	cloudTasksDelayInSeconds, err := strconv.Atoi(cloudTasksDelayInSecondsStr)
	if err != nil {
		log.Fatalf("Incorrect cloud task delay format - must be a number, is: %v\n", cloudTasksDelayInSecondsStr)
	}
	birthdayNotifier = createNotifier(token, managerUrl, cloudTasksQueueId, cloudTasksDeadlineInSeconds, cloudTasksDelayInSeconds)
	http.HandleFunc("/schedule", HandleScheduleBirthdayNotifications)
	http.HandleFunc("/notify", HandleSendBirthdayNotification)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatalf("Failed to start an http server due to: %v\n", err)
	}
}

func createNotifier(token string, managerUrl string, cloudTasksQueueId string, cloudTasksDeadline int, cloudTasksDelay int) *core.BirthdayNotifier {
	telegramBot, err := telegram.New(token)
	if err != nil {
		log.Fatalf("Failed to instantiate telegram bot due to: %v\n", err)
	}
	botWrapper := adapters.NewTelegramWrapper(telegramBot)
	repository := adapters.NewHttpRepositoryAdapter(managerUrl)
	ctx := context.Background()
	cloudTasksClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to instantiate PubSub client due to: %v\n", err)
	}
	clock := common.SystemClock{}
	scheduler := adapters.NewCloudTasksScheduler(
		cloudTasksClient,
		clock,
		cloudTasksQueueId,
		time.Duration(cloudTasksDeadline)*time.Second,
		time.Duration(cloudTasksDelay)*time.Second,
	)
	fileDownloader := adapters.NewHttpFileDownloader()
	videoGenerator := adapters.NewVideoGenerator("/resources")
	birthdayNotifier = core.NewBirthdayNotifier(repository, botWrapper, scheduler, fileDownloader, videoGenerator, clock)
	return birthdayNotifier
}

func HandleScheduleBirthdayNotifications(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a request to schedule birthdays")
	serviceLocation, serviceName := common.GetServiceLocationAndName()
	serviceUrl, err := common.GetServiceUrl(serviceLocation, serviceName)
	if err != nil {
		log.Fatalf("Failed to get service url due to: %v\n", err)
	}
	fmt.Printf("Got service url: %s\n", serviceUrl)
	notificationEndpointUrl := fmt.Sprintf("%s/notify", serviceUrl)
	if err := birthdayNotifier.ScheduleBirthdayNotifications(r.Context(), notificationEndpointUrl); err != nil {
		w.WriteHeader(500)
		common.ErrorLogger.Printf("Failed to schedule birthdays: %v\n", err)
	}
}

func HandleSendBirthdayNotification(w http.ResponseWriter, r *http.Request) {
	birthday := common.BirthdayJson{}
	if err := json.NewDecoder(r.Body).Decode(&birthday); err != nil {
		common.ErrorLogger.Printf("Could not decode birthday from topic message: %v\n", err)
		w.WriteHeader(400)
		return
	}
	log.Printf("Received a request to send a birthday notification: %v\n", birthday)
	if err := birthdayNotifier.SendBirthdayNotification(r.Context(), core.Birthday(birthday)); err != nil {
		common.ErrorLogger.Printf("Could not send birthday notification: %v\n", err)
		w.WriteHeader(500)
	}
}
