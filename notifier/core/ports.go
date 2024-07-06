package core

import (
	"context"
	"time"
)

type Repository interface {
	GetBirthdays(ctx context.Context, date time.Time) ([]Birthday, error)
}

type Birthday struct {
	ChatId int64
	UserId int64
	Name   string
}

type Telegram interface {
	SendMessage(ctx context.Context, chatId int64, text string) error
	SendVideo(ctx context.Context, chatId int64, pathToVideo string) (fileId string, err error)
	SendVideoFromFileId(ctx context.Context, chatId int64, fileId string) error
	GetProfilePictureFileIds(ctx context.Context, userId int64) ([]string, error)
	GetFileLink(ctx context.Context, fileId string) (string, error)
}

type Clock interface {
	Now() time.Time
}

type VideoGenerator interface {
	CreateVideo(pathToProfilePicture string) (string, error)
}

type BirthdayNotificationScheduler interface {
	Schedule(ctx context.Context, birthday Birthday, serviceUrl string)
}

type FileDownloader interface {
	Download(ctx context.Context, link string) (string, error)
}
