/*
	Copyright © 2023 Andrey Melnikov vafilor@gmail.com
*/
package cmd

import (
	"database/sql"
	"filed/data"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

// hash flag variables
var hashTimeRun *bool

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:   "hash [database file]",
	Short: "Hash files in a sqlite database",
	Long: `Hash files in a given sqlite database. 

If no database is given, the current directory contents are examined looking for one.`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			// TODO find the path if none is provided
			fmt.Println("Path must be provided since it is not implemented yet")
			return
		}

		hashDatabasePath := args[0]

		db, err := sql.Open("sqlite3", hashDatabasePath)
		if err != nil {
			log.Fatal(err)
		}

		start := time.Now()
		hasher := data.NewHasher(db, &start)
		if err := hasher.Hash(); err != nil {
			log.Fatal(err)
		}

		end := time.Now()
		duration := end.Sub(start)

		if *hashTimeRun {
			fmt.Printf("Took: %f seconds\n", duration.Seconds())
		}
	},
}

func init() {
	rootCmd.AddCommand(hashCmd)

	hashTimeRun = hashCmd.Flags().Bool("time", false, "--time")
}
