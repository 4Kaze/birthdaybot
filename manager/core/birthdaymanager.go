package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/4Kaze/birthdaybot/common"
	"github.com/araddon/dateparse"
	"github.com/go-telegram/bot/models"
)

type BirthdayManager struct {
	repository Repository
	telegram   Telegram
	id         int64
}

type BirthdayPerson struct {
	ChatId int64
	UserId int64
	Name   string
}

const (
	CHAT_TYPE_PRIVATE    = "private"
	CHAT_TYPE_GROUP      = "group"
	CHAT_TYPE_SUPERGROUP = "supergroup"

	COMMAND_SET_BIRTHDAY   = "/setbirthday"
	COMMAND_UNSET_BIRTHDAY = "/unsetbirthday"
	COMMAND_GET_BIRTHDAY   = "/getbirthday"
	COMMAND_MY_BIRTHDAY    = "/mybirthday"
	COMMAND_NEXT_BIRTHDAY  = "/nextbirthday"
	COMMAND_START          = "/start"
	COMMAND_HELP           = "/help"
	COMMAND_PRIVACY        = "/privacy"
	COMMAND_SOURCE         = "/source"
	COMMAND_CLEAR          = "/clear"
	COMMAND_CLEAR_FULL     = "/clear all data"
	REACTION_THUMBS_UP     = "👍"

	DEFAULT_YEAR       = 2000
	INPUT_DATE_LAYOUT  = "2.1"
	OUTPUT_DATE_LAYOUT = "January 2"
)

func NewBirthdayManager(repository Repository, telegram Telegram, botId int64) *BirthdayManager {
	return &BirthdayManager{repository: repository, telegram: telegram, id: botId}
}

func (birthdayBot *BirthdayManager) HandleUpdate(ctx context.Context, update *models.Update) error {
	if isGroupUpdate(update) {
		if isCommand(update) {
			return birthdayBot.handleGroupCommand(ctx, update)
		}
		if hasLeftMember(update) {
			return birthdayBot.handleMemberLeaving(ctx, update)
		}
	} else if isPrivateChatUpdate(update) {
		if isCommand(update) {
			return birthdayBot.handlePrivateChatCommand(ctx, update)
		} else {
			return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_SHORT_HELP)
		}
	}
	return nil
}

func (birthdayBot *BirthdayManager) GetBirthdays(ctx context.Context, date time.Time) ([]BirthdayPerson, error) {
	birthdays, err := birthdayBot.repository.GetBirthdaysForDate(ctx, date)
	if err != nil {
		return nil, err
	}
	birthdayPeople := make([]BirthdayPerson, len(birthdays))
	for index, birthday := range birthdays {
		birthdayPeople[index] = BirthdayPerson{
			ChatId: birthday.ChatId,
			UserId: birthday.UserId,
			Name:   createBirthdayPersonName(birthday),
		}
	}
	return birthdayPeople, nil
}

func isGroupUpdate(update *models.Update) bool {
	return update.Message != nil &&
		(update.Message.Chat.Type == CHAT_TYPE_GROUP || update.Message.Chat.Type == CHAT_TYPE_SUPERGROUP)
}

func isPrivateChatUpdate(update *models.Update) bool {
	return update.Message != nil && update.Message.Chat.Type == CHAT_TYPE_PRIVATE
}

func isCommand(update *models.Update) bool {
	return update.Message != nil && strings.HasPrefix(update.Message.Text, "/")
}

func hasLeftMember(update *models.Update) bool {
	return update.Message != nil && update.Message.LeftChatMember != nil
}

func extractCommand(text string) string {
	command, _, _ := strings.Cut(text, " ")
	commandWithoutAt, _, _ := strings.Cut(command, "@")
	return strings.ToLower(commandWithoutAt)
}

func (birthdayBot *BirthdayManager) handleGroupCommand(ctx context.Context, update *models.Update) error {
	command := extractCommand(update.Message.Text)
	switch command {
	case COMMAND_SET_BIRTHDAY:
		return birthdayBot.saveBirthday(ctx, update)
	case COMMAND_UNSET_BIRTHDAY:
		return birthdayBot.deleteBirthday(ctx, update)
	case COMMAND_GET_BIRTHDAY, COMMAND_MY_BIRTHDAY:
		return birthdayBot.getBirthday(ctx, update)
	case COMMAND_NEXT_BIRTHDAY:
		return birthdayBot.getNextBirthday(ctx, update)
	}
	return nil
}

func (birthdayBot *BirthdayManager) handlePrivateChatCommand(ctx context.Context, update *models.Update) error {
	command := extractCommand(update.Message.Text)
	switch command {
	case COMMAND_HELP, COMMAND_START:
		return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_FULL_HELP)
	case COMMAND_PRIVACY:
		return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_PRIVACY)
	case COMMAND_SOURCE:
		return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_SOURCE)
	case COMMAND_CLEAR:
		return birthdayBot.deleteAllUserBirthdays(ctx, update)
	case COMMAND_SET_BIRTHDAY, COMMAND_UNSET_BIRTHDAY, COMMAND_GET_BIRTHDAY, COMMAND_NEXT_BIRTHDAY:
		return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_GROUP_COMMAND)
	default:
		return birthdayBot.sendPrivateChatMessage(ctx, update, MESSAGE_SHORT_HELP)
	}
}

func (birthdayBot *BirthdayManager) handleMemberLeaving(ctx context.Context, update *models.Update) error {
	memberThatLeft := update.Message.LeftChatMember
	chatId := update.Message.Chat.ID

	if memberThatLeft.ID == birthdayBot.id {
		err := birthdayBot.repository.DeleteAllChatBirthdays(ctx, chatId)
		if err != nil {
			return fmt.Errorf("could not delete all chat birthdays from the database due to: %v", err)
		}
	} else {
		err := birthdayBot.repository.DeleteBirthday(ctx, chatId, memberThatLeft.ID)
		if err != nil {
			return fmt.Errorf("could not delete birthday from the database due to: %v", err)
		}
	}
	return nil
}

func (birthdayBot *BirthdayManager) sendPrivateChatMessage(ctx context.Context, update *models.Update, message string) error {
	chatId := update.Message.Chat.ID
	return birthdayBot.telegram.SendMessage(ctx, chatId, message)
}

func (birthdayBot *BirthdayManager) saveBirthday(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	messageId := update.Message.ID
	userId := update.Message.From.ID
	firstName := update.Message.From.FirstName
	lastName := update.Message.From.LastName
	userName := update.Message.From.Username

	messagesParts := strings.Fields(update.Message.Text)
	if len(messagesParts) < 2 {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_WRONG_FORMAT)
	}

	date, err := parseDate(messagesParts[1:])
	if err != nil {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_WRONG_FORMAT)
	}

	err = birthdayBot.repository.SaveBirthday(ctx, Birthday{
		Date:          date,
		ChatId:        chatId,
		UserId:        userId,
		Username:      userName,
		UserFirstName: firstName,
		UserLastName:  lastName,
	})
	if err != nil {
		common.ErrorLogger.Printf("could not save birthday (%v) to the database due to: %v\n", date, err)
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_SAVE_FAILURE)
	}

	return birthdayBot.telegram.SendReaction(ctx, chatId, messageId, REACTION_THUMBS_UP)
}

func parseDate(parts []string) (time.Time, error) {
	date, err := time.Parse(INPUT_DATE_LAYOUT, parts[0])
	if err == nil {
		return sanitizeDate(date), nil
	}
	datestr := strings.Join(parts, " ")
	parsedDate, err := dateparse.ParseAny(datestr, dateparse.RetryAmbiguousDateWithSwap(true), dateparse.PreferMonthFirst(false))
	if err != nil {
		return time.Time{}, err
	}
	return sanitizeDate(parsedDate), nil
}

func sanitizeDate(date time.Time) time.Time {
	return time.Date(DEFAULT_YEAR, date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
}

func (birthdayBot *BirthdayManager) getBirthday(ctx context.Context, update *models.Update) error {
	if update.Message.ReplyToMessage == nil {
		return birthdayBot.getOwnBirthday(ctx, update)
	}
	return birthdayBot.getSomeonesBirthday(ctx, update)
}

func (birthdayBot *BirthdayManager) getOwnBirthday(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	userId := update.Message.From.ID
	messageId := update.Message.ID

	date, err := birthdayBot.repository.GetBirthdayDate(ctx, chatId, userId)
	if err != nil {
		common.ErrorLogger.Printf("could not get birthday from the database due to: %v\n", err)
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_GET_FAILURE)
	}
	if date == nil {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_NO_OWN_BIRTHDAY_SET)
	}

	return birthdayBot.telegram.SendReply(ctx, chatId, messageId, fmt.Sprintf(MESSAGE_GET_OWN_BIRTHDAY, formatDateForOutput(*date)))
}

func (birthdayBot *BirthdayManager) getSomeonesBirthday(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	messageId := update.Message.ID
	subjectUserId := update.Message.ReplyToMessage.From.ID

	if subjectUserId == birthdayBot.id {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_GET_BOT_BIRTHDAY)
	}

	date, err := birthdayBot.repository.GetBirthdayDate(ctx, chatId, subjectUserId)
	if err != nil {
		common.ErrorLogger.Printf("could not get birthday from the database due to: %v\n", err)
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_GET_FAILURE)
	}
	if date == nil {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_NO_BIRTHDAY_SET)
	}

	return birthdayBot.telegram.SendReply(ctx, chatId, messageId, fmt.Sprintf(MESSAGE_GET_BIRTHDAY, formatDateForOutput(*date)))
}

func (birthdayBot *BirthdayManager) getNextBirthday(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	messageId := update.Message.ID

	birthdays, err := birthdayBot.repository.GetNextBirthdays(ctx, chatId)
	if err != nil {
		common.ErrorLogger.Printf("could not get next birthday from the database due to: %v\n", err)
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_GET_FAILURE)
	}
	if len(birthdays) == 0 {
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_NO_BIRTHDAYS)
	}

	message := createNextBirthdayMessage(birthdays)
	return birthdayBot.telegram.SendReply(ctx, chatId, messageId, message)
}

func createNextBirthdayMessage(birthdays []Birthday) string {
	name := createBirthdayPersonName(birthdays[0])
	if len(birthdays) == 1 {
		return fmt.Sprintf(MESSAGE_NEXT_BIRTHDAY, name, formatDateForOutput(birthdays[0].Date))
	}
	birthdayNames := name
	for index, birthday := range birthdays[1:] {
		name := createBirthdayPersonName(birthday)
		if index == len(birthdays)-2 {
			birthdayNames = fmt.Sprintf("%v and %v", birthdayNames, name)
		} else {
			birthdayNames = fmt.Sprintf("%v, %v", birthdayNames, name)
		}
	}
	return fmt.Sprintf(MESSAGE_NEXT_BIRTHDAYS, len(birthdays), formatDateForOutput(birthdays[0].Date), birthdayNames)
}

func createBirthdayPersonName(birthday Birthday) string {
	if len(birthday.UserFirstName) == 0 && len(birthday.UserLastName) == 0 {
		return fmt.Sprintf("@%v", birthday.Username)
	}
	name := birthday.UserFirstName
	if len(birthday.UserLastName) > 0 {
		name = fmt.Sprintf("%v %v", name, birthday.UserLastName)
	}
	return fmt.Sprintf("<a href=\"tg://user?id=%v\">%v</a>", birthday.UserId, name)
}

func (birthdayBot *BirthdayManager) deleteBirthday(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	userId := update.Message.From.ID
	messageId := update.Message.ID

	err := birthdayBot.repository.DeleteBirthday(ctx, chatId, userId)
	if err != nil {
		common.ErrorLogger.Printf("could not delete birthday from the database due to: %v\n", err)
		return birthdayBot.telegram.SendReply(ctx, chatId, messageId, MESSAGE_UNSET_FAILURE)
	}

	return birthdayBot.telegram.SendReaction(ctx, chatId, messageId, REACTION_THUMBS_UP)
}

func (birthdayBot *BirthdayManager) deleteAllUserBirthdays(ctx context.Context, update *models.Update) error {
	chatId := update.Message.Chat.ID
	userId := update.Message.From.ID

	if update.Message.Text != COMMAND_CLEAR_FULL {
		return birthdayBot.telegram.SendMessage(ctx, chatId, MESSAGE_WRONG_CLEAR_DATA_COMMAND)
	}

	err := birthdayBot.repository.DeleteAllUserBirthdays(ctx, userId)
	if err != nil {
		common.ErrorLogger.Printf("could not delete birthdays from the database due to: %v\n", err)
		return birthdayBot.telegram.SendMessage(ctx, chatId, MESSAGE_UNSET_FAILURE)
	}

	return birthdayBot.telegram.SendMessage(ctx, chatId, MESSAGE_DATA_CLEARED)
}

func formatDateForOutput(date time.Time) string {
	preformattedDate := date.Format(OUTPUT_DATE_LAYOUT)
	suffix := getDateSuffix(date)
	return preformattedDate + suffix
}

func getDateSuffix(date time.Time) string {
	day := date.Day()
	if day >= 11 && day <= 13 {
		return "th"
	}
	switch day % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
