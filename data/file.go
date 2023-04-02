package data

type File struct {
	ID          int64
	Path        string
	Size        int64
	ModifiedAt  int64
	HashedAt    int64
	IsDirectory bool
	Hash        *string
}
