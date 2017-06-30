package api

////////////////////////////
//// REQUEST URL MODELS ////
////////////////////////////
type RequestUploadURL struct {
	ProtocolVersion      string `json:"protocolVersion"`
	CreateSessionRequest CreateSessionUploadURLRequest `json:"createSessionRequest"`
}

type CreateSessionUploadURLRequest struct {
	Fields []interface{} `json:"fields"`
}

type ExternalFieldUploadURLRequest struct {
	External ExternalFieldObject `json:"external"`
}

type InlinedField struct {
	Inlined InlinedFieldObject `json:"inlined"`
}

type ExternalFieldObject struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type InlinedFieldObject struct {
	Name        string `json:"name"`
	Content     string `json:"contentType"`
	ContentType string `json:"contentType"`
}


///////////////////////////////////////
//// REQUEST URL MODELS - RESPONSE ////
///////////////////////////////////////

type UploadURLRequestResponse struct {
	SessionStatus SessionStatusUploadURLResponse `json:"sessionStatus"`
}

type SessionStatusUploadURLResponse struct {
	ExternalFieldTransfers []ExternalFieldTransferUploadURLResponse `json:"externalFieldTransfers"`
}

type ExternalFieldTransferUploadURLResponse struct {
	Name    string `json:"name"`
	PutInfo PutInfoUploadURLResponse `json:"putInfo"`
}

type PutInfoUploadURLResponse struct {
	Url string `json:"url"`
}



/////////////////////////////////
//// UPLOAD IMAGE - RESPONSE ////
/////////////////////////////////


type UploadImageResponse struct {
	SessionStatus SessionStatusUploadImageResponse `json:"sessionStatus"`
}

type SessionStatusUploadImageResponse struct {
	AdditionalInfo AdditionalInfoUploadImageResponse `json:"additionalInfo"`
}

type AdditionalInfoUploadImageResponse struct {
	UploadService ServiceUploadImageResponse `json:"uploader_service.GoogleRupioAdditionalInfo"`
}

type ServiceUploadImageResponse struct {
	CompletionInfo CompletionInfoUploadImageResponse `json:"completionInfo"`
}

type CompletionInfoUploadImageResponse struct {
	CustomerSpecificInfo CustomerSpecificInfoUploadImageResponse `json:"customerSpecificInfo"`
}

type CustomerSpecificInfoUploadImageResponse struct {
	UploadTokenBase64 string `json:"upload_token_base64"`
}


//////////////////////////////
//// ENABLE IMAGE REQUEST ////
//////////////////////////////
type EnableImageRequest []interface{}

type FirstItemEnableImageRequest []InnerItemFirstItemEnableImageRequest

type InnerItemFirstItemEnableImageRequest interface{}

type SecondInnerArray []MapOfItemsToEnable

type MapOfItemsToEnable map[string]ItemToEnable

type ItemToEnable []ItemToEnableArray

type ItemToEnableArray []InnerItemToEnableArray

type InnerItemToEnableArray interface{}



///////////////////////////////
//// Object with API token ////
///////////////////////////////

type ApiTokenContainer struct {
	Token string `json:"SNlM0e"`
	UserId string `json:"S06Grb"`
}