package data

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
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

	var waitGroup sync.WaitGroup
	pendingFiles := make([]*File, 0, i.BatchSize)

	progressChan := make(chan int)
	doneChan := make(chan bool)

	go func() {
		total := 0

		for delta := range progressChan {
			total += delta
			fmt.Printf("Processed %v files\n", total)
		}

		doneChan <- true
	}()

	skipPath := i.repository.filePath + "-journal"
	err := filepath.Walk(*rootPath, func(path string, info fs.FileInfo, err error) error {
		if path == i.repository.filePath || path == skipPath {
			return nil
		}

		//go fmt.Println(path)
		// TODO try having the method calls in the go routines - so just pass along the path and info
		file := &File{
			Path:        path,
			Size:        info.Size(),
			ModifiedAt:  info.ModTime().Unix(),
			IsDirectory: info.IsDir(),
		}

		if err != nil {
			fmt.Println(err)
		}

		pendingFiles = append(pendingFiles, file)

		if len(pendingFiles) == cap(pendingFiles) {
			waitGroup.Add(1)

			go func(files []*File) {
				defer waitGroup.Done()
				insertFiles(i.repository, files)

				progressChan <- len(files)
			}(pendingFiles)

			pendingFiles = make([]*File, 0, i.BatchSize)
		}

		return nil
	})

	if len(pendingFiles) != 0 {
		waitGroup.Add(1)
		go func(files []*File) {
			defer waitGroup.Done()

			insertFiles(i.repository, files)
			progressChan <- len(files)
		}(pendingFiles)
	}

	waitGroup.Wait()
	close(progressChan)

	<-doneChan

	return err
}
