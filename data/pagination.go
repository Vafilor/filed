package data

import (
	"database/sql"
	"fmt"
)

type FilePaginator struct {
	db     *sql.DB
	limit  int
	query  string
	offset int
}

type FilesStatisticsPaginator struct {
	db     *sql.DB
	limit  int
	query  string
	offset int
}

func NewFilePaginator(db *sql.DB, query *string, limit int) *FilePaginator {
	return &FilePaginator{
		db:     db,
		limit:  limit,
		query:  fmt.Sprintf("%v LIMIT %v", *query, limit),
		offset: 0,
	}
}

// Next will proceed to the next offset in the query
// when there is no more data left, the number of returned files with be 0, so check that for the end
func (f *FilePaginator) Next() ([]*File, error) {
	query := f.query

	if f.offset > 0 {
		query += fmt.Sprintf(" OFFSET %v", f.offset)
	}

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resultCount := 0
	result := make([]*File, f.limit)

	for rows.Next() {
		var file File

		if err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.Size,
			&file.ModifiedAt,
			&file.IsDirectory,
			&file.Hash); err != nil {
			return nil, err
		}

		result[resultCount] = &file
		resultCount++
	}

	if resultCount == 0 {
		return []*File{}, nil
	}

	f.offset += resultCount

	if resultCount < f.limit {
		return result[0:resultCount], nil
	}

	return result, nil
}

func NewFileStatisticsPaginator(db *sql.DB, query *string, limit int) *FilesStatisticsPaginator {
	return &FilesStatisticsPaginator{
		db:     db,
		limit:  limit,
		query:  fmt.Sprintf("%v LIMIT %v", *query, limit),
		offset: 0,
	}
}

// Next will proceed to the next offset in the query
// when there is no more data left, the number of returned files with be 0, so check that for the end
func (f *FilesStatisticsPaginator) Next() ([]*FilesStatistic, error) {
	query := f.query

	if f.offset > 0 {
		query += fmt.Sprintf(" OFFSET %v", f.offset)
	}

	rows, err := f.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	resultCount := 0
	result := make([]*FilesStatistic, f.limit)

	for rows.Next() {
		var filesStatistic FilesStatistic

		if err := rows.Scan(
			&filesStatistic.Hash,
			&filesStatistic.FileSize,
			&filesStatistic.FileCount,
			&filesStatistic.TotalFileSize); err != nil {
			return nil, err
		}

		result[resultCount] = &filesStatistic
		resultCount++
	}

	if resultCount == 0 {
		return []*FilesStatistic{}, nil
	}

	f.offset += resultCount

	if resultCount < f.limit {
		return result[0:resultCount], nil
	}

	return result, nil
}
