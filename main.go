package main

import (
	"gphotosuploader/api"
	"os"
	"gphotosuploader/auth"
)

func main () {

	file, err := os.Open("/Volumes/RamDisk/download.png")
	if err != nil {
		panic(err)
	}

	options := api.UploadOptions{
		Name: "logo.png",
		FileToUpload: file,
	}

	// Credentials
	authFile, err := os.Open("./cookies.json")
	if err != nil {
		panic(err)
	}
	credentials := auth.NewCookieCredentialsFromFile(authFile)


	upload, err := api.NewUpload(&options, credentials)
	if err != nil {
		panic(err)
	}

	if err = upload.TryUpload(); err != nil {
		panic(err)
	}
}