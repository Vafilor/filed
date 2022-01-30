package data

import (
	"crypto/sha512"
	"fmt"
	"io"
	"os"
)

type Hasher struct {
	repository     *SQLiteRepository
	maxConcurrency int // Max number of files to be hashing at a time. Defaults to 5
	batchSize      int // Max number of files to load from database at a time. Defaults to 1000
}

func NewHasher(repository *SQLiteRepository) *Hasher {
	return &Hasher{
		repository:     repository,
		maxConcurrency: 5,
		batchSize:      1000,
	}
}

func sha512Hash(path *string) (*string, error) {
	file, err := os.Open(*path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	buffer := make([]byte, 30*1024)
	hasher := sha512.New()

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			_, err := hasher.Write(buffer[:n])
			if err != nil {
				return nil, err
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
	}

	sum := hasher.Sum(nil)
	hashValue := fmt.Sprintf("%x", sum)

	return &hashValue, nil
}

func updateFilesHash(repository *SQLiteRepository, files []*File) error {
	if len(files) == 0 {
		return nil
	}

	return repository.UpdateFilesHash(files)
}

// TODO - if a previous database exists, we might be able to reuse cache if modified_at is the same
func (h *Hasher) Hash() error {
	// Note we do not get unhashed files only because we are updating the hash also
	// If we were to get unhashed files only, we might skip results because of the offset
	// Doing this guarantees a linear order, though we do have to check if there is a hash already
	paginator := h.repository.ListFiles(h.batchSize)

	filesHashed := make([]*File, h.batchSize)
	hashedFileIndex := 0
	totalFilesHashed := 0

	for {
		files, err := paginator.Next()
		if err != nil {
			return err
		}

		if len(files) == 0 {
			break
		}

		for _, file := range files {
			if file.Hash != nil {
				continue
			}

			file.Hash, err = sha512Hash(&file.Path)
			if err != nil {
				fmt.Println(err)
				continue
			}

			filesHashed[hashedFileIndex] = file
			hashedFileIndex++

			if hashedFileIndex == h.batchSize {
				if err := updateFilesHash(h.repository, filesHashed); err != nil {
					return err
				}
				totalFilesHashed += h.batchSize
				hashedFileIndex = 0
				fmt.Printf("Hashed %v files\n", totalFilesHashed)
			}
		}
	}

	if hashedFileIndex > 0 {
		if err := updateFilesHash(h.repository, filesHashed[0:hashedFileIndex]); err != nil {
			return err
		}
		totalFilesHashed += hashedFileIndex
		fmt.Printf("Hashed %v files\n", totalFilesHashed)
	}

	return nil
}
