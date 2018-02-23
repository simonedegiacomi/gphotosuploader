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
)

const (
	// NewUploadURL : Url to which send the request to get a new url to upload a new image
	NewUploadURL = "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=2"

	// Url to which send the request to enable an uploaded image
	EnablePhotoUrl = "https://photos.google.com/_/PhotosUi/mutate"

	// (Magic) Key to send in the request to enable the image
	EnablePhotoKey = 137530650

	// Url to move an enabled photo into a specific album
	MoveToAlbumUrl = "https://photos.google.com/u/2/_/PhotosUi/data/batchexecute"
)

// Method that send a request with the file name and size to generate an  upload url.
// This method returns the url or an error
func (u *Upload) requestUploadURL() error {
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
						Content:     u.Credentials.GetPersistentParameters().UserId,
						ContentType: "text/plain",
					},
				},
				InlinedField{
					Inlined: InlinedFieldObject{
						Name:        "owner_name",
						Content:     u.Credentials.GetPersistentParameters().UserId,
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
	res, err := u.Credentials.GetClient().Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error during the request to get the upload URL: %v", err.Error()))
	}
	defer res.Body.Close()

	// PArse the json response
	jsonResponse := UploadURLRequestResponse{}
	if err := json.NewDecoder(res.Body).Decode(&jsonResponse); err != nil {
		return errors.New(fmt.Sprintf("Can't parse json response for upload URL request: %v", err.Error()))
	}

	if len(jsonResponse.SessionStatus.ExternalFieldTransfers) <= 0 {
		return errors.New("An array of the request URL response is empty")
	}

	// Set the URL to which upload the file
	u.url = jsonResponse.SessionStatus.ExternalFieldTransfers[0].PutInfo.Url
	return nil
}

// This method upload the file to the URL received from requestUploadUrl.
// When the upload is completed, the method updates the base64UploadToken field
func (u *Upload) uploadFile() (*UploadImageResponse, error) {
	if u.url == "" {
		return nil, errors.New("The url field is empty, make sure to call requestUploadUrl first")
	}

	// Create the request
	req, err := http.NewRequest("POST", u.url, u.Options.Stream)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't create upload URL request: %v", err.Error()))
	}

	// Prepare request headers
	req.Header.Add("content-type", "application/octet-stream")
	req.Header.Add("content-length", fmt.Sprintf("%v", u.Options.FileSize))
	req.Header.Add("X-HTTP-Method-Override", "PUT")

	// Upload the image
	res, err := u.Credentials.GetClient().Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't upload the image, got: %v", err))
	}
	defer res.Body.Close()

	// Parse the response
	jsonRes := &UploadImageResponse{}
	if err := json.NewDecoder(res.Body).Decode(&jsonRes); err != nil {
		return nil, err
	}
	return jsonRes, nil
}

// Request that enables the image once it gets uploaded
func (u *Upload) enablePhoto(uploadResponse *UploadImageResponse) (*EnableImageResponse, error) {

	// Form that contains the two request field
	form := url.Values{}

	// First form field
	uploadTokenBase64 := uploadResponse.SessionStatus.AdditionalInfo.UploadService.CompletionInfo.CustomerSpecificInfo.UploadTokenBase64
	mapOfItems := MapOfItemsToEnable{}
	jsonReq := EnableImageRequest{
		"af.maf",
		[]FirstItemEnableImageRequest{
			[]InnerItemFirstItemEnableImageRequest{
				"af.add",
				EnablePhotoKey,
				SecondInnerArray{
					mapOfItems,
				},
			},
		},
	}
	mapOfItems[fmt.Sprintf("%v", EnablePhotoKey)] = ItemToEnable{
		ItemToEnableArray{
			[]InnerItemToEnableArray{
				uploadTokenBase64,
				u.Options.Name,
				u.Options.Timestamp,
			},
		},
	}

	// Stringify the first field
	jsonStr, err := json.Marshal(jsonReq)
	if err != nil {
		return nil, err
	}

	// And add it to the form
	form.Add("f.req", string(jsonStr))

	// Second field of the form: "at", which should be an API key or something
	form.Add("at", u.Credentials.GetRuntimeParameters().AtToken)

	// Create the request
	req, err := http.NewRequest("POST", EnablePhotoUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can't create the request to enable the image: %v", err.Error()))
	}

	// Add headers
	req.Header.Add("content-type", "application/x-www-form-urlencoded;charset=UTF-8")

	// Send the request
	res, err := u.Credentials.GetClient().Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error during the request to enable the image: %v", err.Error()))
	}
	defer res.Body.Close()

	// Read the response as a string
	bytesResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response
	jsonRes := &EnableImageResponse{}
	if err := json.Unmarshal(bytesResponse[6:], &jsonRes); err != nil {
		return nil, err
	}
	u.enabledImageId = jsonRes.getEnabledImageId()

	// Image enabled
	return jsonRes, nil
}

// This method add the image to an existing album given the id
func (u *Upload) moveToAlbum(albumId string) error {
	form := url.Values{}

	var innerJson [2]interface{}
	innerJson[0] = [1]string{u.enabledImageId}
	innerJson[1] = albumId
	innerJsonString, err := json.Marshal(innerJson)
	if err != nil {
		return err
	}

	var jsonReq [1][1][4]interface{}
	jsonReq[0][0][0] = "E1Cajb" // TODO: Extract a significant constant
	jsonReq[0][0][1] = string(innerJsonString)
	jsonReq[0][0][3] = "generic"
	jsonString, err := json.Marshal(jsonReq)
	if err != nil {
		return err
	}

	form.Add("f.req", string(string(jsonString)))
	form.Add("at", u.Credentials.GetRuntimeParameters().AtToken)

	req, err := http.NewRequest("POST", MoveToAlbumUrl, strings.NewReader(form.Encode()))
	if err != nil {
		return errors.New(fmt.Sprintf("Can't create the request to add the image into the album: %v", err.Error()))
	}
	req.Header.Add("content-type", "application/x-www-form-urlencoded;charset=UTF-8")

	res, err := u.Credentials.GetClient().Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Error sending the request to move the image: %v", err.Error()))
	}
	defer res.Body.Close()

	// The image should now be part of the album
	return nil
}
