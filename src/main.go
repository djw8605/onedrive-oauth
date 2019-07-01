package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/mkideal/cli"
)

type argT struct {
	cli.Helper
	Debug bool `cli:"d,debug" usage:"Debug output"`
}

type service interface {
	Download(source *url.URL, dest *url.URL) error
	Upload(source *url.URL, dest *url.URL) error
}

func main() {
	os.Exit(cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		if argv.Debug {
			log.SetLevel(log.DebugLevel)
		}

		if ctx.NArg() != 2 {
			fmt.Print(ctx.Usage())
			return errors.New("Not enough arguments, must have src and dest")
		}

		args := ctx.Args()
		// Download
		src, err := url.Parse(args[0])
		if err != nil {
			fmt.Println("Failed to parse source")
			return nil
		}
		dst, err := url.Parse(args[1])
		if err != nil {
			fmt.Println("Failed to parse destination")
			return nil
		}

		if src.Scheme == "onedrive" && dst.Scheme == "" {
			o := onedrive{}
			return o.Download(src, dst)
		} else if src.Scheme == "" && dst.Scheme == "onedrive" {
			o := onedrive{}
			return o.Upload(src, dst)
		}

		fmt.Printf("src = %s, dst = %s\n", src, dst)

		// Upload

		return nil
	}))
}

// Retrieves the token from the HTCondor directory
func getToken(scheme string) (string, error) {
	// Get the environment variable TOKEN
	log.Trace("In getToken")

	credsDir := os.Getenv("_CONDOR_CREDS")
	if credsDir == "" {
		return "", errors.New("Environment variable not found: _CONDOR_CREDS")
	}
	log.Debug("Found _CONDOR_CREDS = ", credsDir)

	// Try opening the file
	fileName := path.Join(credsDir, scheme+".use")
	log.Debug("Trying to open: ", fileName)

	// Open the file, read the json, and get the access token
	jsonFile, err := os.Open(fileName)
	// if we os.Open returns an error then handle it
	if err != nil {
		return "", err
	}
	log.Debug("Open credential file: ", fileName)

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our map array
	var result map[string]interface{}
	err = json.Unmarshal([]byte(byteValue), &result)

	if err != nil {
		return "", err
	}

	token, _ := result["access_token"].(string)
	log.Debug("Got a token. (not printing for security purposes)")

	return token, nil
}

func createReq(method string, token string) *http.Request {
	req, _ := http.NewRequest("GET", method, nil)
	req.Header.Add("Authorization", "Bearer "+token)

	return req
}
