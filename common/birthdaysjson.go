package common

type BirthdaysJson struct {
	Birthdays []BirthdayJson `json:"birthdays"`
}

type BirthdayJson struct {
	ChatId int64  `json:"chatId"`
	UserId int64  `json:"userId"`
	Name   string `json:"name"`
}
