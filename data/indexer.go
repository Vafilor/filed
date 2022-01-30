package data

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
)

var (
	ErrInvalidBatchSize = errors.New("invalid batch size, must be greater than 0")
)

type Indexer struct {
	repository        *SQLiteRepository
	BatchSize         int  // how many inserts to do at a time. Defaults to 1000
	IgnoreHiddenFiles bool // if true, hidden files are ignored. Defaults to false
}

func NewIndexer(repository *SQLiteRepository) *Indexer {
	return &Indexer{
		repository:        repository,
		BatchSize:         1000,
		IgnoreHiddenFiles: false,
	}
}

func insertFiles(repository *SQLiteRepository, files []*File) error {
	if len(files) == 0 {
		return nil
	}

	return repository.InsertFiles(files)
}

func (i *Indexer) Index(rootPath *string) error {
	if i.BatchSize < 1 {
		return ErrInvalidBatchSize
	}

	pendingFiles := make([]*File, i.BatchSize)
	pendingIndex := 0

	totalFilesProcessed := 0

	// TODO hidden files
	err := filepath.Walk(*rootPath, func(path string, info fs.FileInfo, err error) error {
		file := &File{
			Path:        path,
			Size:        info.Size(),
			ModifiedAt:  info.ModTime().Unix(),
			IsDirectory: info.IsDir(),
		}

		pendingFiles[pendingIndex] = file
		pendingIndex++

		if pendingIndex == i.BatchSize {
			pendingIndex = 0
			totalFilesProcessed += i.BatchSize

			if err := insertFiles(i.repository, pendingFiles); err != nil {
				return err
			}
			fmt.Printf("Processed %v files/folders\n", totalFilesProcessed)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if pendingIndex != 0 {
		if err := insertFiles(i.repository, pendingFiles[0:pendingIndex]); err != nil {
			return err
		}

		totalFilesProcessed += pendingIndex
		fmt.Printf("Processed %v files/folders\n", totalFilesProcessed)
	}

	return nil
}
