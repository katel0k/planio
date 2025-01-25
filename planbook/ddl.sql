START TRANSACTION ISOLATION LEVEL REPEATABLE READ;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    nickname VARCHAR(32) NOT NULL UNIQUE
);

CREATE TYPE time_scale AS ENUM ('undefined', 'year', 'month', 'week', 'day', 'hour');

CREATE TABLE plans (
    id SERIAL PRIMARY KEY,
    author_id INTEGER NOT NULL,
    synopsis VARCHAR(128) NOT NULL,
    creation_dttm TIMESTAMP DEFAULT NOW()::TIMESTAMP,
    parent_id INTEGER DEFAULT NULL,
    scale time_scale DEFAULT 'undefined'
);

ALTER TABLE plans ADD CONSTRAINT FK_users_plans FOREIGN KEY
    (author_id) REFERENCES users(id);

ALTER TABLE plans ADD CONSTRAINT FK_parent_plan FOREIGN KEY
    (parent_id) REFERENCES plans(id);

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

CREATE VIEW user_plans_roots
AS (
    SELECT author_id AS user_id, id AS plan_id
    FROM plans
    WHERE scale='undefined' AND parent_id IS NULL
);

COMMIT;
