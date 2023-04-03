//go:build !windows

package hidden

// IsHiddenFile returns true if the file name is a hidden file
// this is denoted by having a . for the first character, as in .bashrc
func IsHiddenFileName(path, name string) (bool, error) {
	return name[0] == '.', nil
}
