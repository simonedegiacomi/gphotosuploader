package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"github.com/buger/jsonparser"
)

const (
	// NewUploadURL : Url to which send the request to get a new url to upload a new image
	NewUploadURL = "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=2"

	// Url to which send the request to enable an uploaded image
	//EnablePhotoUrl = "https://photos.google.com/_/PhotosUi/mutate"
	EnablePhotoUrl = "https://photos.google.com/u/2/_/PhotosUi/data/batchexecute"

	// Url to move an enabled photo into a specific album
	MoveToAlbumUrl = "https://photos.google.com/u/2/_/PhotosUi/data/batchexecute"
)

// Method that send a request with the file name and size to generate an upload url.
func (u *Upload) requestUploadURL() error {
	credentialsPersistentParameters := u.Credentials.PersistentParameters
	if credentialsPersistentParameters == nil {
		return fmt.Errorf("failed getting Credentials persistent parameters. Not set")
	}

	// Prepare json request
	jsonReq := RequestUploadURL{
		ProtocolVersion: "0.8",
		CreateSessionRequest: CreateSessionRequest{
			Fields: []interface{}{
				ExternalField{
					External: ExternalFieldObject{
						Name:     "file",
						Filename: u.Options.Name,
						Size:     u.Options.FileSize,
					},
				},

				// Additional fields
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "auto_create_album",
						Content:     "camera_sync.active",
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "auto_downsize",
						Content:     "true",
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "storage_policy",
						Content:     "use_manual_setting",
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "disable_asbe_notification",
						Content:     "true",
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "client",
						Content:     "photoweb",
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "effective_id",
						Content:     credentialsPersistentParameters.UserId,
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "owner_name",
						Content:     credentialsPersistentParameters.UserId,
						ContentType: "text/plain",
					},
				},
			},
		},
	}

	// Create http request
	jsonStr, err := json.Marshal(jsonReq)
	req, err := http.NewRequest("POST", NewUploadURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		return errors.New(fmt.Sprintf("Can't create upload URL request: %v", err.Error()))
	}

	// Add headers for the request
	req.Header.Add("x-guploader-client-info", "mechanism=scotty xhr resumable; clientVersion=156351954")

	// Make the request
	res, err := u.Credentials.Client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error during the request to get the upload URL: %v", err.Error()))
	}
	defer res.Body.Close()

	// Parse the json response
	jsonResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return responseReadingError()
	}

	u.url, err = jsonparser.GetString(jsonResponse, "sessionStatus", "externalFieldTransfers", "[0]", "putInfo", "url")
	return err
}

func responseReadingError() error {
	return fmt.Errorf("can't read response")
}

// This method upload the file to the URL received from requestUploadUrl.
// When the upload is completed, the method updates the base64UploadToken field
func (u *Upload) uploadFile() (token string, err error) {
	if u.url == "" {
		return "", errors.New("the url field is empty, make sure to call requestUploadUrl first")
	}

	// Create the request
	req, err := http.NewRequest("POST", u.url, u.Options.Stream)
	if err != nil {
		return "", fmt.Errorf("can't create upload URL request: %v", err.Error())
	}

	// Prepare request headers
	req.Header.Add("content-type", "application/octet-stream")
	req.Header.Add("content-length", fmt.Sprintf("%v", u.Options.FileSize))
	req.Header.Add("X-HTTP-Method-Override", "PUT")

	// Upload the image
	res, err := u.Credentials.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("can't upload the image, got: %v", err)
	}
	defer res.Body.Close()

	// Parse the response
	jsonRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", nil
	}

	return jsonparser.GetString(jsonRes, "sessionStatus", "additionalInfo", "uploader_service.GoogleRupioAdditionalInfo", "completionInfo", "customerSpecificInfo", "upload_token_base64")
}

// Request that enables the image once it gets uploaded
func (u *Upload) enablePhoto(uploadTokenBase64 string) (enabledUrl string, err error) {

	innerJson := []interface{}{
		[]interface{}{
			[]interface{}{
				uploadTokenBase64,
				u.Options.Name,
				u.Options.Timestamp,
			},
		},
	}
	innerJsonStr, err := json.Marshal(innerJson)
	if err != nil {
		return "", err
	}

	jsonReq := []interface{}{
		[]interface{}{
			[]interface{}{
				"mdpdU",
				string(innerJsonStr),
				nil,
				"generic",
			},
		},
	}

	jsonStr, err := json.Marshal(jsonReq)
	if err != nil {
		return "", err
	}

	// Form that contains the two request field
	form := url.Values{}

	// And add it to the form
	form.Add("f.req", string(jsonStr))

	// Second field of the form: "at", which should be an API key or something
	form.Add("at", u.Credentials.RuntimeParameters.AtToken)

	// Create the request
	req, err := http.NewRequest("POST", EnablePhotoUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("can't create the request to enable the image: %v", err.Error())
	}

	// Add headers
	req.Header.Add("content-type", "application/x-www-form-urlencoded;charset=UTF-8")

	// Send the request
	res, err := u.Credentials.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during the request to enable the image: %v", err.Error())
	}
	defer res.Body.Close()

	// Read the response as a string
	jsonRes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", responseReadingError()
	}

	// Skip first characters which are not valid json
	jsonRes = jsonRes[6:]

	innerJsonRes, err := jsonparser.GetString(jsonRes, "[0]", "[2]")
	if err != nil {
		return "", unexpectedResponse()
	}
	eUrl, err := jsonparser.GetString([]byte(innerJsonRes), "[0]", "[0]", "[1]", "[1]", "[0]")
	if err != nil {
		return "", unexpectedResponse()
	}
	u.idToMoveIntoAlbum, err = jsonparser.GetString([]byte(innerJsonRes), "[0]", "[0]", "[1]", "[0]")
	if err != nil {
		return "", unexpectedResponse()
	}

	if err != nil {
		fmt.Println(err)
	}

	return eUrl, nil
}

func unexpectedResponse() error {
	return fmt.Errorf("unexpected JSON response structure")
}

// This method add the image to an existing album given the id
func (u *Upload) moveToAlbum(albumId string) error {
	if u.idToMoveIntoAlbum == "" {
		return errors.New(fmt.Sprint("can't move image to album without the enabled image id"))
	}

	innerJson := [2]interface{}{
		[1]string{u.idToMoveIntoAlbum},
		albumId,
	}
	innerJsonString, err := json.Marshal(innerJson)
	if err != nil {
		return err
	}

	jsonReq := []interface{}{
		[]interface{}{
			[]interface{}{
				"E1Cajb",
				string(innerJsonString),
				"generic",
			},
		},
	}
	jsonString, err := json.Marshal(jsonReq)
	if err != nil {
		return err
	}

	form := url.Values{}
	form.Add("f.req", string(jsonString))
	form.Add("at", u.Credentials.RuntimeParameters.AtToken)

	req, err := http.NewRequest("POST", MoveToAlbumUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("can't create the request to add the image into the album: %v", err.Error())
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded;charset=UTF-8")

	res, err := u.Credentials.Client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending the request to move the image: %v", err.Error())
	}
	defer res.Body.Close()

	// The image should now be part of the album
	return nil
}
