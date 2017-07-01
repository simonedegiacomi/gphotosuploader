# How can this library upload photos?
As said in the repository README, there are no APIs for GooglePhotos, so this library acts
like the Google Photos Web App.

If you open the Developers Tools of your browser and try to upload an image, you'll notice
that the browser send three main requests:


## First request: obtain a new URL
The first request that the browser sends is used to get a new URL to which you can upload
the photo.
The request is a POST at the 'https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=2' url.
In the request body you'll find a JSON object like this:

```json
{
	"protocolVersion": "0.8",
	"createSessionRequest": {
		"fields": [
			{
				"external": {
					"name": "file",
					"filename": "name of the file you uploaded",
					"put": {},
					"size": 12345
				}
			},
		]
	}
}
```

Where '12345' is the size of the uploaded file. In the fields array you'll find other objects
 which are not always needed. You can find a better structure of this object [here](https://github.com/3846masa/upload-gphotos/blob/master/src/utils/request-json-template-generator.js).
 

As response for this request you'll find another JSON object, this time like this:
```json
{
	"sessionStatus": {
		"state": "OPEN",
		"externalFieldTransfers": [
			{
				"name": "file",
				"status": "IN_PROGRESS",
				"bytesTransferred": 0,
				"bytesTotal": 12345,
				"putInfo": {
					"url": "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=0&upload_id=long-upload-id&file_id=000"
				},
				"content_type": "image/jpeg"
			}
		],
		"upload_id": "long-upload-id"
	}
}
```

## Second request: Upload of the image
The second request that the browser sends is the real image upload.
In this case the request is a simple POST at the URL found on the first request, which the image as 
request body.

The JSON object you'll find as a result is something like this:
```json
{
	"sessionStatus": {
		"state": "FINALIZED",
		"externalFieldTransfers": [
			{
				"name": "file",
				"status": "COMPLETED",
				"bytesTransferred": 12345,
				"bytesTotal": 12345,
				"putInfo": {
					"url": "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=0&upload_id=long-upload-id&file_id=000"
				},
				"content_type": "image/jpeg"
			}
		],
		"additionalInfo": {
			"uploader_service.GoogleRupioAdditionalInfo": {
				"completionInfo": {
					"status": "SUCCESS",
					"customerSpecificInfo": {
						"upload_token_base64": "upload-token"
					}
				}
			}
		},
		"upload_id": "long-upload-id"
	}
}
```

## Third request: Enable the image
Now the image is uploaded, but it's still not visible. It seems like you need to enable it or move it to an album before
you can see the image.
To make the image visible in the homepage (outside any album) the browser send a request with the following JSON object:
```json
[
	"af.maf",
	[
		[
			"af.add",
			137530650,
			[
				{
					"137530650": [
						[
							[
								"upload-token",
								"name of the file you uploaded",
								1234567890123
							]
						]
					]
				}
			]
		]
	]
]
```