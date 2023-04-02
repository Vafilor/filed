package hidden

import "syscall"

// IsHiddenFileName returns true if the file name is a hidden file
// this is denoted by having a . for the first character, as in .bashrc
func IsHiddenFileName(name string) (bool, error) {
	pointer, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return false, err
	}

	attributes, err := syscall.GetFileAttributes(pointer)
	if err != nil {
		return false, err
	}

	return attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}
