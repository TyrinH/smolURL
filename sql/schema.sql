CREATE TABLE IF NOT EXISTS authors (
  id   INTEGER PRIMARY KEY,
  name text    NOT NULL,
  bio  text
);

CREATE TABLE IF NOT EXISTS websiteRedirects (
    id   INTEGER PRIMARY KEY,
    originalUrl text    NOT NULL,
    redirectUrl  text    NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );