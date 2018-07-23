# How can this library upload photos?
This library acts like the Google Photos Web App.

If you open the Developers Tools of your browser and try to upload an image, you'll notice that the browser send three main requests:

## First request: obtain a new URL
The first request that the browser sends is used to get a new URL to which  upload the file.
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

Where '12345' is the size of the uploaded file. In the fields array you'll find other objects, which are not always needed. You can find a better structure of this object [here](https://github.com/3846masa/upload-gphotos/blob/master/src/util/uploadInfoTemplate.ts).
 

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
					"url": "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=0&upload_id=long_upload_id&file_id=000"
				},
				"content_type": "image/jpeg"
			}
		],
		"upload_id": "long_upload_id"
	}
}
```

## Second request: Upload of the image
The second request that the browser sends is the real file upload. In this case the request is a simple POST at the URL found on the first request, with the file as request body.

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
					"url": "https://photos.google.com/_/upload/uploadmedia/rupio/interactive?authuser=0&upload_id=long_upload_id&file_id=000"
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
		"upload_id": "long_upload_id"
	}
}
```

## Third request: Enable the image
Now the image is uploaded, but it's still not visible. It seems like you need to enable it or move it to an album before you can see the image.
To make the image visible in the homepage (outside any album) the browser sends a request containing a form with two values:
- at: a token that can be scraped from the index html page;
- f.req: The following JSON object:
```json
[  
   [  
      [  
         "mdpdU",
         "[[[\"base64 token\",\"name of the file you uploaded\",12345]]]",
         null,
         "generic"
      ]
   ]
]
```
Where 12345 is the timestamp of the image.

The response is the following JSON object:
```json
[
   [
      "wrb.fr",
      "mdpdU",
      "[[[\"base64 token\",[\"https://lh3.googleusercontent.com/id_of_the_image\",1920,1080,null,null,null,null,null,[1920,1080,1]\n]\n,2222,\"id to use to move the image into an album\",444444,555555,null,null,2]\n,0]\n]\n]\n",
      null,
      null,
      null,
      "generic"
   ],
   [
      "di",
      1111
   ],
   [
      "af.httprm",
      1111,
      "2222",
      33
   ]
]
```