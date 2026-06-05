# SpotifyLive

SpotifyLive is a Spotify friend activity monitor written in GO. It continuously checks your Spotify friends feed, detects new listening activity, and stores newly observed songs and related metadata in a local database.

The goal is to help you track your friends' music taste and listening habits over time: what they are currently playing, which artists and albums appear often, and what playlists or listening contexts they use.


## Features

- Monitors Spotify friend activity on a fixed interval.
- Stores newly detected friend activity in a SQLite database.
- Tracks users, tracks, artists, albums, and track contexts.
- Avoids duplicating unchanged activity by comparing the latest stored activity timestamp.
- Can run locally from the CLI.
- Includes setup/deployment scripts for running the monitor as a `systemd` service on a VPS.

## Requirements

- Go `1.26+`
- A Spotify `sp_dc` cookie from an authenticated Spotify web session
- SQLite database file path
- Optional, for deployment:
    - SSH access to a Linux VPS
    - `systemd`
    - `scp`
    - `make`

## Getting Your Spotify Cookie

The monitor needs your Spotify `sp_dc` cookie to call Spotify's friend activity endpoint.

1. Open Spotify Web in your browser.
2. Log in to your Spotify account.
3. Open your browser developer tools.
4. Go to the **Application / Storage** tab.
5. Find cookies for Spotify.
6. Copy the value of the `sp_dc` cookie.

> Keep this cookie private. Anyone with access to it may be able to access Spotify data as your account.

## Running Locally

You can run the monitor directly with Go:

```sh
go run ./cmd/scraper --cookie YOUR_SP_DC_COOKIE --db database.db
```

The `--db` flag controls where the SQLite database is stored. If omitted, it defaults to `database.db`.


## Running on a VPS

The project uses a `.env` file for local setup and deployment configuration. Copy the example `.env.example` file to `.env` and fill in the required variables.


### Variables

| Variable | Required | Description | 
| --- | --- | --- |
| `SP_DC_COOKIE` | Yes | Spotify authentication cookie used by the monitor. |
| `SSH_USER` | Deployment only | SSH username for the remote server. |
| `SSH_HOST` | Deployment only | SSH host/IP for the remote server. |
| `SSH_PORT` | Deployment only | SSH port. Defaults to `22`. |
| `REMOTE_DEPLOY_DIR` | Deployment only | Directory on the remote server where the monitor binary and database live. |
| `RUNAS_USER` | Deployment only | System user used to run the monitor service. Defaults to `spotify-app`. |
| `RUNAS_GROUP` | Deployment only | System group used to run the monitor service. Defaults to `spotify-app`. |

### Deployment

The project includes a deployment script that copies the built monitor binary to a remote server and installs/updates a `systemd` service.

Deploy the monitor to a VPS with:
```sh
make deploy
```