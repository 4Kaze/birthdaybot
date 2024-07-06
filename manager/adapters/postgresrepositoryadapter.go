package adapters

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/4Kaze/birthdaybot/common"
	birthday_bot "github.com/4Kaze/birthdaybot/manager/core"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepositoryAdapter struct {
	database *pgxpool.Pool
	clock    *common.SystemClock
}

func NewSqlRepositoryAdapter(database *pgxpool.Pool, clock *common.SystemClock) *PostgresRepositoryAdapter {
	return &PostgresRepositoryAdapter{database: database, clock: clock}
}

func (adapter *PostgresRepositoryAdapter) SaveBirthday(ctx context.Context, birthday birthday_bot.Birthday) error {
	log.Printf("Inserting birthday into the database: %v\n", birthday)
	statement := `INSERT INTO birthdays (chat_id, user_id, date, adjusted_day_of_year, username, first_name, last_name)
						VALUES ($1, $2, $3, $4, $5, $6, $7)
						ON CONFLICT (chat_id, user_id) DO UPDATE SET
						date = $3, adjusted_day_of_year = $4, username = $5, first_name = $6, last_name = $7`
	if _, err := adapter.database.Exec(
		ctx,
		statement,
		birthday.ChatId,
		birthday.UserId,
		birthday.Date,
		getAdjustedDayOfYear(birthday.Date),
		birthday.Username,
		birthday.UserFirstName,
		birthday.UserLastName,
	); err != nil {
		common.ErrorLogger.Printf("Failed to insert a birthday: %v into the database: %v\n", birthday, err)
		return err
	}
	return nil
}

func (adapter *PostgresRepositoryAdapter) GetBirthdayDate(ctx context.Context, chatId int64, userId int64) (*time.Time, error) {
	log.Printf("Getting birthday from the database for chatId: %v, userId: %v\n", chatId, userId)
	statement := `SELECT date FROM birthdays WHERE chat_id = $1 AND user_id = $2`
	result := adapter.database.QueryRow(ctx, statement, chatId, userId)

	date := time.Time{}
	if err := result.Scan(&date); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		common.ErrorLogger.Printf("Failed to get a birthday date for chat: %v, userId: %v from the database: %v\n", chatId, userId, err)
		return nil, err
	}
	return &date, nil
}

func (adapter *PostgresRepositoryAdapter) GetNextBirthdays(ctx context.Context, chatId int64) ([]birthday_bot.Birthday, error) {
	currentAdjustedDayOfYear := getAdjustedDayOfYear(adapter.clock.Now())
	var birthdaysThisYear []birthday_bot.Birthday
	var err error
	birthdaysThisYear, err = adapter.getClosestBirthdaysAfterAdjustedDay(ctx, chatId, currentAdjustedDayOfYear)
	if err != nil {
		return nil, err
	}
	if len(birthdaysThisYear) == 0 {
		return adapter.getClosestBirthdaysAfterAdjustedDay(ctx, chatId, 0)
	}
	return birthdaysThisYear, nil
}

func (adapter *PostgresRepositoryAdapter) GetBirthdaysForDate(ctx context.Context, date time.Time) ([]birthday_bot.Birthday, error) {
	log.Printf("Getting birthdays from the database for date: %v\n", date)
	statement := `SELECT chat_id, user_id, date, username, first_name, last_name
					FROM birthdays
					WHERE adjusted_day_of_year = $1`
	adjustedDayOfYear := getAdjustedDayOfYear(date)
	var rows pgx.Rows
	var err error
	if rows, err = adapter.database.Query(ctx, statement, adjustedDayOfYear); err != nil {
		common.ErrorLogger.Printf("Failed to get birthdays for date: %v from the database: %v\n", date, err)
		return nil, err
	}
	var birthdays []birthday_bot.Birthday
	for rows.Next() {
		var birthday birthday_bot.Birthday
		if err = rows.Scan(
			&birthday.ChatId,
			&birthday.UserId,
			&birthday.Date,
			&birthday.Username,
			&birthday.UserFirstName,
			&birthday.UserLastName,
		); err != nil {
			common.ErrorLogger.Printf("Failed to scan rows for birthdays for date: %v due to: %v\n", date, err)
			return birthdays, err
		}
		birthdays = append(birthdays, birthday)
	}
	return birthdays, nil
}

func (adapter *PostgresRepositoryAdapter) DeleteBirthday(ctx context.Context, chatId int64, userId int64) error {
	log.Printf("Deleting birthday from the database for chatId: %v, userId: %v\n", chatId, userId)
	statement := `DELETE FROM birthdays WHERE chat_id = $1 AND user_id = $2`
	if _, err := adapter.database.Exec(ctx, statement, chatId, userId); err != nil {
		common.ErrorLogger.Printf("Failed to delete birthday for chatId: %v, userId: %v from the database: %v\n", chatId, userId, err)
		return err
	}
	return nil
}

func (adapter *PostgresRepositoryAdapter) DeleteAllChatBirthdays(ctx context.Context, chatId int64) error {
	log.Printf("Deleting birthdays from the database for chatId: %v\n", chatId)
	statement := `DELETE FROM birthdays WHERE chat_id = $1`
	if _, err := adapter.database.Exec(ctx, statement, chatId); err != nil {
		common.ErrorLogger.Printf("Failed to delete all birthdays for chatId: %v from the database: %v\n", chatId, err)
		return err
	}
	return nil
}

func (adapter *PostgresRepositoryAdapter) DeleteAllUserBirthdays(ctx context.Context, userId int64) error {
	log.Printf("Deleting birthday from the database for userId: %v\n", userId)
	statement := `DELETE FROM birthdays WHERE user_id = $1`
	if _, err := adapter.database.Exec(ctx, statement, userId); err != nil {
		common.ErrorLogger.Printf("Failed to delete all birthdays for userId: %v from the database: %v\n", userId, err)
		return err
	}
	return nil
}

func (adapter *PostgresRepositoryAdapter) getClosestBirthdaysAfterAdjustedDay(ctx context.Context, chatId int64, day int) ([]birthday_bot.Birthday, error) {
	log.Printf("Getting closest birthdays from the database for chatId: %v, day: %v\n", chatId, day)
	statement := `WITH closest_birthday AS (
						SELECT adjusted_day_of_year
						FROM birthdays
						WHERE chat_id = $1 AND adjusted_day_of_year > $2 ORDER BY adjusted_day_of_year LIMIT 1)
				SELECT chat_id, user_id, date, username, first_name, last_name
				FROM birthdays
				WHERE chat_id = $1 AND adjusted_day_of_year = (SELECT adjusted_day_of_year FROM closest_birthday)`
	var rows pgx.Rows
	var err error
	if rows, err = adapter.database.Query(ctx, statement, chatId, day); err != nil {
		common.ErrorLogger.Printf("Failed to get closest birthdays for chat: %v, day: %v from the database: %v\n", chatId, day, err)
		return nil, err
	}
	var birthdays []birthday_bot.Birthday
	for rows.Next() {
		var birthday birthday_bot.Birthday
		if err = rows.Scan(
			&birthday.ChatId,
			&birthday.UserId,
			&birthday.Date,
			&birthday.Username,
			&birthday.UserFirstName,
			&birthday.UserLastName,
		); err != nil {
			common.ErrorLogger.Printf("Failed to scan rows for closest birthdays for chat: %v, day: %v due to: %v\n", chatId, day, err)
			return birthdays, err
		}
		birthdays = append(birthdays, birthday)
	}
	return birthdays, nil
}

func getAdjustedDayOfYear(date time.Time) int {
	yearDay := date.YearDay()
	if !isLeapYear(date.Year()) && yearDay > FEBRUARY_28TH_YEAR_DAY {
		return yearDay + 1
	}
	return yearDay
}

func isLeapYear(year int) bool {
	return year%4 == 0 && year%100 != 0 || year%400 == 0
}

const FEBRUARY_28TH_YEAR_DAY = 59
