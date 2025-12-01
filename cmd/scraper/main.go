package main

import (
	monitor "SpotifyLive"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

const flagCookie = "cookie"
const flagDatabase = "db"

func main() {
	var cli = &cobra.Command{
		Use:   "spotyLive",
		Short: "Spotify tool CLI",
		Run: func(cmd *cobra.Command, args []string) {
			cookie, err := cmd.Flags().GetString(flagCookie)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			if cookie == "" {
				fmt.Println("Empty cookie")
				os.Exit(2)
			}
			db, err := cmd.Flags().GetString(flagDatabase)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			monitor.StartMonitor(cookie, db)
		},
	}
	cli.Flags().String(flagCookie, "", "Spotify cookie")
	cli.Flags().String(flagDatabase, "database.db", "Sqlite database")

	err := cli.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
