package monitor

import (
	"SpotifyLive/db"
	"SpotifyLive/spotify"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed sql/migrations
var _ embed.FS

func StartMonitor(spDcCookie string, dbPath string) {
	ctx := context.Background()

	// Create a channel to receive OS signals (e.g., SIGINT for Ctrl+C)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a ticker that ticks every minute (60 seconds)
	ticker := time.NewTicker(5 * time.Second)

	api := spotify.NewApiClient(spDcCookie)

	// Init Database
	database, err := db.InitDatabase(ctx, dbPath)
	if err != nil {
		fmt.Printf("\nError initializing the database: %s", err)
		return
	}
	defer database.Close()
	queries := db.New(database)

	// Goroutine that performs an action every minute
	processActivity(api, ctx, queries)
	go func() {
		for {
			select {
			case <-ticker.C:
				processActivity(api, ctx, queries)
			}
		}
	}()

	// Wait for the termination signal (Ctrl + C)
	<-signalChan
	fmt.Println("\nProgram terminated.")
	ticker.Stop()

}

func processActivity(api *spotify.ApiClient, ctx context.Context, queries *db.Queries) {
	// Get Friend Activity
	activityResponse, err := api.GetFriendActivity()
	if err != nil {
		fmt.Println("\nError getting activity:", err)
		return
	}

	for _, activity := range activityResponse.Friends {

		lastActivity, err := queries.GetLastFriendActivityByUserUri(ctx, sql.NullString{String: activity.User.Uri, Valid: true})
		if err != nil {
			fmt.Printf("\nNo previous user activity")
		} else if lastActivity.Timestamp == activity.Timestamp {
			//fmt.Printf("\nActivity already logged")
			continue
		}

		_, err = queries.GetUserByUri(ctx, activity.User.Uri)
		if err != nil {
			_, err = queries.CreateUser(ctx, db.CreateUserParams{
				Uri:      activity.User.Uri,
				Name:     activity.User.Name,
				ImageUrl: sql.NullString{String: activity.User.ImageUrl, Valid: true},
			})
			if err != nil {
				fmt.Printf("\nError inserting user %x: [%s]", activity.User, err)
				continue
			}
		}

		_, err = queries.GetAlbumByUri(ctx, activity.Track.Album.Uri)
		if err != nil {
			_, err = queries.CreateAlbum(ctx, db.CreateAlbumParams{
				Uri:  activity.Track.Album.Uri,
				Name: activity.Track.Album.Name,
			})
			if err != nil {
				fmt.Printf("\nError inserting album %x: [%s]", activity.Track.Album, err)
				continue
			}
		}

		_, err = queries.GetArtistByUri(ctx, activity.Track.Artist.Uri)
		if err != nil {
			_, err = queries.CreateArtist(ctx, db.CreateArtistParams{
				Uri:  activity.Track.Artist.Uri,
				Name: activity.Track.Artist.Name,
			})
			if err != nil {
				fmt.Printf("\nError inserting artist %x: [%s]", activity.Track.Artist, err)
				continue
			}
		}

		_, err = queries.GetTrackContextByUri(ctx, activity.Track.Context.Uri)
		if err != nil {
			_, err = queries.CreateTrackContext(ctx, db.CreateTrackContextParams{
				Uri:        activity.Track.Context.Uri,
				Name:       activity.Track.Context.Name,
				TrackIndex: int64(activity.Track.Context.Index),
			})
			if err != nil {
				fmt.Printf("\nError inserting track context %x: [%s]", activity.Track.Context, err)
				continue
			}
		}

		_, err = queries.GetTrackByUri(ctx, activity.Track.Uri)
		if err != nil {
			_, err = queries.CreateTrack(ctx, db.CreateTrackParams{
				Uri:       activity.Track.Uri,
				Name:      activity.Track.Name,
				ImageUrl:  sql.NullString{String: activity.Track.ImageUrl, Valid: true},
				AlbumUri:  sql.NullString{String: activity.Track.Album.Uri, Valid: true},
				ArtistUri: sql.NullString{String: activity.Track.Artist.Uri, Valid: true},
			})
			if err != nil {
				fmt.Printf("\nError inserting track %x: [%s]", activity.Track, err)
				continue
			}
		}

		_, err = queries.CreateFriendActivity(ctx, db.CreateFriendActivityParams{
			Timestamp:  activity.Timestamp,
			UserUri:    sql.NullString{String: activity.User.Uri, Valid: true},
			TrackUri:   sql.NullString{String: activity.Track.Uri, Valid: true},
			ContextUri: sql.NullString{String: activity.Track.Context.Uri, Valid: true},
		})
		if err != nil {
			fmt.Printf("\nError inserting activity %x: [%s]", activity, err)
			continue
		} else {
			fmt.Printf("\nInserted new user activity for user %s: %s [ %s ]", activity.User.Name, activity.Track.Name, activity.Track.Artist.Name)
		}
	}
}
