-- USERS --

-- name: CreateUser :one
INSERT INTO Users (uri, name, image_url)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetUserByUri :one
SELECT *
FROM Users
WHERE uri = $1;

-- name: UpdateUser :one
UPDATE Users
SET name      = $2,
    image_url = $3
WHERE uri = $1 RETURNING *;

-- name: DeleteUser :one
DELETE
FROM Users
WHERE uri = $1 RETURNING *;


-- ALBUMS --

-- name: CreateAlbum :one
INSERT INTO Albums (uri, name)
VALUES ($1, $2) RETURNING *;

-- name: GetAlbumByUri :one
SELECT *
FROM Albums
WHERE uri = $1;

-- name: UpdateAlbum :one
UPDATE Albums
SET name = $2
WHERE uri = $1 RETURNING *;

-- name: DeleteAlbum :one
DELETE
FROM Albums
WHERE uri = $1 RETURNING *;


-- ARTIST --

-- name: CreateArtist :one
INSERT INTO Artists (uri, name)
VALUES ($1, $2) RETURNING *;

-- name: GetArtistByUri :one
SELECT *
FROM Artists
WHERE uri = $1;


-- name: UpdateArtist :one
UPDATE Artists
SET name = $2
WHERE uri = $1 RETURNING *;

-- name: DeleteArtist :one
DELETE
FROM Artists
WHERE uri = $1 RETURNING *;


-- TRACK CONTEXT --

-- name: CreateTrackContext :one
INSERT INTO TrackContexts (uri, name, track_index)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetTrackContextByUri :one
SELECT *
FROM TrackContexts
WHERE uri = $1;

-- name: UpdateTrackContext :one
UPDATE TrackContexts
SET name        = $2,
    track_index = $3
WHERE uri = $1 RETURNING *;

-- name: DeleteTrackContext :one
DELETE
FROM TrackContexts
WHERE uri = $1 RETURNING *;


-- TRACKS --

-- name: CreateTrack :one
INSERT INTO Tracks (uri, name, image_url, album_uri, artist_uri, context_uri)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetTrackByUri :one
SELECT *
FROM Tracks
WHERE uri = $1;

-- name: UpdateTrack :one
UPDATE Tracks
SET name        = $2,
    image_url   = $3,
    album_uri   = $4,
    artist_uri  = $5,
    context_uri = $6
WHERE uri = $1 RETURNING *;

-- name: DeleteTrack :one
DELETE
FROM Tracks
WHERE uri = $1 RETURNING *;


-- FRIENDS --

-- name: CreateFriend :one
INSERT INTO Friends (timestamp, user_uri, track_uri)
VALUES ($1, $2, $3) RETURNING *;

-- name: GetFriendsByUserUri :many
SELECT *
FROM Friends
WHERE user_uri = $1;

-- name: UpdateFriendTimestamp :one
UPDATE Friends
SET timestamp = $2
WHERE user_uri = $1
  AND track_uri = $3 RETURNING *;

-- name: DeleteFriend :one
DELETE
FROM Friends
WHERE user_uri = $1
  AND track_uri = $2 RETURNING *;
