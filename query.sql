-- name: CreateQuery :one
INSERT INTO
    queries (keywords, location)
VALUES
    (?, ?) RETURNING *;

-- name: GetQuery :one
SELECT
    *
FROM
    queries
WHERE
    keywords = ?
    AND location = ?;

-- name: UpdateQueryTS :exec
UPDATE queries
SET
    queried_at = CURRENT_TIMESTAMP
WHERE
    id = ?;

-- name: CreateOffer :exec
INSERT
OR IGNORE INTO offers (id, title, company, location, posted_at)
VALUES
    (?, ?, ?, ?, ?);

-- name: ListOffers :many
SELECT
    o.*
FROM
    queries q
    JOIN query_offers qo ON q.id = qo.query_id
    JOIN offers o ON qo.offer_id = o.id
WHERE
    q.id = ?
    -- AND DATE(o.posted_at) >= ?
ORDER BY
    o.posted_at DESC;

-- name: CreateQueryOfferAssoc :exec
INSERT INTO
    query_offers (query_id, offer_id)
VALUES
    (?, ?);
