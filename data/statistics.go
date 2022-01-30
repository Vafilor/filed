package data

import "fmt"

type Statistics struct {
	repository *SQLiteRepository
	batchSize  int // Max number of files to load from database at a time. Defaults to 1000
}

func NewStatistics(repository *SQLiteRepository) *Statistics {
	return &Statistics{
		repository: repository,
		batchSize:  1000,
	}
}

func (s *Statistics) Calculate() error {
	paginator := s.repository.CalculateFilesStatistics(s.batchSize)

	totalFilesInserted := 0

	for {
		filesStats, err := paginator.Next()
		if err != nil {
			return err
		}

		if len(filesStats) == 0 {
			break
		}

		if err := s.repository.InsertFilesStatistics(filesStats); err != nil {
			fmt.Println(err)
		} else {
			totalFilesInserted += s.batchSize
			fmt.Printf("Files Statistics Processed %v \n", totalFilesInserted)
		}
	}

	return nil
}
