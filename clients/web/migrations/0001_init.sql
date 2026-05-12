-- Slycrel D1 schema. Characters + shared world state + auth sessions.
-- The "shared world" pieces (inn, gladiator fights/bets) match the original
-- BBS feel: anyone logged in can mutate any row. The cheat resistance bar
-- is intentionally low — hobby game with a known group of friends.

CREATE TABLE characters (
  bbs_name TEXT PRIMARY KEY,
  -- Case-insensitive lookup for the in-game character name (used by arena
  -- challenges and inn mail). NULL until character_create runs.
  char_name_lower TEXT,
  password_hash TEXT NOT NULL,
  password_salt TEXT NOT NULL,
  data TEXT NOT NULL,
  updated_at INTEGER NOT NULL
);
CREATE INDEX idx_characters_char_name_lower ON characters(char_name_lower);

-- Bag of singleton JSON blobs: 'inn', 'gladiator_fights', 'gladiator_bets'.
CREATE TABLE singletons (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

CREATE TABLE news (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  recipient TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at INTEGER NOT NULL
);
CREATE INDEX idx_news_recipient ON news(recipient);

CREATE TABLE sessions (
  token TEXT PRIMARY KEY,
  bbs_name TEXT NOT NULL,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL
);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
