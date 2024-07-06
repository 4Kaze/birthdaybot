CREATE TABLE IF NOT EXISTS birthdays
(
    chat_id              BIGINT NOT NULL,
    user_id              BIGINT NOT NULL,
    date                 DATE   NOT NULL,
    adjusted_day_of_year INT    NOT NULL,
    username             VARCHAR(32),
    first_name           VARCHAR(64),
    last_name            VARCHAR(64),
    PRIMARY KEY (chat_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_days ON birthdays (adjusted_day_of_year);

CREATE INDEX IF NOT EXISTS idx_user_ids ON birthdays (user_id);

CREATE INDEX IF NOT EXISTS idx_chat_adjusted_day_of_year ON birthdays (chat_id, adjusted_day_of_year);