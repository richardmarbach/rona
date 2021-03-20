CREATE TABLE quick_tests (
  id BLOB PRIMARY KEY,
  person TEXT,
  expired INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  registered_at TEXT
);
