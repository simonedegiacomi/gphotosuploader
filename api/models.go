package api

// Structure of the JSON object that it's sent to request a new url to upload a new photo
type RequestUploadURL struct {
	ProtocolVersion      string `json:"protocolVersion"`
	CreateSessionRequest CreateSessionRequest `json:"createSessionRequest"`
}

// Inner object of the request to get a new url to upload a photo.
type CreateSessionRequest struct {
	// The fields array is a slice that should contain only ExternalField or InternalField structs
	Fields []interface{} `json:"fields"`
}

// Possible field for the Fields slice in the CreateSessionRequest struct
type ExternalField struct {
	External ExternalFieldObject `json:"external"`
}

// Possible field for the Fields slice in the CreateSessionRequest struct
type InlinedField struct {
	Inlined InlinedFieldObject `json:"inlined"`
}

// Struct that describes the file that need to be uploaded. This objects should be contained in a ExternalField
type ExternalFieldObject struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// Struct used to define parameters of the upload. This object should be contained in a InternalField
type InlinedFieldObject struct {
	Name        string `json:"name"`
	Content     string `json:"contentType"`
	ContentType string `json:"contentType"`
}




// Struct that represents the JSON response from the request to get an upload url
type UploadURLRequestResponse struct {
	SessionStatus SessionStatus `json:"sessionStatus"`
}

// Struct that represents the inner JSON object of the UploadURLRequestResponse
type SessionStatus struct {
	// Field used in the response for the request to get a new upload URL
	ExternalFieldTransfers []ExternalFieldTransfer `json:"externalFieldTransfers"`

	// Field used in the UploadImageResponse
	AdditionalInfo         AdditionalInfo `json:"additionalInfo"`
}

// Item of the ExternalFieldTransfers slice of SessionStatus
type ExternalFieldTransfer struct {
	Name    string `json:"name"`
	PutInfo PutInfo `json:"putInfo"`
}

// Container of the url to use to upload a new photo. It's contained in the ExternalFieldTransfer
type PutInfo struct {
	Url string `json:"url"`
}




// JSON representation of the response from the upload image request.
type UploadImageResponse struct {
	SessionStatus SessionStatus`json:"sessionStatus"`
}

// Struct used in SessionStatus in the response of the upload of a new image
type AdditionalInfo struct {
	UploadService GoogleRupioAdditionalInfo `json:"uploader_service.GoogleRupioAdditionalInfo"`
}

// Used in AdditionalInfo for image upload response
type GoogleRupioAdditionalInfo struct {
	CompletionInfo CompletionInfo `json:"completionInfo"`
}

// Used in GoogleRupioAdditionalInfo for image upload response
type CompletionInfo struct {
	CustomerSpecificInfo CustomerSpecificInfo `json:"customerSpecificInfo"`
}

// Used in CompletitionInfo and contains a token field used to enable the image in the future
type CustomerSpecificInfo struct {
	UploadTokenBase64 string `json:"upload_token_base64"`
}



type EnableImageRequest []interface{}

type FirstItemEnableImageRequest []InnerItemFirstItemEnableImageRequest

type InnerItemFirstItemEnableImageRequest interface{}

type SecondInnerArray []MapOfItemsToEnable

type MapOfItemsToEnable map[string]ItemToEnable

type ItemToEnable []ItemToEnableArray

type ItemToEnableArray []InnerItemToEnableArray

type InnerItemToEnableArray interface{}





type ApiTokenContainer struct {
	Token string `json:"SNlM0e"`
}