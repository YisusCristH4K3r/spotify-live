drop view curated;

CREATE TABLE FriendActivity_tmp
(
    timestamp BIGINT NOT NULL,
    user_uri  text,
    track_uri text,
    context_uri text,
    PRIMARY KEY (timestamp, user_uri),
    FOREIGN KEY (user_uri) REFERENCES Users (uri),
    FOREIGN KEY (track_uri) REFERENCES Tracks (uri),
    FOREIGN KEY (context_uri) REFERENCES TrackContexts(uri)
);

insert into FriendActivity_tmp(timestamp, user_uri, track_uri, context_uri)
select F.timestamp, F.user_uri, F.track_uri, T.context_uri
from FriendActivity F join Tracks T on F.track_uri = T.uri;

drop table FriendActivity;

alter table FriendActivity_tmp
    rename to FriendActivity;

create table Tracks_tmp
(
    uri        text
        primary key,
    name       text not null,
    image_url  text,
    album_uri  text
        references Albums,
    artist_uri text
        references Artists
);

insert into Tracks_tmp(uri, name, image_url, album_uri, artist_uri)
select uri, name, image_url, album_uri, artist_uri
from Tracks;

drop table Tracks;

alter table Tracks_tmp
    rename to Tracks;

create view Curated as
select datetime(round(timestamp / 1000), 'unixepoch') time,
       F.user_uri                                     user,
       T.name                                         track,
       T.uri                                          uri,
       A2.name                                        artist,
       A.name                                         album,
       C.name as                                      playlist
from FriendActivity F
         join Tracks T on T.uri = F.track_uri
         join Albums A on A.uri = T.album_uri
         join Artists A2 on T.artist_uri = A2.uri
         join TrackContexts C on F.context_uri = C.uri
order by time desc