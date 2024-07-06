package core

import (
	"context"
	"fmt"

	"github.com/4Kaze/birthdaybot/common"
)

type BirthdayNotifier struct {
	repository                Repository
	scheduler                 BirthdayNotificationScheduler
	telegram                  Telegram
	fileDownloader            FileDownloader
	videoGenerator            VideoGenerator
	clock                     Clock
	userIdToCachedVideoFileId map[int64]string
}

func NewBirthdayNotifier(repository Repository, telegram Telegram, scheduler BirthdayNotificationScheduler, fileDownloader FileDownloader, videoGenerator VideoGenerator, clock Clock) *BirthdayNotifier {
	return &BirthdayNotifier{
		repository:                repository,
		scheduler:                 scheduler,
		telegram:                  telegram,
		fileDownloader:            fileDownloader,
		videoGenerator:            videoGenerator,
		clock:                     clock,
		userIdToCachedVideoFileId: make(map[int64]string),
	}
}

func (notifier BirthdayNotifier) ScheduleBirthdayNotifications(ctx context.Context, serviceUrl string) error {
	today := notifier.clock.Now()
	todayBirthdays, err := notifier.repository.GetBirthdays(ctx, today)
	if err != nil {
		return err
	}
	for _, birthday := range todayBirthdays {
		notifier.scheduler.Schedule(ctx, birthday, serviceUrl)
	}
	return nil
}

func (notifier BirthdayNotifier) SendBirthdayNotification(ctx context.Context, birthday Birthday) error {
	if fileId, isCached := notifier.userIdToCachedVideoFileId[birthday.UserId]; isCached {
		err := notifier.telegram.SendVideoFromFileId(ctx, birthday.ChatId, fileId)
		if err != nil {
			return err
		}
	} else {
		err := notifier.generateAndSendVideo(ctx, birthday)
		if err != nil {
			return err
		}
	}
	err := notifier.telegram.SendMessage(ctx, birthday.ChatId, fmt.Sprintf(BIRTHDAY_MESSAGE, birthday.Name))
	if err != nil {
		common.ErrorLogger.Printf("Could not send birthday message: %v\n", err)
	}
	return nil
}

func (notifier BirthdayNotifier) generateAndSendVideo(ctx context.Context, birthday Birthday) error {
	profilePictureFileIds, err := notifier.telegram.GetProfilePictureFileIds(ctx, birthday.UserId)
	if err != nil {
		return err
	}
	if len(profilePictureFileIds) > 0 {
		linkToProfilePicture, err := notifier.telegram.GetFileLink(ctx, profilePictureFileIds[0])
		if err != nil {
			return err
		}
		if len(linkToProfilePicture) == 0 {
			return nil
		}
		pathToImage, err := notifier.fileDownloader.Download(ctx, linkToProfilePicture)
		if err != nil {
			return err
		}
		pathToVideo, err := notifier.videoGenerator.CreateVideo(pathToImage)
		if err != nil {
			return err
		}
		fileId, err := notifier.telegram.SendVideo(ctx, birthday.ChatId, pathToVideo)
		if err != nil {
			return err
		}
		notifier.userIdToCachedVideoFileId[birthday.UserId] = fileId
	}
	return nil
}

const (
	BIRTHDAY_MESSAGE = "Aah %s\nHappy birthday, senpai! 🎂✨ I hope your day is as wonderful as you are!\n(⁄ ⁄>⁄ ▽ ⁄&lt;⁄ ⁄)♡"
)
