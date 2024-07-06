package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/4Kaze/birthdaybot/common"
	"github.com/4Kaze/birthdaybot/manager/adapters"
	"github.com/4Kaze/birthdaybot/manager/core"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/pgxpool"
)

var birthdayManager *core.BirthdayManager

const REQUEST_PARAM_DATE_LAYOUT = "2006-01-02"

func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	token := os.Getenv("BOT_TOKEN")
	port := os.Getenv("PORT")
	ctx := context.Background()
	telegramBot, err := telegram.New(token)
	if err != nil {
		log.Fatalf("Failed to instantiate telegram bot due to: %v\n", err)
	}
	birthdayManager = createManager(ctx, databaseUrl, telegramBot)
	go setWebhook(ctx, telegramBot, token)
	http.HandleFunc(fmt.Sprintf("/%s", token), HandleUpdate)
	http.HandleFunc("/birthdays", GetBirthdays)
	err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		log.Fatalf("Failed to start an http server due to: %v\n", err)
	}
}

func createManager(ctx context.Context, databaseUrl string, telegramBot *telegram.Bot) *core.BirthdayManager {
	db, err := pgxpool.New(ctx, databaseUrl)
	if err != nil {
		log.Fatalf("Failed to instantiate database client due to: %v\n", err)
	}
	botUser, err := telegramBot.GetMe(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch bot profile due to: %v\n", err)
	}
	log.Printf("Fetched bot profile: %v\n", botUser)
	botWrapper := adapters.NewTelegramWrapper(ctx, telegramBot)
	repository := adapters.NewSqlRepositoryAdapter(db, &common.SystemClock{})
	return core.NewBirthdayManager(repository, botWrapper, botUser.ID)
}

func setWebhook(ctx context.Context, bot *telegram.Bot, telegramToken string) {
	serviceLocation, serviceName := common.GetServiceLocationAndName()
	serviceUrl, err := common.GetServiceUrl(serviceLocation, serviceName)
	if err != nil {
		log.Fatalf("Failed to get service url: %v\n", err)
	}
	webhookUrl := fmt.Sprintf("%s/%s", strings.TrimRight(serviceUrl, "/"), telegramToken)
	fmt.Printf("Setting webhook to: %v/*****\n", serviceUrl)
	_, err = bot.SetWebhook(ctx, &telegram.SetWebhookParams{
		URL: webhookUrl,
	})
	if err != nil {
		log.Fatalf("Failed to set webhook: %v\n", err)
	}
}

func HandleUpdate(w http.ResponseWriter, r *http.Request) {
	update := models.Update{}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		common.ErrorLogger.Printf("Could not decode update: %v\n", err)
		w.WriteHeader(400)
		return
	}
	if err := birthdayManager.HandleUpdate(r.Context(), &update); err != nil {
		common.ErrorLogger.Printf("Error handling update: %v\n", err)
		w.WriteHeader(500)
		return
	}
}

func GetBirthdays(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a new request: %s %s\n", r.Method, r.URL)
	dateString := r.URL.Query().Get("date")
	date, err := time.Parse(REQUEST_PARAM_DATE_LAYOUT, dateString)
	if err != nil {
		common.ErrorLogger.Printf("Could not decode date (%v): %v\n", dateString, err)
		w.WriteHeader(400)
		return
	}
	birthdays, err := birthdayManager.GetBirthdays(r.Context(), date)
	if err != nil {
		common.ErrorLogger.Printf("Error getting birthdays: %v\n", err)
		w.WriteHeader(500)
		return
	}
	log.Printf("Returning response: %v\n", birthdays)
	response := common.BirthdaysJson{Birthdays: mapBirthdays(birthdays)}
	responseBytes, err := json.Marshal(response)
	if err != nil {
		common.ErrorLogger.Printf("Error marshalling birthdays response: %v\n", err)
		w.WriteHeader(500)
		return
	}
	_, err = w.Write(responseBytes)
	if err != nil {
		common.ErrorLogger.Printf("Error writing response: %v\n", err)
		w.WriteHeader(500)
	}
}

func mapBirthdays(birthdays []core.BirthdayPerson) []common.BirthdayJson {
	birthdaysJson := make([]common.BirthdayJson, len(birthdays))
	for index, birthday := range birthdays {
		birthdaysJson[index] = common.BirthdayJson{
			ChatId: birthday.ChatId,
			UserId: birthday.UserId,
			Name:   birthday.Name,
		}
	}
	return birthdaysJson
}
