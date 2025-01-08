START TRANSACTION ISOLATION LEVEL REPEATABLE READ;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nickname VARCHAR(32) NOT NULL UNIQUE
);

CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    author_id INTEGER NOT NULL,
    synopsis VARCHAR(128) NOT NULL,
    creation_dttm TIMESTAMP DEFAULT NOW()::TIMESTAMP
);

ALTER TABLE plans ADD CONSTRAINT FK_users_plans FOREIGN KEY
    (author_id) REFERENCES users(id);

CREATE TABLE descriptions (
    plan_id INTEGER PRIMARY KEY,
    body VARCHAR(1024)
);

ALTER TABLE descriptions ADD CONSTRAINT FK_description_for_plan FOREIGN KEY
    (plan_id) REFERENCES plans(id);

CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    author_id INTEGER,
    receiver_id INTEGER,
    sent_dttm TIMESTAMP DEFAULT NOW()::TIMESTAMP,
    body VARCHAR(1024)
);

ALTER TABLE messages ADD CONSTRAINT FK_users_messages FOREIGN KEY
    (author_id) REFERENCES users(id);
ALTER TABLE messages ADD CONSTRAINT FK_messages_to_user FOREIGN KEY
    (receiver_id) REFERENCES users(id);

COMMIT;
