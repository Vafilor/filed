/*
	Copyright © 2022 Andrey Melnikov vafilor@gmail.com
*/
package cmd

import (
	"filed/data"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"time"
)

// stats flag variables
var statsDatabasePath *string
var statsTimeRun *bool

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Calculate statistics on hashed files",
	Long: `Calculates statistics on hashed files given a sqlite database file

The number of duplicate files (by hash) and their total size and file count are recorded`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Calculating Stats of %v\n", *statsDatabasePath)

		fileRepository, err := data.NewSQLiteRepositoryFromFile(*statsDatabasePath)
		if err != nil {
			log.Fatal(err)
		}

		if err := fileRepository.Migrate(); err != nil {
			log.Fatal(err)
		}

		start := time.Now()
		statistics := data.NewStatistics(fileRepository)
		if err := statistics.Calculate(); err != nil {
			log.Fatal(err)
		}

		end := time.Now()
		duration := end.Sub(start)

		if *statsTimeRun {
			fmt.Printf("Took: %f seconds\n", duration.Seconds())
		}
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsDatabasePath = statsCmd.Flags().StringP("database", "d", "", "filed_1643492165.db")
	statsTimeRun = statsCmd.Flags().Bool("time", false, "--time")
}
