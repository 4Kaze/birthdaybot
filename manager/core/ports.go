package core

import (
	"context"
	"time"
)

type Birthday struct {
	Date          time.Time
	ChatId        int64
	UserId        int64
	Username      string
	UserFirstName string
	UserLastName  string
}

type Repository interface {
	SaveBirthday(ctx context.Context, birthday Birthday) error
	GetBirthdayDate(ctx context.Context, chatId int64, userId int64) (*time.Time, error)
	GetNextBirthdays(ctx context.Context, chatId int64) ([]Birthday, error)
	GetBirthdaysForDate(ctx context.Context, date time.Time) ([]Birthday, error)
	DeleteBirthday(ctx context.Context, chatId int64, userId int64) error
	DeleteAllChatBirthdays(ctx context.Context, chatId int64) error
	DeleteAllUserBirthdays(ctx context.Context, userId int64) error
}

type Telegram interface {
	SendMessage(ctx context.Context, chatId int64, text string) error
	SendReply(ctx context.Context, chatId int64, messageId int, text string) error
	SendReaction(ctx context.Context, chatId int64, messageId int, reaction string) error
}
