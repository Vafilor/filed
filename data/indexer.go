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

func (i *Indexer) newFileGenerator(rootPath *string) <-chan *File {
	c := make(chan *File)

	go func() {
		// TODO hidden files
		err := filepath.Walk(*rootPath, func(path string, info fs.FileInfo, err error) error {
			c <- &File{
				Path:        path,
				Size:        info.Size(),
				ModifiedAt:  info.ModTime().Unix(),
				IsDirectory: info.IsDir(),
			}

			if err != nil {
				fmt.Println(err)
			}

			return nil
		})

		if err != nil {
			fmt.Println(err)
		}

		close(c)
	}()

	return c
}

func (i *Indexer) Index(rootPath *string) error {
	if i.BatchSize < 1 {
		return ErrInvalidBatchSize
	}

	pendingFiles := make([]*File, 0, i.BatchSize)

	totalFilesProcessed := 0

	fileChannel := i.newFileGenerator(rootPath)

	for file := range fileChannel {
		pendingFiles = append(pendingFiles, file)

		if len(pendingFiles) == cap(pendingFiles) {
			totalFilesProcessed += len(pendingFiles)

			if err := insertFiles(i.repository, pendingFiles); err != nil {
				return err
			}

			pendingFiles = pendingFiles[:0]
			fmt.Printf("Processed %v files/folders\n", totalFilesProcessed)
		}
	}

	if len(pendingFiles) != 0 {
		if err := insertFiles(i.repository, pendingFiles); err != nil {
			return err
		}

		totalFilesProcessed += len(pendingFiles)
		fmt.Printf("Processed %v files/folders\n", totalFilesProcessed)
	}

	return nil
}
