package adapters

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/4Kaze/birthdaybot/common"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBotWrapper struct {
	bot *telegram.Bot
}

func NewTelegramWrapper(bot *telegram.Bot) *TelegramBotWrapper {
	return &TelegramBotWrapper{
		bot: bot,
	}
}

func (wrapper *TelegramBotWrapper) SendMessage(ctx context.Context, chatId int64, text string) error {
	log.Printf("Sending message to chatId: %v, text: %v\n", chatId, text)
	_, err := wrapper.bot.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID:    chatId,
		Text:      text,
		ParseMode: models.ParseModeHTML,
	})
	if err != nil {
		common.ErrorLogger.Printf("Failed to send message: %v to chatId: %v due to: %v\n", text, chatId, err)
	}
	return err
}

func (wrapper *TelegramBotWrapper) SendVideo(ctx context.Context, chatId int64, pathToVideo string) (fileId string, err error) {
	log.Printf("Sending video to chatId: %v, path: %v\n", chatId, pathToVideo)
	videoFile, err := os.Open(pathToVideo)
	if err != nil {
		common.ErrorLogger.Printf("Failed to open video file: %v due to: %v\n", pathToVideo, err)
		return "", err
	}
	defer videoFile.Close()
	sentVideo, err := wrapper.bot.SendVideo(ctx, &telegram.SendVideoParams{
		ChatID: chatId,
		Video: &models.InputFileUpload{
			Filename: "happy birthday!.mp4",
			Data:     videoFile,
		},
	})
	if err != nil {
		common.ErrorLogger.Printf("Failed to send video to chatId: %v due to: %v\n", chatId, err)
		return "", err
	}
	return sentVideo.Video.FileID, nil
}

func (wrapper *TelegramBotWrapper) SendVideoFromFileId(ctx context.Context, chatId int64, fileId string) error {
	log.Printf("Sending video to chatId: %v, fileId: %v\n", chatId, fileId)
	_, err := wrapper.bot.SendVideo(ctx, &telegram.SendVideoParams{
		ChatID: chatId,
		Video: &models.InputFileString{
			Data: fileId,
		},
	})
	if err != nil {
		common.ErrorLogger.Printf("Failed to send video to chatId: %v from fileID: %v due to: %v\n", chatId, fileId, err)
		return err
	}
	return nil
}

func (wrapper *TelegramBotWrapper) GetProfilePictureFileIds(ctx context.Context, userId int64) ([]string, error) {
	log.Printf("Getting profile pictures for userId: %v\n", userId)
	photos, err := wrapper.bot.GetUserProfilePhotos(ctx, &telegram.GetUserProfilePhotosParams{
		UserID: userId,
		Limit:  1,
	})
	if err != nil {
		if errors.Is(err, telegram.ErrorBadRequest) {
			return nil, nil
		}
		common.ErrorLogger.Printf("Failed to get profile photos for userId: %v due to: %v\n", userId, err)
		return nil, err
	}

	if len(photos.Photos) == 0 || len(photos.Photos[0]) == 0 {
		log.Printf("UserId: %v has no photos. The video will not be generated.\n", userId)
		return nil, nil
	}

	fileId := photos.Photos[0][0].FileID
	return []string{fileId}, nil
}

func (wrapper *TelegramBotWrapper) GetFileLink(ctx context.Context, fileId string) (string, error) {
	log.Printf("Getting a link to a fileID: %v\n", fileId)
	file, err := wrapper.bot.GetFile(ctx, &telegram.GetFileParams{
		FileID: fileId,
	})
	if err != nil {
		if errors.Is(err, telegram.ErrorBadRequest) {
			return "", nil
		}
		common.ErrorLogger.Printf("Failed to get link to fileId: %v due to: %v\n", fileId, err)
		return "", err
	}

	link := wrapper.bot.FileDownloadLink(file)
	return link, nil
}
