package data

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteRepository provides methods to interact with database
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository creates a new SQLiteRepository providing the minimum needed data to start
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}

// NewSQLiteRepositoryFromFile attempts to create a SQLiteRepository by loading the
// sqlite database from the given path
func NewSQLiteRepositoryFromFile(path string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	fileRepository := NewSQLiteRepository(db)

	return fileRepository, nil
}

// Migrate will create all tables and run necessary migrations on a sqlite database
// It is safe to run this command multiple times
func (s *SQLiteRepository) Migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS files(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			size INTEGER,
			modified_at INTEGER,
			is_directory INTEGER,
			hashed_at INTEGER,
			hash TEXT
		);

		CREATE TABLE IF NOT EXISTS file_statistics(
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			hash TEXT NOT NULL,
			file_count INTEGER,
			file_size INTEGER,
			total_file_size INTEGER
		);
	`
	_, err := s.db.Exec(query)
	return err
}

// InsertFiles writes files to the underlying database
// the input files argument is not modified in any way, ie the id is not updated.
func (s *SQLiteRepository) InsertFiles(files []*File) error {
	ctx := context.Background()
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, err := transaction.Exec(
			"INSERT INTO files(path, size, modified_at, is_directory) VALUES(?,?, ?, ?)",
			file.Path, file.Size, file.ModifiedAt, file.IsDirectory)
		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}

// InsertFilesStatistics writes FilesStatistics to the underlying database
// the input statistics argument is not modified in any way, ie the id is not updated.
func (s *SQLiteRepository) InsertFilesStatistics(statistics []*FilesStatistic) error {
	ctx := context.Background()
	transaction, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, statistic := range statistics {
		_, err := transaction.Exec(
			"INSERT INTO file_statistics(hash, file_count, file_size, total_file_size) VALUES(?,?, ?, ?)",
			statistic.Hash, statistic.FileCount, statistic.FileSize, statistic.TotalFileSize)
		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}

// ListFiles returns a FilePaginator for all files in the database ordered by id ASC
// each result is limited to "limit" length
func (s *SQLiteRepository) ListFiles(limit int) *FilePaginator {
	query := `
		SELECT id, path, size, modified_at, is_directory, hash, hashed_at
		FROM files 
		WHERE is_directory = 0
		ORDER BY id ASC
	`

	return NewFilePaginator(s.db, &query, limit)
}

// ListHashedFiles returns a FilePaginator for all files in the database that have a hash,
// ordered by hash ASC. each result is limited to "limit" length
func (s *SQLiteRepository) ListHashedFiles(limit int) *FilesStatisticsPaginator {
	query := `
		SELECT hash, size, COUNT(*) file_count, SUM(size) total_file_size
		FROM files 
		WHERE is_directory = 0
			  AND hash IS NOT NULL
		ORDER BY hash ASC
	`

	return NewFileStatisticsPaginator(s.db, &query, limit)
}

func (s *SQLiteRepository) CalculateFilesStatistics(limit int) *FilesStatisticsPaginator {
	query := `
		SELECT hash, size, COUNT(*), SUM(size)
		FROM files
		WHERE hash IS NOT NULL
		GROUP BY hash
	`

	return NewFileStatisticsPaginator(s.db, &query, limit)
}
