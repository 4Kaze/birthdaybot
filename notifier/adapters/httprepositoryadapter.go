package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/4Kaze/birthdaybot/common"
	"github.com/4Kaze/birthdaybot/notifier/core"
)

type HttpRepositoryAdapter struct {
	repositoryUrl string
}

func NewHttpRepositoryAdapter(repositoryUrl string) *HttpRepositoryAdapter {
	return &HttpRepositoryAdapter{repositoryUrl: repositoryUrl}
}

func (adapter HttpRepositoryAdapter) GetBirthdays(ctx context.Context, date time.Time) ([]core.Birthday, error) {
	url := fmt.Sprintf("%s/birthdays?date=%s", adapter.repositoryUrl, date.Format(DATE_LAYOUT))
	log.Printf("Sending a request to get birthdays: GET %v\n", url)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		common.ErrorLogger.Printf("Failed to create a request to fetch birthdays: %v\n", err)
		return nil, err
	}
	response, err := http.DefaultClient.Do(request.WithContext(ctx))
	if err != nil {
		common.ErrorLogger.Printf("Failed to send a request to fetch birthdays: %v\n", err)
		return nil, err
	}
	var birthdays common.BirthdaysJson
	err = json.NewDecoder(response.Body).Decode(&birthdays)
	if err != nil {
		common.ErrorLogger.Printf("Failed to decode a response with birthdays: %v\n", err)
		return nil, err
	}
	log.Printf("Received a response with birthdays: %v\n", birthdays)
	return mapBirthdays(birthdays), nil
}

func mapBirthdays(birthdaysJson common.BirthdaysJson) []core.Birthday {
	birthdays := make([]core.Birthday, len(birthdaysJson.Birthdays))
	for index, birthday := range birthdaysJson.Birthdays {
		birthdays[index] = core.Birthday{
			ChatId: birthday.ChatId,
			UserId: birthday.UserId,
			Name:   birthday.Name,
		}
	}
	return birthdays
}

const DATE_LAYOUT = "2006-01-02"
