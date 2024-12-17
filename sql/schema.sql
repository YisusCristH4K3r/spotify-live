-- Table for storing User information
CREATE TABLE Users
(
    uri       text PRIMARY KEY,
    name      text NOT NULL,
    image_url text
);

-- Table for storing Album information
CREATE TABLE Albums
(
    uri  text PRIMARY KEY,
    name text NOT NULL
);

-- Table for storing Artist information
CREATE TABLE Artists
(
    uri  text PRIMARY KEY,
    name text NOT NULL
);

-- Table for storing TrackContext information
CREATE TABLE TrackContexts
(
    uri         text PRIMARY KEY,
    name        text NOT NULL,
    track_index INT          NOT NULL
);

-- Table for storing Tracks, linking to Album, Artist, and TrackContext
CREATE TABLE Tracks
(
    uri         text PRIMARY KEY,
    name        text NOT NULL,
    image_url   text,
    album_uri   text,
    artist_uri  text,
    context_uri text,
    FOREIGN KEY (album_uri) REFERENCES Albums (uri),
    FOREIGN KEY (artist_uri) REFERENCES Artists (uri),
    FOREIGN KEY (context_uri) REFERENCES TrackContexts (uri)
);

-- Table for storing FriendActivity, linking to User and Track
CREATE TABLE FriendActivity
(
    timestamp BIGINT NOT NULL,
    user_uri  text,
    track_uri text,
    PRIMARY KEY (timestamp, user_uri),
    FOREIGN KEY (user_uri) REFERENCES Users (uri),
    FOREIGN KEY (track_uri) REFERENCES Tracks (uri)
);