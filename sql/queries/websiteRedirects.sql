-- name: GetWebsiteRedirect :one
SELECT * FROM websiteRedirects
WHERE id = ? LIMIT 1;

-- name: ListWebsiteRedirects :many
SELECT * FROM websiteRedirects
ORDER BY originalUrl;

-- name: CreateWebsiteRedirect :one
INSERT INTO websiteRedirects (
  originalUrl, redirectUrl
) VALUES (
  ?, ?
)
RETURNING *;

-- name: GetWebsiteRedirectByRedirectUrl :one
SELECT originalUrl FROM websiteRedirects
WHERE redirectUrl = ? LIMIT 1;
