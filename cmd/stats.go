/*
	Copyright © 2023 Andrey Melnikov vafilor@gmail.com
*/
package cmd

import "github.com/spf13/cobra"

import (
	"filed/data"
	"fmt"
	"log"
	"time"
)

// stats flag variables
var statsTimeRun *bool

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Calculate statistics on hashed files",
	Long: `Calculates statistics on hashed files given a sqlite database file

The number of duplicate files (by hash) and their total size and file count are recorded`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			// TODO find the path if none is provided
			fmt.Println("Path must be provided since it is not implemented yet")
			return
		}

		statsDatabasePath := args[0]

		fmt.Printf("Calculating Stats of %v\n", statsDatabasePath)

		fileRepository, err := data.NewSQLiteRepositoryFromFile(statsDatabasePath)
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

	statsTimeRun = statsCmd.Flags().Bool("time", false, "--time")
}
