package adapters

import (
	"context"
	"log"

	"github.com/4Kaze/birthdaybot/common"
	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBotWrapper struct {
	bot   *telegram.Bot
	BotId int64
}

func NewTelegramWrapper(ctx context.Context, bot *telegram.Bot) *TelegramBotWrapper {
	botUser, err := bot.GetMe(ctx)
	if err != nil {
		panic(err)
	}
	return &TelegramBotWrapper{
		bot:   bot,
		BotId: botUser.ID,
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

func (wrapper *TelegramBotWrapper) SendReply(ctx context.Context, chatId int64, messageId int, text string) error {
	log.Printf("Sending replay to chatId: %v, messageId: %v, text: %v\n", chatId, messageId, text)
	_, err := wrapper.bot.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID:    chatId,
		Text:      text,
		ParseMode: models.ParseModeHTML,
		ReplyParameters: &models.ReplyParameters{
			ChatID:    chatId,
			MessageID: messageId,
		},
	})
	if err != nil {
		common.ErrorLogger.Printf("Failed to send reply: %v to messageId: %v in chatId: %v due to: %v\n", text, messageId, chatId, err)
	}
	return err
}

func (wrapper *TelegramBotWrapper) SendReaction(ctx context.Context, chatId int64, messageId int, reaction string) error {
	log.Printf("Sending reaction to chatId: %v, messageId: %v, reaction: %v\n", chatId, messageId, reaction)
	_, err := wrapper.bot.SetMessageReaction(ctx, &telegram.SetMessageReactionParams{
		ChatID:    chatId,
		MessageID: messageId,
		Reaction: []models.ReactionType{
			{
				Type: models.ReactionTypeTypeEmoji,
				ReactionTypeEmoji: &models.ReactionTypeEmoji{
					Emoji: reaction,
				},
			},
		},
	})
	if err != nil {
		common.ErrorLogger.Printf("Failed to send reaction: %v to messageId: %v in chatId: %v due to: %v\n", reaction, messageId, chatId, err)
	}
	return err
}
