/*
	Copyright © 2022 Andrey Melnikov vafilor@gmail.com
*/
package cmd

import (
	"filed/data"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

// hash flag variables
var hashDatabasePath *string
var hashTimeRun *bool

// hashCmd represents the hash command
var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "Hash files in a sqlite database",
	Long: `Hash files in a given sqlite database. 

If no database is given, the current directory contents are examined looking for one.`,
	Run: func(cmd *cobra.Command, args []string) {
		if *hashDatabasePath == "" {
			// TODO find the path if none is provided
			fmt.Println("Path must be provided since it is not implemented yet")
			return
		}

		fileRepository, err := data.NewSQLiteRepositoryFromFile(*hashDatabasePath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Hashing %v\n", *hashDatabasePath)

		start := time.Now()
		hasher := data.NewHasher(fileRepository)
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

	hashDatabasePath = hashCmd.Flags().StringP("database", "d", "", "filed_1643492165.db")
	hashTimeRun = hashCmd.Flags().Bool("time", false, "--time")
}
