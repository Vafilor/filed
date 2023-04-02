package data

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"io"
	"os"
	"time"
)

const HashBufferSize = 30 * 1024 // 30MB

var SizeSuffixes = []string{"bytes", "kb", "mb", "gb", "tb"}

type Hasher struct {
	db             *sql.DB
	maxConcurrency int        // Max number of files to be hashing at a time. Defaults to 5
	batchSize      int        // Max number of files to load from database at a time. Defaults to 1000
	hashOlderThan  *time.Time // calculate hash for any files that were hashed before this
}

func NewHasher(db *sql.DB, hashOlderThan *time.Time) *Hasher {
	return &Hasher{
		db:             db,
		maxConcurrency: 5,
		batchSize:      1000,
		hashOlderThan:  hashOlderThan,
	}
}

func humanizeFileSize(size int64) string {
	finalSize := float64(size)
	for _, suffix := range SizeSuffixes {
		if finalSize < 1024.0 {
			return fmt.Sprintf("%.2f %s", finalSize, suffix)
		}

		finalSize /= 1024.0
	}

	lastSuffix := SizeSuffixes[len(SizeSuffixes)-1]
	return fmt.Sprintf("%f %s", finalSize, lastSuffix)
}

func sha512Hash(path *string) (*string, error) {
	file, err := os.Open(*path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	buffer := make([]byte, HashBufferSize)
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

func hashFiles(when *time.Time, files []*File) int {
	filesHashed := 0
	var err error = nil

	totalSize := int64(0)
	for _, file := range files {
		totalSize += file.Size
	}

	humanFriendlySize := humanizeFileSize(totalSize)
	fmt.Printf("Hashing %s of files \n", humanFriendlySize)

	for _, file := range files {
		file.HashedAt = when.Unix()
		file.Hash, err = sha512Hash(&file.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
		} else {
			filesHashed += 1
		}
	}

	return filesHashed
}

// TODO - if a previous database exists, we might be able to reuse cache if modified_at is the same
func (h *Hasher) Hash() error {
	now := time.Now()

	var hashOlderThan int64
	if h.hashOlderThan == nil {
		hashOlderThan = time.Now().Unix()
	} else {
		hashOlderThan = (*h.hashOlderThan).Unix()
	}

	totalFilesHashed := 0
	lastId := int64(0)

	for {
		files, err := getFilesToHash(h.db, lastId, hashOlderThan, int64(h.batchSize))
		if err != nil {
			return err
		}

		if len(files) == 0 {
			break
		}

		lastFile := files[len(files)-1]
		lastId = lastFile.ID

		totalFilesHashed += hashFiles(&now, files)
		if err := updateHashedFiles(h.db, files); err != nil {
			return err
		}

		fmt.Printf("Hashed %v files\n", totalFilesHashed)
	}

	return nil
}

// getFilesToHash returns files that should be hashed. Either the files have no hash,
// or were hashed before $hashedBefore
func getFilesToHash(db *sql.DB, afterId int64, hashedBefore int64, limit int64) ([]*File, error) {
	args := []interface{}{false, hashedBefore}

	query := `
		SELECT id, path, size
		FROM files
		WHERE is_directory = ?
		  AND (hashed_at IS NULL or hashed_at < ?)		
	`

	if afterId != 0 {
		query += " AND id > ?"
		args = append(args, afterId)
	}

	query += " ORDER BY id ASC LIMIT ?"

	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resultCount := 0
	result := make([]*File, limit)

	for rows.Next() {
		var file File

		if err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.Size,
		); err != nil {
			return nil, err
		}

		result[resultCount] = &file
		resultCount++
	}

	if resultCount == 0 {
		return []*File{}, nil
	}

	if int64(resultCount) < limit {
		return result[0:resultCount], nil
	}

	return result, nil
}

func updateHashedFiles(db *sql.DB, files []*File) error {
	ctx := context.Background()
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, file := range files {
		hashValue := "null"
		if file.Hash != nil {
			hashValue = *file.Hash
		}

		_, err := transaction.Exec(
			"UPDATE files SET hash = ?, hashed_at = ? WHERE id = ?",
			hashValue, file.HashedAt, file.ID)

		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}
