/*
	Copyright © 2022 Andrey Melnikov vafilor@gmail.com
*/
package cmd

import (
	"filed/data"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

// Index flag variables
var indexPath *string
var indexTimeRun *bool

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Indexes files and directories into a sqlite database",
	Long: `Indexes files and directories into a sqlite database.

A path is required that tells the program where to start. The process is recursive and will
index all files/folders under the given path. A new sqlite database is created each use, using the name
filed_<timestamp> where timestamp is the current unix timestamp`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("index called")

		now := time.Now()
		fileName := fmt.Sprintf("filed_%v.db", now.Unix())
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		fileName = dir + "/" + fileName

		fileRepository, err := data.NewSQLiteRepositoryFromFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		if err := fileRepository.Migrate(); err != nil {
			log.Fatal(err)
		}

		fmt.Println("Indexing " + *indexPath)

		indexer := data.NewIndexer(fileRepository)
		start := time.Now()

		if err := indexer.Index(indexPath); err != nil {
			log.Fatal(err)
		}

		end := time.Now()
		duration := end.Sub(start)
		fmt.Println("Finished indexing " + *indexPath)

		if *indexTimeRun {
			fmt.Printf("Took: %f seconds\n", duration.Seconds())
		}
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexPath = indexCmd.Flags().StringP("path", "p", "", "/Users/me")
	indexCmd.MarkFlagRequired("path")

	indexTimeRun = indexCmd.Flags().Bool("time", false, "--time")
}
