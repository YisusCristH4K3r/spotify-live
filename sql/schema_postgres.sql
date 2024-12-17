-- Table for storing User information
CREATE TABLE Users
(
    uri       VARCHAR(255) PRIMARY KEY,
    name      VARCHAR(255) NOT NULL,
    image_url VARCHAR(255)
);

-- Table for storing Album information
CREATE TABLE Albums
(
    uri  VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Table for storing Artist information
CREATE TABLE Artists
(
    uri  VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Table for storing TrackContext information
CREATE TABLE TrackContexts
(
    uri         VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    track_index INT          NOT NULL
);

-- Table for storing Tracks, linking to Album, Artist, and TrackContext
CREATE TABLE Tracks
(
    uri         VARCHAR(255) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    image_url   VARCHAR(255),
    album_uri   VARCHAR(255),
    artist_uri  VARCHAR(255),
    context_uri VARCHAR(255),
    FOREIGN KEY (album_uri) REFERENCES Albums (uri),
    FOREIGN KEY (artist_uri) REFERENCES Artists (uri),
    FOREIGN KEY (context_uri) REFERENCES TrackContexts (uri)
);

-- Table for storing Friends, linking to User and Track
CREATE TABLE Friends
(
    timestamp BIGINT NOT NULL,
    user_uri  VARCHAR(255),
    track_uri VARCHAR(255),
    PRIMARY KEY (timestamp, user_uri),
    FOREIGN KEY (user_uri) REFERENCES Users (uri),
    FOREIGN KEY (track_uri) REFERENCES Tracks (uri)
);