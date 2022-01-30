package data

type FilesStatistic struct {
	Hash          string
	FileCount     int
	FileSize      int64
	TotalFileSize int64
}

func NewFilesStatistic(hash string, fileCount int, fileSize int64) *FilesStatistic {
	return &FilesStatistic{
		Hash:          hash,
		FileCount:     fileCount,
		FileSize:      fileSize,
		TotalFileSize: int64(fileCount) * fileSize,
	}
}
