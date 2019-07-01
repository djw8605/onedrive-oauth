package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const graphURL = "https://graph.microsoft.com/v1.0/me/drive/root"

type onedrive struct {
}

func (s onedrive) Upload(source *url.URL, dest *url.URL) error {
	log.Trace("Upload started")

	// Get the bearer token
	token, err := getToken(dest.Scheme)
	if err != nil {
		return err
	}
	fmt.Println(token)

	return nil
}

func (s onedrive) Download(source *url.URL, dest *url.URL) error {
	log.Trace("Download started")
	log.Debug("In Download.  Source: ", source)
	log.Debug("Destination: ", dest)

	// Get the bearer token from the environment
	token, err := getToken(source.Scheme)
	if err != nil {
		return err
	}

	newURL := graphURL + ":/" + source.Path + ":/content"
	log.Debug("Using URL: ", newURL)
	// Set the headers
	client := &http.Client{}
	req := createReq(newURL, token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	log.Debug("Returned status: ", resp.Status)
	log.Debug("Returned headers: ", resp.Header)

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Debug("Response:", string(body[:]))
		return errors.New("Unable to download, error code: " + resp.Status)
	}

	destPath, _ := filepath.Abs(dest.Path)
	destFile := destPath

	if fi, err := os.Stat(destPath); os.IsNotExist(err) {
		// The destination does not exist, so treat it as a file
		destFile = destPath
	} else {
		// Path exists, check if it's a file or directory
		if fi.IsDir() {
			// Destination is a directory, so use the filename from the source
			destFile = path.Join(dest.Path, filepath.Base(source.Path))
		} else {
			// Destination is a file, so overwrite the file
			destFile = destPath
		}
	}

	log.Debug("Writing to: ", destFile)
	out, err := os.Create(destFile)
	if err != nil {
		log.Error("Unable to open or create file ", destFile, ": ", err)
		return err
	}
	defer out.Close()
	log.Trace("After create")

	_, err = io.Copy(out, resp.Body)

	defer resp.Body.Close()
	// Get the bearer token

	return nil
}
