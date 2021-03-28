CREATE TABLE quick_tests (
  id BLOB PRIMARY KEY,
  person TEXT,
  expired INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  registered_at TEXT
);


-- Only index valid registered tests for quick bulk expiration.
CREATE INDEX po_quick_tests_expired ON quick_tests(expired, registered_at) 
WHERE expired = 0 AND registered_at IS NOT NULL;

-- Only available tests
CREATE INDEX po_quick_tests_free ON quick_tests(expired, registered_at)
WHERE expired = 0 AND registered_at is NULL;
