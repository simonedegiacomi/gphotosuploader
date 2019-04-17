// The utils package contains methods used by the tool. All these methods are not needed if gphotosuploader is used as
// a library
package utils

import (
	"fmt"
	"gopkg.in/headzoo/surf.v1/errors"
	"os"
)

// Slice of name of file and directories to upload
type FilesToUpload []string

// Slice of names of directories to watch
type DirectoriesToWatch []string

func (a *FilesToUpload) String() string {
	return "File or directory to upload"
}

func (a *FilesToUpload) Set(name string) error {
	if _, err := os.Stat(name); os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("File or directory '%v' does not exist", name))
	}

	// https://stackoverflow.com/questions/24726341/append-to-stuct-that-only-has-one-slice-field-in-golang
	*a = append(*a, name)
	return nil
}

func (a *DirectoriesToWatch) String() string {
	return "Directory to watch"
}

func (a *DirectoriesToWatch) Set(name string) error {
	stat, err := os.Stat(name)
	if err != nil && os.IsNotExist(err) {
		return errors.New(fmt.Sprintf("Directory '%v' does not exist", name))
	}

	if !stat.IsDir() {
		return errors.New(fmt.Sprintf("'%v' is not a directory", name))
	}

	*a = append(*a, name)
	return nil
}
