package utils

import (
	"gopkg.in/headzoo/surf.v1/errors"
	"github.com/gphotosuploader/auth"
	"github.com/gphotosuploader/api"
	"sync"
)

type ConcurrentUploader struct {
	credentials       auth.Credentials

	// Buffered channel to limit concurrent uploads
	concurrentLimiter chan bool

	// Map of uploaded files (used as a set)
	uploadedFiles     map[string]bool

	waitingGroup      sync.WaitGroup
	waiting           bool

	CompletedUploads  chan *api.UploadOptions
	IgnoredUploads    chan *api.UploadOptions
	Errors            chan error
}

// Creates a new ConcurrentUploader using the specified credentials. The second argument is the maximum number
// of concurrent uploads (which must not be 0).
func NewUploader (credentials auth.Credentials, maxConcurrentUploads int) (*ConcurrentUploader, error) {
	if maxConcurrentUploads <= 0 {
		return nil, errors.New("maxConcurrentUploads must be greather than zero")
	}

	return &ConcurrentUploader{
		credentials: credentials,

		concurrentLimiter: make(chan bool, maxConcurrentUploads),

		uploadedFiles: make(map[string]bool),

		CompletedUploads: make(chan *api.UploadOptions),
		IgnoredUploads: make(chan *api.UploadOptions),
		Errors: make(chan error),
	}, nil
}

// Add files to the list of already uploaded files
func (u *ConcurrentUploader) AddUploadedFiles (files ...string) {
	for _, name := range files {
		u.uploadedFiles[name] = true
	}
}

// Enqueue a new upload. You must not call this method while waiting for some uploads to finish (The method return an
// error if you try to do it).
// Due to the fact that this method is asynchronous, if nil is return, it doesn't mean the the upload was completed,
// for that check the Errors and CompletedUploads channels
func (u *ConcurrentUploader) EnqueueUpload(options *api.UploadOptions) error {
	if u.waiting {
		return errors.New("Can't add new uploads when waiting")
	}
	if _, uploaded := u.uploadedFiles[options.FileToUpload]; uploaded {
		u.IgnoredUploads <- options
		return nil
	}

	// Check if the file is an image or a video
	if valid, err := IsImageOrVideo(options.FileToUpload); err != nil {
		u.Errors <- err
		return nil
	} else if !valid {
		u.IgnoredUploads <- options
		return nil
	}

	u.waitingGroup.Add(1)
	go u.uploadFile(options)

	return nil
}

func (u *ConcurrentUploader) uploadFile(options *api.UploadOptions) {
	u.concurrentLimiter <- true

	// Create a new upload
	upload, err := api.NewUpload(options, u.credentials)
	if err != nil {
		panic(err)
	}

	// Try to upload the image
	if err := upload.TryUpload(); err != nil {
		u.Errors <- err
	} else {
		u.uploadedFiles[options.FileToUpload] = true
		u.CompletedUploads <- options
	}

	u.waitingGroup.Done()
	<- u.concurrentLimiter
}

func (u *ConcurrentUploader) WaitUploadsCompleted () {
	u.waiting = true
	u.waitingGroup.Wait()
	u.waiting = false
}