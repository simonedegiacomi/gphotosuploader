package utils

import (
	"net/http"
	"os"
	"path"
	"strings"
)

const sniffLen = 512

// Check if the file at the given path is an image or a video
func IsImageOrVideo(fileName string) (bool, error) {
	extension := path.Ext(fileName)
	if isExtensionSupported(extension) {
		return true, nil
	}

	// If extension check fails, try with mime
	if file, err := os.Open(fileName); err == nil {
		defer file.Close()
		return IsFileImageOrVideo(file)
	} else {
		return false, err
	}
}

// Check if the file has a image or video mime. This function read the first 512 bytes of the file.
// Before and after the reading of the file offset is reset
func IsFileImageOrVideo(file *os.File) (bool, error) {
	// Read first 512 bytes
	file.Seek(0, 0)
	buffer := make([]byte, sniffLen)
	if _, err := file.Read(buffer); err != nil {
		return false, err
	}

	// Reset the file
	file.Seek(0, 0)

	// Detect content type
	mime := http.DetectContentType(buffer)

	return strings.Contains(mime, "image/") || strings.Contains(mime, "video/"), nil
}
