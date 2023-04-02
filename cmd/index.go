package cmd

import (
	"filed/data"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// Index flag variables
var indexTimeRun *bool
var skipHiddenFiles *bool

// tryGetDirectory attempts to get a directory from the input path, following links if needed
// if the path points to a link of a directory, it will attempt to resolve it
// an error is returned if the path can not be resolved to a directory
func tryGetDirectory(path string) (string, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	rootInfo, err := os.Lstat(absolutePath)
	if err != nil {
		return "", err
	}

	if rootInfo.Mode()&os.ModeSymlink != 0 {
		linkPath, err := filepath.EvalSymlinks(absolutePath)
		if err != nil {
			return "", err
		}

		return tryGetDirectory(linkPath)
	}

	if !rootInfo.IsDir() {
		if path[len(path)-1] == os.PathSeparator {
			return "", fmt.Errorf("'%v' is not a directory", path)
		}

		return tryGetDirectory(path + string(os.PathSeparator))
	}

	return absolutePath, nil
}

// indexCmd represents the index command
var indexCmd = &cobra.Command{
	Use:   "index [path to index]",
	Short: "Indexes files and directories into a sqlite database",
	Long: `Indexes files and directories into a sqlite database.

A path is required that tells the program where to start. The process is recursive and will
index all files/folders under the given path. A new sqlite database is created each use, using the name
filed_<timestamp> where timestamp is the current unix timestamp`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		indexPath := args[0]

		now := time.Now()
		fileName := fmt.Sprintf("filed_%v.db", now.Unix())

		fileRepository, err := data.NewSQLiteRepositoryFromFile(fileName)
		if err != nil {
			log.Fatal(err)
		}

		if err := fileRepository.Migrate(); err != nil {
			log.Fatal(err)
		}

		indexer := data.NewIndexer(fileRepository)
		indexer.SkipHiddenFiles = skipHiddenFiles != nil && *skipHiddenFiles

		start := time.Now()

		absolutePath, err := tryGetDirectory(indexPath)
		if err != nil {
			log.Fatal(err.Error())
		}

		fmt.Println("Indexing " + absolutePath)
		if err := indexer.Index(absolutePath); err != nil {
			log.Fatal(err)
		}

		end := time.Now()
		duration := end.Sub(start)
		fmt.Println("Finished indexing " + absolutePath)
		fmt.Printf("File: %s\n", fileName)

		if *indexTimeRun {
			fmt.Printf("Took: %f seconds\n", duration.Seconds())
		}
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)

	indexTimeRun = indexCmd.Flags().Bool("time", false, "--time")
	skipHiddenFiles = indexCmd.Flags().Bool("skipHidden", false, "--skipHidden")
}
