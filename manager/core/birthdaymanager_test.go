package core_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/4Kaze/birthdaybot/manager/core"
	"github.com/go-telegram/bot/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Birthday manager", func() {
	var repository FakeRepository
	var telegram FakeTelegram
	var bot core.BirthdayManager

	BeforeEach(func() {
		repository = FakeRepository{}
		telegram = FakeTelegram{}
		bot = *core.NewBirthdayManager(&repository, &telegram, BOT_ID)
	})

	Describe("setting birthday", func() {
		DescribeTable("should save correct date", func(groupType string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID:        USER_ID_1,
							FirstName: FIRST_NAME_1,
							LastName:  LAST_NAME,
							Username:  USER_NAME_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: groupType,
						},
						Text: "/setbirthday 20.02",
					},
				},
			)

			Expect(repository.savedBirthdays).To(HaveExactElements(core.Birthday{
				ChatId:        CHAT_ID_1,
				UserId:        USER_ID_1,
				Date:          time.Date(DEFAULT_YEAR, 2, 20, 0, 0, 0, 0, time.UTC),
				UserFirstName: FIRST_NAME_1,
				UserLastName:  LAST_NAME,
				Username:      USER_NAME_1,
			}))
			Expect(telegram.sentReactions).To(HaveExactElements(Reaction{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				reaction:  "👍",
			}))
		},
			Entry("for supergroup", "supergroup"),
			Entry("for private group", "group"),
		)

		DescribeTable("should save non-standard format dates", func(dateString string, expectedDate time.Time) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID:        USER_ID_1,
							FirstName: FIRST_NAME_1,
							LastName:  LAST_NAME,
							Username:  USER_NAME_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: fmt.Sprintf("/setbirthday %s", dateString),
					},
				},
			)

			Expect(repository.savedBirthdays).To(HaveExactElements(core.Birthday{
				ChatId:        CHAT_ID_1,
				UserId:        USER_ID_1,
				Date:          expectedDate,
				UserFirstName: FIRST_NAME_1,
				UserLastName:  LAST_NAME,
				Username:      USER_NAME_1,
			}))
			Expect(telegram.sentReactions).To(HaveExactElements(Reaction{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				reaction:  "👍",
			}))
		},
			Entry("for day without leading zero", "1.01", monthAndDay(1, 1)),
			Entry("for month without leading zero", "01.1", monthAndDay(1, 1)),
			Entry("for '/' separator", "31/01", monthAndDay(1, 31)),
			Entry("for month in short-text format", "Jan 31", monthAndDay(1, 31)),
			Entry("for month in full-text format", "January 31", monthAndDay(1, 31)),
			Entry("for reverse order when not ambigous", "01.31", monthAndDay(1, 31)),
			Entry("for reverse order with different separator", "01/31", monthAndDay(1, 31)),
		)

		DescribeTable("should reply with a help message when date is incorrect", func(incorrectDate string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: fmt.Sprintf("/setbirthday %v", incorrectDate),
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_WRONG_FORMAT,
			}))

			Expect(repository.savedBirthdays).To(BeEmpty())
		},
			Entry("for non-complete date", "31"),
			Entry("for non-date string", "not-a-date"),
			Entry("for empty date string", ""),
		)

		It("should send an error reply when saving birthday fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/setbirthday 20.02",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_SAVE_FAILURE,
			}))
		})
	})

	Describe("getting own birthday", func() {
		DescribeTable("should reply with birthday date", func(groupType string, command string) {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date: time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
			})
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: groupType,
						},
						ReplyToMessage: nil,
						Text:           command,
					},
				},
			)

			Expect(repository.requestedBirthdays).To(HaveExactElements(RequestedBirthday{
				chatId: CHAT_ID_1,
				userId: USER_ID_1,
			}))
			Expect(telegram.sentReplies).To(HaveLen(1))
			Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
			Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("You were born on <b>January 31st</b>, right?"))
		},
			Entry("for supergroup with /getbirthday", "supergroup", "/getbirthday"),
			Entry("for private group with /getbirthday", "group", "/getbirthday"),
			Entry("for supergroup with /mybirthday", "supergroup", "/mybirthday"),
			Entry("for private group with /mybirthday", "group", "/mybirthday"),
		)

		It("should send an error reply when getting birthday fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						ReplyToMessage: nil,
						Text:           "/getbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_GET_FAILURE,
			}))
		})

		It("should send a reply when birthday is not set", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						ReplyToMessage: nil,
						Text:           "/getbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_NO_OWN_BIRTHDAY_SET,
			}))
		})
	})

	Describe("getting someone's birthday", func() {
		DescribeTable("should reply with birthday of an author of a message that was replied to", func(groupType string) {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date: time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
			})
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: groupType,
						},
						ReplyToMessage: &models.Message{
							From: &models.User{
								ID: USER_ID_2,
							},
						},
						Text: "/getbirthday",
					},
				},
			)

			Expect(repository.requestedBirthdays).To(HaveExactElements(RequestedBirthday{
				chatId: CHAT_ID_1,
				userId: USER_ID_2,
			}))
			Expect(telegram.sentReplies).To(HaveLen(1))
			Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
			Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("It's on <b>January 31st</b>"))
		},
			Entry("for supergroup", "supergroup"),
			Entry("for private group", "group"),
		)

		It("should send an error reply when getting birthday fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						ReplyToMessage: &models.Message{
							From: &models.User{
								ID: USER_ID_2,
							},
						},
						Text: "/getbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_GET_FAILURE,
			}))
		})

		It("should send a reply when birthday is not set", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						ReplyToMessage: &models.Message{
							From: &models.User{
								ID: USER_ID_2,
							},
						},
						Text: "/getbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_NO_BIRTHDAY_SET,
			}))
		})

		It("should not care about letter case", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/gEtBIRthdAy",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_NO_OWN_BIRTHDAY_SET,
			}))
		})

		It("should work for commands with @", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/getbirthday@thisbot",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_NO_OWN_BIRTHDAY_SET,
			}))
		})

		It("should reply when asking for birthday bot's birthday", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						ReplyToMessage: &models.Message{
							From: &models.User{
								ID: BOT_ID,
							},
						},
						Text: "/getbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_GET_BOT_BIRTHDAY,
			}))
		})
	})

	Describe("other messages", func() {
		DescribeTable("should not reply to private chat commands", func(command string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: command,
					},
				},
			)

			Expect(telegram.sentReplies).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		},
			Entry("/help", "/help"),
			Entry("/privacy", "/privacy"),
			Entry("/source", "/source"),
		)
	})

	Describe("getting next birthday", func() {
		DescribeTable("should reply with next birthday", func(groupType string) {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserId:        USER_ID_1,
				Username:      USER_NAME_1,
				UserFirstName: FIRST_NAME_1,
				UserLastName:  LAST_NAME,
			})
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: groupType,
						},
						Text: "/nextbirthday",
					},
				},
			)

			expectedMessage := fmt.Sprintf("It's our precious <a href=\"tg://user?id=%v\">%v %v</a>'s birthday on <b>January 31st</b>!", USER_ID_1, FIRST_NAME_1, LAST_NAME)
			Expect(repository.requestedNextBirthdayChatIds).To(HaveExactElements(CHAT_ID_1))
			Expect(telegram.sentReplies).To(HaveLen(1))
			Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
			Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring(expectedMessage))
		},
			Entry("for supergroup", "supergroup"),
			Entry("for private group", "group"),
		)

		DescribeTable("should reply with different names depending on saved data",
			func(username string, firstname string, lastname string, expectedName string) {
				_ = repository.SaveBirthday(context.Background(), core.Birthday{
					Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
					UserId:        123,
					Username:      username,
					UserFirstName: firstname,
					UserLastName:  lastname,
				})
				bot.HandleUpdate(
					context.Background(),
					&models.Update{
						Message: &models.Message{
							ID: MESSAGE_ID,
							From: &models.User{
								ID: USER_ID_1,
							},
							Chat: models.Chat{
								ID:   CHAT_ID_1,
								Type: "supergroup",
							},
							Text: "/nextbirthday",
						},
					},
				)

				expectedMessage := fmt.Sprintf("It's our precious %v's birthday on <b>January 31st</b>!", expectedName)
				Expect(repository.requestedNextBirthdayChatIds).To(HaveExactElements(CHAT_ID_1))
				Expect(telegram.sentReplies).To(HaveLen(1))
				Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
				Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
				Expect(telegram.sentReplies[0].text).To(ContainSubstring(expectedMessage))
			},
			Entry("only first name", "", "FIRST_NAME_1", "", "<a href=\"tg://user?id=123\">FIRST_NAME_1</a>"),
			Entry("first and last name", "", "FIRST_NAME_1", "LAST_NAME", "<a href=\"tg://user?id=123\">FIRST_NAME_1 LAST_NAME</a>"),
			Entry("only username", "USER_NAME_1", "", "", "@USER_NAME_1"),
		)

		It("should reply with two birthdays", func() {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserId:        1,
				Username:      "hackergirl",
				UserFirstName: "Iwakura",
				UserLastName:  "Lain",
			})
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserId:        2,
				UserFirstName: "Mizuki",
				UserLastName:  "Alice",
			})
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/nextbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveLen(1))
			Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
			Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("<b>2</b> people have their birthday on <b>January 31st</b>!"))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("It's <a href=\"tg://user?id=1\">Iwakura Lain</a> and <a href=\"tg://user?id=2\">Mizuki Alice</a>!"))
		})

		It("should reply with three birthdays", func() {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserId:        1,
				Username:      "hackergirl",
				UserFirstName: "Iwakura",
				UserLastName:  "Lain",
			})
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				Date:     time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserId:   2,
				Username: "test",
			})
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				UserId:        3,
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				UserFirstName: "Yan",
			})
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/nextbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveLen(1))
			Expect(telegram.sentReplies[0].chatId).To(Equal(CHAT_ID_1))
			Expect(telegram.sentReplies[0].messageId).To(Equal(MESSAGE_ID))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("<b>3</b> people have their birthday on <b>January 31st</b>!"))
			Expect(telegram.sentReplies[0].text).To(ContainSubstring("It's <a href=\"tg://user?id=1\">Iwakura Lain</a>, @test and <a href=\"tg://user?id=3\">Yan</a>!"))
		})

		It("should reply when there are no birthdays set", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/nextbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_NO_BIRTHDAYS,
			}))
		})

		It("should send an error reply when getting birthday fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/nextbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_GET_FAILURE,
			}))
		})
	})

	Describe("unsetting birthday by command", func() {
		DescribeTable("should delete birthday", func(groupType string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: groupType,
						},
						Text: "/unsetbirthday",
					},
				},
			)
			Expect(repository.deletedBirthdays).To(HaveExactElements(DeletedBirthday{
				userId: USER_ID_1,
				chatId: CHAT_ID_1,
			}))
			Expect(telegram.sentReactions).To(HaveExactElements(Reaction{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				reaction:  "👍",
			}))
		},
			Entry("for supergroup", "supergroup"),
			Entry("for private group", "group"),
		)

		It("should send a reply when failed to unset birthday", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						Text: "/unsetbirthday",
					},
				},
			)

			Expect(telegram.sentReplies).To(HaveExactElements(Reply{
				chatId:    CHAT_ID_1,
				messageId: MESSAGE_ID,
				text:      core.MESSAGE_UNSET_FAILURE,
			}))
		})
	})

	Describe("unsetting birthday of a leaving member", func() {
		It("should delete a birthday without sending any message", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						LeftChatMember: &models.User{
							ID: USER_ID_1,
						},
					},
				},
			)

			Expect(repository.deletedBirthdays).To(HaveExactElements(DeletedBirthday{
				userId: USER_ID_1,
				chatId: CHAT_ID_1,
			}))
			Expect(telegram.sentReplies).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should delete all birthdays when the bot is removed", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						LeftChatMember: &models.User{
							ID: BOT_ID,
						},
					},
				},
			)

			Expect(repository.deletedGroupBirthdays).To(HaveExactElements(CHAT_ID_1))
		})

		It("should not send any message when deleting fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						ID: MESSAGE_ID,
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "supergroup",
						},
						LeftChatMember: &models.User{
							ID: USER_ID_1,
						},
					},
				},
			)

			Expect(telegram.sentReplies).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})
	})

	Describe("removing all user data", func() {
		It("should remove user's data", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/clear all data",
					},
				},
			)

			Expect(repository.deletedUserBirthdays).To(HaveExactElements(USER_ID_1))
			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_DATA_CLEARED,
			}))
		})

		DescribeTable("should not remove user's data when command is not complete", func(command string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: command,
					},
				},
			)

			Expect(repository.deletedUserBirthdays).To(BeEmpty())
			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_WRONG_CLEAR_DATA_COMMAND,
			}))
		},
			Entry("no additional text", "/clear"),
			Entry("additional text not complete", "/clear all"),
			Entry("wrong additional text", "/clear something else"),
			Entry("too long additional text", "/clear all data and more"),
		)

		It("should send message when deleting data fails", func() {
			repository.shouldFail = true
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						From: &models.User{
							ID: USER_ID_1,
						},
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/clear all data",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_UNSET_FAILURE,
			}))
		})
	})

	Describe("private chat", func() {
		DescribeTable("should respond with help message", func(command string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: command,
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_FULL_HELP,
			}))
		},
			Entry("/help command", "/help"),
			Entry("/start command", "/start"),
		)

		It("should describe privacy policy", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/privacy",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_PRIVACY,
			}))
		})

		It("should respond with a link to the git repository", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/source",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_SOURCE,
			}))
		})

		It("should not care about letter case", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/sOUrcE",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_SOURCE,
			}))
		})

		It("should respond to commands with @", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/source@thisbot",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_SOURCE,
			}))
		})

		It("should respond with a short help when user sends an unknown command", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "/test",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_SHORT_HELP,
			}))
		})

		It("should respond with short help message when user doesn't send a command", func() {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: "blahblah",
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_SHORT_HELP,
			}))
		})

		DescribeTable("should respond when user sends a group chat command", func(command string) {
			bot.HandleUpdate(
				context.Background(),
				&models.Update{
					Message: &models.Message{
						Chat: models.Chat{
							ID:   CHAT_ID_1,
							Type: "private",
						},
						Text: command,
					},
				},
			)

			Expect(telegram.sentMessages).To(HaveExactElements(Message{
				chatId: CHAT_ID_1,
				text:   core.MESSAGE_GROUP_COMMAND,
			}))
		},
			Entry("set birthday", "/setbirthday"),
			Entry("unset birthday", "/unsetbirthday"),
			Entry("get birthday", "/getbirthday"),
			Entry("next birthday", "/nextbirthday"),
		)
	})

	Describe("getting birthday people", func() {
		It("should return birthday people for a given date", func() {
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				ChatId:        CHAT_ID_1,
				UserId:        USER_ID_1,
				Date:          time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				Username:      "hackergirl",
				UserFirstName: "Iwakura",
				UserLastName:  "Lain",
			})
			_ = repository.SaveBirthday(context.Background(), core.Birthday{
				ChatId:   CHAT_ID_2,
				UserId:   USER_ID_2,
				Date:     time.Date(DEFAULT_YEAR, 01, 31, 0, 0, 0, 0, time.UTC),
				Username: "mizukisan",
			})

			result, err := bot.GetBirthdays(context.Background(), NOW)

			Expect(repository.requestedBirthdaysForDates).To(HaveExactElements(NOW))
			Expect(err).To(BeNil())
			Expect(result).To(ConsistOf(
				core.BirthdayPerson{
					ChatId: CHAT_ID_1,
					UserId: USER_ID_1,
					Name:   fmt.Sprintf("<a href=\"tg://user?id=%v\">Iwakura Lain</a>", USER_ID_1),
				},
				core.BirthdayPerson{
					ChatId: CHAT_ID_2,
					UserId: USER_ID_2,
					Name:   "@mizukisan",
				},
			))
		})

		It("should return empty list when there are no birthdays", func() {
			result, err := bot.GetBirthdays(context.Background(), NOW)

			Expect(repository.requestedBirthdaysForDates).To(HaveExactElements(NOW))
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should pass error from repository", func() {
			repository.shouldFail = true

			result, err := bot.GetBirthdays(context.Background(), NOW)

			Expect(err).To(Not(BeNil()))
			Expect(result).To(BeNil())
		})
	})
})

// ===== FAKES =====

type DeletedBirthday struct {
	chatId int64
	userId int64
}

type RequestedBirthday struct {
	chatId int64
	userId int64
}

type FakeRepository struct {
	savedBirthdays               []core.Birthday
	deletedBirthdays             []DeletedBirthday
	deletedGroupBirthdays        []int64
	deletedUserBirthdays         []int64
	requestedBirthdays           []RequestedBirthday
	requestedNextBirthdayChatIds []int64
	requestedBirthdaysForDates   []time.Time
	shouldFail                   bool
}

func (repository *FakeRepository) SaveBirthday(_ context.Context, birthday core.Birthday) error {
	if repository.shouldFail {
		return errors.New("test")
	}
	repository.savedBirthdays = append(repository.savedBirthdays, birthday)
	return nil
}

func (repository *FakeRepository) GetBirthdayDate(ctx context.Context, chatId int64, userId int64) (*time.Time, error) {
	repository.requestedBirthdays = append(
		repository.requestedBirthdays,
		RequestedBirthday{chatId: chatId, userId: userId},
	)
	if repository.shouldFail {
		return nil, errors.New("test")
	}
	if len(repository.savedBirthdays) == 0 {
		return nil, nil
	}
	return &repository.savedBirthdays[0].Date, nil
}

func (repository *FakeRepository) GetNextBirthdays(ctx context.Context, chatId int64) ([]core.Birthday, error) {
	repository.requestedNextBirthdayChatIds = append(
		repository.requestedNextBirthdayChatIds,
		chatId,
	)
	if repository.shouldFail {
		return nil, errors.New("test")
	}
	return repository.savedBirthdays, nil
}

func (repository *FakeRepository) GetBirthdaysForDate(ctx context.Context, date time.Time) ([]core.Birthday, error) {
	repository.requestedBirthdaysForDates = append(repository.requestedBirthdaysForDates, date)
	if repository.shouldFail {
		return nil, errors.New("test")
	}
	return repository.savedBirthdays, nil
}

func (repository *FakeRepository) DeleteBirthday(_ context.Context, chatId int64, userId int64) error {
	if repository.shouldFail {
		return errors.New("test")
	}
	repository.deletedBirthdays = append(repository.deletedBirthdays, DeletedBirthday{chatId: chatId, userId: userId})
	return nil
}

func (repository *FakeRepository) DeleteAllChatBirthdays(_ context.Context, chatId int64) error {
	if repository.shouldFail {
		return errors.New("test")
	}
	repository.deletedGroupBirthdays = append(repository.deletedGroupBirthdays, chatId)
	return nil
}

func (repository *FakeRepository) DeleteAllUserBirthdays(_ context.Context, userId int64) error {
	if repository.shouldFail {
		return errors.New("test")
	}
	repository.deletedUserBirthdays = append(repository.deletedUserBirthdays, userId)
	return nil
}

type Message struct {
	chatId int64
	text   string
}

type Reply struct {
	chatId    int64
	messageId int
	text      string
}

type Reaction struct {
	chatId    int64
	messageId int
	reaction  string
}

type FakeTelegram struct {
	sentMessages  []Message
	sentReplies   []Reply
	sentReactions []Reaction
}

func (fake *FakeTelegram) SendReply(ctx context.Context, chatId int64, messageId int, text string) error {
	fake.sentReplies = append(fake.sentReplies, Reply{
		chatId:    chatId,
		messageId: messageId,
		text:      text,
	})
	return nil
}

func (fake *FakeTelegram) SendMessage(ctx context.Context, chatId int64, text string) error {
	fake.sentMessages = append(fake.sentMessages, Message{
		chatId: chatId,
		text:   text,
	})
	return nil
}

func (fake *FakeTelegram) SendReaction(ctx context.Context, chatId int64, messageId int, reaction string) error {
	fake.sentReactions = append(fake.sentReactions, Reaction{
		chatId:    chatId,
		messageId: messageId,
		reaction:  reaction,
	})
	return nil
}

// ===== TEST DATA =====
const (
	MESSAGE_ID   int   = 101
	CHAT_ID_1    int64 = 981
	CHAT_ID_2    int64 = 781
	USER_ID_1    int64 = 123
	USER_ID_2    int64 = 456
	BOT_ID       int64 = 666
	FIRST_NAME_1       = "Johnny"
	FIRST_NAME_2       = "Brad"
	LAST_NAME          = "Testowski"
	USER_NAME_1        = "test1"
	USER_NAME_2        = "test2"
	DEFAULT_YEAR       = 2000
)

var NOW = time.Now()

// ==== UTILS ====
func monthAndDay(month int, day int) time.Time {
	return time.Date(DEFAULT_YEAR, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
