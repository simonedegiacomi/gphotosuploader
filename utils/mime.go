package utils

import (
	"os"
	"net/http"
	"strings"
)

// Check if the file has a image or video mime. This function read the first 512 bytes of the file.
// Before and after the reading of the file offset is reset
func IsFileImageOrVideo (file *os.File) (bool, error) {
	// Read first 512 bytes
	file.Seek(0, 0)
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return false, err
	}

	// Reset the file
	file.Seek(0, 0)

	// Detect content type
	mime := http.DetectContentType(buffer)

	return (strings.Contains(mime, "image/") || strings.Contains(mime, "video/")), nil
}

// Check if the file at the given path is an image or a video
func IsImageOrVideo (path string) (bool, error) {
	if file, err := os.Open(path); err != nil {
		return false, err
	} else {
		defer file.Close()
		return IsFileImageOrVideo(file)
	}
}