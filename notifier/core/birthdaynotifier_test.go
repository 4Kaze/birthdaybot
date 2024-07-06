package core_test

import (
	"context"
	"errors"
	"time"

	"github.com/4Kaze/birthdaybot/notifier/core"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Birthday notifier", func() {
	var repository FakeRepository
	var telegram FakeTelegram
	var clock FakeClock
	var scheduler FakeBirthdayScheduler
	var fileDownloader FakeFileDownloader
	var videoGenerator FakeVideoGenerator

	var notifier core.BirthdayNotifier

	BeforeEach(func() {
		repository = FakeRepository{}
		telegram = FakeTelegram{}
		scheduler = FakeBirthdayScheduler{}
		fileDownloader = FakeFileDownloader{}
		videoGenerator = FakeVideoGenerator{}
		notifier = *core.NewBirthdayNotifier(&repository, &telegram, &scheduler, &fileDownloader, &videoGenerator, &clock)
	})

	Describe("scheduling", func() {
		It("should schedule a birthday notification for multiple groups", func() {
			// given
			clock.now = NOW
			birthday1 := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			birthday2 := core.Birthday{
				ChatId: CHAT_ID_2,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			repository.thereAre(birthday1, birthday2)

			// when
			result := notifier.ScheduleBirthdayNotifications(context.Background(), SERVICE_URL)

			// then
			Expect(result).To(BeNil())
			Expect(repository.requestedDates).To(HaveExactElements(NOW))
			Expect(scheduler.scheduledTasks).To(ContainElements(ScheduledTask{birthday1, SERVICE_URL}, ScheduledTask{birthday2, SERVICE_URL}))

		})

		It("should schedule a birthday notification for multiple birthdays in a group", func() {
			// given
			clock.now = NOW
			birthday1 := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			birthday2 := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_2,
				Name:   USER_NAME_1,
			}
			repository.thereAre(birthday1, birthday2)

			// when
			result := notifier.ScheduleBirthdayNotifications(context.Background(), SERVICE_URL)

			// then
			Expect(result).To(BeNil())
			Expect(repository.requestedDates).To(HaveExactElements(NOW))
			Expect(scheduler.scheduledTasks).To(ContainElements(ScheduledTask{birthday1, SERVICE_URL}, ScheduledTask{birthday2, SERVICE_URL}))

		})

		It("should not schedule birthdays if there aren't any", func() {
			// given
			clock.now = NOW
			repository.thereAreNoBirthdays()

			// when
			result := notifier.ScheduleBirthdayNotifications(context.Background(), SERVICE_URL)

			// then
			Expect(result).To(BeNil())
			Expect(repository.requestedDates).To(HaveExactElements(NOW))
			Expect(scheduler.scheduledTasks).To(BeEmpty())
		})

		It("should return an error when fetching birthdays fails", func() {
			// given
			clock.now = NOW
			repository.shouldFail = true

			// when
			result := notifier.ScheduleBirthdayNotifications(context.Background(), SERVICE_URL)

			// then
			Expect(repository.requestedDates).To(HaveExactElements(NOW))
			Expect(result).To(Not(BeNil()))
		})
	})

	Describe("notifying", func() {
		It("should generate a video and send a birthday message for birthday", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			fileDownloader.filePathToReturn = PICTURE_PATH
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(BeNil())
			Expect(telegram.profilePictureRequests).To(HaveExactElements(USER_ID_1))
			Expect(telegram.fileLinkRequests).To(HaveExactElements(FILE_ID_1))
			Expect(fileDownloader.requests).To(HaveExactElements(FILE_LINK))
			Expect(videoGenerator.videoGenerationRequests).To(HaveExactElements(PICTURE_PATH))
			Expect(telegram.sentVideos).To(HaveExactElements(Video{chatId: CHAT_ID_1, path: VIDEO_PATH}))
			Expect(telegram.sentMessages).To(HaveExactElements(Message{chatId: CHAT_ID_1, text: EXPECTED_USER_1_BIRTHDAY_MESSAGE}))
		})

		It("should generate the video only for the first profile picture", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereAreProfilePictureFileIds(FILE_ID_1, FILE_ID_2)
			telegram.fileLinkToReturn = FILE_LINK
			fileDownloader.filePathToReturn = PICTURE_PATH
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(BeNil())
			Expect(telegram.profilePictureRequests).To(HaveExactElements(USER_ID_1))
			Expect(telegram.fileLinkRequests).To(HaveExactElements(FILE_ID_1))
			Expect(fileDownloader.requests).To(HaveExactElements(FILE_LINK))
			Expect(videoGenerator.videoGenerationRequests).To(HaveExactElements(PICTURE_PATH))
		})

		It("should not send a video when there is no profile picture", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereAreNoProfilePictures()

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(BeNil())
			Expect(telegram.profilePictureRequests).To(HaveExactElements(USER_ID_1))
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(HaveExactElements(Message{chatId: CHAT_ID_1, text: EXPECTED_USER_1_BIRTHDAY_MESSAGE}))
		})

		It("should not send a video when there is no link to profile picture file", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereAreProfilePictureFileIds(FILE_ID_1)
			telegram.fileLinkToReturn = ""

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(BeNil())
			Expect(telegram.profilePictureRequests).To(HaveExactElements(USER_ID_1))
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(HaveExactElements(Message{chatId: CHAT_ID_1, text: EXPECTED_USER_1_BIRTHDAY_MESSAGE}))
		})

		It("should return an error when fetching profile picture file id fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.shouldFailOnGettingProfilePicture = true

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(Not(BeNil()))
			Expect(videoGenerator.videoGenerationRequests).To(BeEmpty())
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should return an error when fetching profile picture file link fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			telegram.shouldFailOnGettingFile = true
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(Not(BeNil()))
			Expect(videoGenerator.videoGenerationRequests).To(BeEmpty())
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should return an error when downloading profile picture fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			fileDownloader.shouldFail = true
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(Not(BeNil()))
			Expect(videoGenerator.videoGenerationRequests).To(BeEmpty())
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should return an error when generating a video fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			videoGenerator.shouldFail = true

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(Not(BeNil()))
			Expect(telegram.sentVideos).To(BeEmpty())
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should return an error when sending a video fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			telegram.shouldFailOnSendingVideoFromPath = true
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(Not(BeNil()))
			Expect(telegram.sentMessages).To(BeEmpty())
		})

		It("should not return an error when sending a birthday message fails", func() {
			// given
			birthday := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			telegram.shouldFailOnSendingMessage = true
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			result := notifier.SendBirthdayNotification(context.Background(), birthday)

			// then
			Expect(result).To(BeNil())
			Expect(telegram.sentVideos).To(Not(BeEmpty()))
		})

		It("should reuse video's fileId when sending birthday notifications for the same user", func() {
			// given
			birthday1 := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			birthday2 := core.Birthday{
				ChatId: CHAT_ID_2,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			telegram.videoFileIdToReturn = FILE_ID_2
			videoGenerator.videoPathToReturn = VIDEO_PATH

			// when
			notifier.SendBirthdayNotification(context.Background(), birthday1)
			notifier.SendBirthdayNotification(context.Background(), birthday2)

			// then
			Expect(videoGenerator.videoGenerationRequests).To(HaveLen(1))
			Expect(telegram.sentVideos).To(HaveExactElements(
				Video{chatId: CHAT_ID_1, path: VIDEO_PATH},
				Video{chatId: CHAT_ID_2, path: FILE_ID_2},
			))
			Expect(telegram.sentMessages).To(HaveExactElements(
				Message{chatId: CHAT_ID_1, text: EXPECTED_USER_1_BIRTHDAY_MESSAGE},
				Message{chatId: CHAT_ID_2, text: EXPECTED_USER_1_BIRTHDAY_MESSAGE},
			))
		})

		It("should return an error when sending video from fileId fails", func() {
			// given
			birthday1 := core.Birthday{
				ChatId: CHAT_ID_1,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			birthday2 := core.Birthday{
				ChatId: CHAT_ID_2,
				UserId: USER_ID_1,
				Name:   USER_NAME_1,
			}
			telegram.thereIsProfilePicture(FILE_ID_1, FILE_LINK)
			telegram.videoFileIdToReturn = FILE_ID_2
			telegram.shouldFailOnSendingVideoFromFileId = true

			// when
			result1 := notifier.SendBirthdayNotification(context.Background(), birthday1)
			result2 := notifier.SendBirthdayNotification(context.Background(), birthday2)

			// then
			Expect(result1).To(BeNil())
			Expect(result2).To(Not(BeNil()))
			Expect(videoGenerator.videoGenerationRequests).To(HaveLen(1))
			Expect(telegram.sentMessages).To(HaveLen(1))
		})
	})
})

// ===== FAKES =====

type FakeRepository struct {
	birthdays      []core.Birthday
	requestedDates []time.Time
	shouldFail     bool
}

func (repository *FakeRepository) GetBirthdays(_ context.Context, date time.Time) ([]core.Birthday, error) {
	repository.requestedDates = append(repository.requestedDates, date)
	if repository.shouldFail {
		return nil, errors.New("test")
	}
	return repository.birthdays, nil
}

func (repository *FakeRepository) thereAre(birthday ...core.Birthday) {
	repository.birthdays = append(repository.birthdays, birthday...)
}

func (repository *FakeRepository) thereAreNoBirthdays() {
	repository.birthdays = make([]core.Birthday, 0)
}

type Message struct {
	chatId int64
	text   string
}

type Video struct {
	chatId int64
	path   string
}

type FakeTelegram struct {
	sentMessages                       []Message
	sentVideos                         []Video
	profilePictureRequests             []int64
	profilePictureFileIdsToReturn      []string
	fileLinkRequests                   []string
	fileLinkToReturn                   string
	videoFileIdToReturn                string
	shouldFailOnSendingMessage         bool
	shouldFailOnSendingVideoFromPath   bool
	shouldFailOnSendingVideoFromFileId bool
	shouldFailOnGettingProfilePicture  bool
	shouldFailOnGettingFile            bool
}

func (fake *FakeTelegram) SendMessage(_ context.Context, chatId int64, text string) error {
	if fake.shouldFailOnSendingMessage {
		return errors.New("test error")
	}
	fake.sentMessages = append(fake.sentMessages, Message{
		chatId: chatId,
		text:   text,
	})
	return nil
}

func (fake *FakeTelegram) SendVideo(_ context.Context, chatId int64, pathToVideo string) (fileId string, err error) {
	if fake.shouldFailOnSendingVideoFromPath {
		return "", errors.New("test error")
	}
	fake.sentVideos = append(fake.sentVideos, Video{
		chatId: chatId,
		path:   pathToVideo,
	})
	return fake.videoFileIdToReturn, nil
}

func (fake *FakeTelegram) SendVideoFromFileId(ctx context.Context, chatId int64, fileId string) error {
	if fake.shouldFailOnSendingVideoFromFileId {
		return errors.New("test error")
	}
	fake.sentVideos = append(fake.sentVideos, Video{
		chatId: chatId,
		path:   fileId,
	})
	return nil
}

func (fake *FakeTelegram) GetProfilePictureFileIds(_ context.Context, userId int64) ([]string, error) {
	if fake.shouldFailOnGettingProfilePicture {
		return nil, errors.New("test error")
	}
	fake.profilePictureRequests = append(fake.profilePictureRequests, userId)
	return fake.profilePictureFileIdsToReturn, nil
}

func (fake *FakeTelegram) GetFileLink(_ context.Context, fileId string) (string, error) {
	if fake.shouldFailOnGettingFile {
		return "", errors.New("test error")
	}
	fake.fileLinkRequests = append(fake.fileLinkRequests, fileId)
	return fake.fileLinkToReturn, nil
}

func (fake *FakeTelegram) thereAreProfilePictureFileIds(fileIds ...string) {
	fake.profilePictureFileIdsToReturn = fileIds

}

func (fake *FakeTelegram) thereIsProfilePicture(fileId string, fileLink string) {
	fake.profilePictureFileIdsToReturn = append(fake.profilePictureFileIdsToReturn, fileId)
	fake.fileLinkToReturn = fileLink
}

func (fake *FakeTelegram) thereAreNoProfilePictures() {
	fake.profilePictureFileIdsToReturn = nil
}

type FakeClock struct {
	now time.Time
}

func (clock FakeClock) Now() time.Time {
	return clock.now
}

type FakeFileDownloader struct {
	requests         []string
	filePathToReturn string
	shouldFail       bool
}

func (fake *FakeFileDownloader) Download(ctx context.Context, link string) (string, error) {
	if fake.shouldFail {
		return "", errors.New("test error")
	}
	fake.requests = append(fake.requests, link)
	return fake.filePathToReturn, nil
}

type FakeVideoGenerator struct {
	videoPathToReturn       string
	videoGenerationRequests []string
	shouldFail              bool
}

func (fake *FakeVideoGenerator) CreateVideo(linkToProfilePicture string) (string, error) {
	if fake.shouldFail {
		return "", errors.New("test error")
	}
	fake.videoGenerationRequests = append(fake.videoGenerationRequests, linkToProfilePicture)
	return fake.videoPathToReturn, nil
}

type FakeBirthdayScheduler struct {
	scheduledTasks []ScheduledTask
}

type ScheduledTask struct {
	birthday core.Birthday
	url      string
}

func (fake *FakeBirthdayScheduler) Schedule(ctx context.Context, birthday core.Birthday, serviceUrl string) {
	fake.scheduledTasks = append(fake.scheduledTasks, ScheduledTask{birthday, serviceUrl})
}

// ===== TEST DATA =====
const (
	CHAT_ID_1    int64 = 981
	CHAT_ID_2    int64 = 881
	USER_ID_1    int64 = 123
	USER_ID_2    int64 = 456
	USER_NAME_1        = "test 1"
	USER_NAME_2        = "test 2"
	FILE_ID_1          = "file_id_1"
	FILE_ID_2          = "file_id_2"
	FILE_LINK          = "some://file-link"
	PICTURE_PATH       = "/some/path"
	VIDEO_PATH         = "some/video.mp4"
	SERVICE_URL        = "http://this-service/test"

	EXPECTED_USER_1_BIRTHDAY_MESSAGE = "Aah test 1\nHappy birthday, senpai! 🎂✨ I hope your day is as wonderful as you are!\n(⁄ ⁄>⁄ ▽ ⁄&lt;⁄ ⁄)♡"
)

var NOW = time.Now()
