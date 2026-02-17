
-- Users 
CREATE TABLE IF NOT EXISTS "user" (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS user_password (
    user_id TEXT PRIMARY KEY,
    password_hash TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_password_updated ON user_password(updated_at);

-- vote session
CREATE TABLE IF NOT EXISTS vote_session (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    ends_at TEXT
);

CREATE TABLE IF NOT EXISTS session_and_participant (
    user_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    invited_at TEXT NOT NULL,
    PRIMARY KEY (user_id, session_id)
);

CREATE INDEX IF NOT EXISTS idx_participant_user ON session_and_participant(user_id);
CREATE INDEX IF NOT EXISTS idx_participant_session ON session_and_participant(session_id);

-- Questions and choices
-- TODO unique id ? indepotent
CREATE TABLE IF NOT EXISTS question (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    text TEXT NOT NULL,
    order_num INTEGER NOT NULL,
    allow_multiple INTEGER NOT NULL DEFAULT 0,
    max_choices INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (session_id, order_num)
);

CREATE INDEX IF NOT EXISTS idx_question_session ON question(session_id);

CREATE TABLE IF NOT EXISTS choice (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    order_num INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (question_id, order_num),
    FOREIGN KEY (question_id) REFERENCES question(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_choice_question ON choice(question_id);

-- Votes
CREATE TABLE IF NOT EXISTS vote (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    question_id INTEGER NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (user_id, question_id)
);

CREATE INDEX IF NOT EXISTS idx_vote_user ON vote(user_id);
CREATE INDEX IF NOT EXISTS idx_vote_session ON vote(session_id);
CREATE INDEX IF NOT EXISTS idx_vote_question ON vote(question_id);

CREATE TABLE IF NOT EXISTS vote_and_choice (
    vote_id TEXT NOT NULL,
    choice_id INTEGER NOT NULL,
    PRIMARY KEY (vote_id, choice_id),
    FOREIGN KEY (vote_id) REFERENCES vote(id) ON DELETE CASCADE,
    FOREIGN KEY (choice_id) REFERENCES choice(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_vote_choice_vote ON vote_and_choice(vote_id);
CREATE INDEX IF NOT EXISTS idx_vote_choice_choice ON vote_and_choice(choice_id);

-- User history and results
CREATE TABLE IF NOT EXISTS user_history (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    session_id TEXT,
    version INTEGER NOT NULL,
    string_size INTEGER NOT NULL,
    receipt_data BLOB NOT NULL,
    checksum TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (user_id, session_id),
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES vote_session(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_user_history_user ON user_history(user_id);
CREATE INDEX IF NOT EXISTS idx_user_history_session ON user_history(session_id);
CREATE INDEX IF NOT EXISTS idx_user_history_checksum ON user_history(checksum);

CREATE TABLE IF NOT EXISTS result_history (
    id TEXT PRIMARY KEY,
    session_id TEXT,
    version INTEGER NOT NULL,
    string_size INTEGER NOT NULL,
    result_data BLOB NOT NULL,
    checksum TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (session_id),
    FOREIGN KEY (session_id) REFERENCES vote_session(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_result_history_session ON result_history(session_id);
CREATE INDEX IF NOT EXISTS idx_result_history_checksum ON result_history(checksum);