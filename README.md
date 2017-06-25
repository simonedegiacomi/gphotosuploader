# G Photos Uploader - Beta
## Why? What is this?
Google Photos does not have a desktop uploader for Linux, neither an API to upload a photo programmatically ... So here
they are!

G Photos Uploader let's you upload photos from Linux (and, in theory, any OS for which you can compile a Go program)
specifying the file name or watching a directory for changes.
Furthermore, the project can also be used as a library that you can include in other Go programs.

## Disclaimer
G Photos Uploader is an unofficial tool, I (and any possible contributor) don't guarantee any result. Any security or
other kind of issues are at your own risk.

## How can i use it?
### Standalone tool
To use G Photos Uploader as a standalone tool you need to create two files:
- cookies.json: A JSON file that contains cookies to authenticate the HTTPS requests to upload the images. An example of
the file can be found [here](examples/cookies-sample.json);
- number: A simple text file with only one line with a special number. [Here](documentation/how-to-authenticate.md) you
can read more about how to get it;

Once you have the two files you're ready to go, just launch the program. For example, to upload a file named image.png:
```sh
go run main.go --upload ./image.png --cookies ./cookies.json --number ./number
```

Or to watch a directory:
```sh
go run main.go --watch path/to/photos --cookies ./cookies.json --number ./number --maxConcurrent 4
```

You can even upload all the photos of a directory and then start to watch another one:
```sh
go run main.go --upload /path/to/old/photos --upload /downloads/cat.png --watch path/to/new/photos
```


The tool crates a file (which the default name is uploaded.txt) which is a list of uploaded files, which will not be
re-uploaded. You can specify your own file using the uploadedList argument.
To see all the available arguments, use --help.

To watch for file system events, G Photos Uploader uses [fsnotify](https://github.com/fsnotify/fsnotify).

### Library
As the Standalone client mode, you need to get your cookies and a special number.
You can read a simple example [here](examples/simple.go).

## Current State and Problems
This project is at a very beginning state.
Here are the major problems:
- You need to manually get the needed cookies and store them into a file;
- You need to manually get the enable number;
- Currently the responses from Google are taken for good ()an error handling is needed);
- The cookies will expire and there isn't any refresh system yet;
- The standalone tool will try to upload any file it founds, even if they are not image;

As you can see, contributions are welcome!