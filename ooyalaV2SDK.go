package ooyalaV2SDK

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// NewAPI Factory function returns a new api instance.
func NewAPI(apiKey, apiSecret string, expires int64) *OoyalaApi {
	api := OoyalaApi{}
	params := make(map[string]string)
	params["api_key"] = apiKey
	params["expires"] = fmt.Sprint(int64(time.Now().Unix() + expires))
	api.Params = params

	api.BaseURL = "https://api.ooyala.com"
	api.CacheBaseURL = "https://cdn-api.ooyala.com"
	api.Secret = apiSecret

	return &api
}

// OoyalaAPI implements the Ooyala Api
type OoyalaAPI struct {
	Params          map[string]string
	BaseURL         string
	CacheBaseURL    string
	UsedBaseURL     string
	Secret          string
	Body            string
	HTTPMethod      string
	RequestPath     string
	Response        string
	ResponseHeaders string
	Filter          string
	Signature       string
	FinalURL        string
}

// Request retry wrapper ensures success
// eliminating transient errors
func (a *OoyalaAPI) send() error {
	currentRetry := 0
	retryCount := 3

	for {
		err := a.sendRequest()
		if err != nil {
			currentRetry++
			if currentRetry > retryCount {
				return err
			}
			continue
		}
		return err
	}
}

func (a *OoyalaAPI) generateFinalURL() {
	if val, ok := a.Params["where"]; ok {
		a.FinalURL = fmt.Sprintf(
			"%s%s?api_key=%s&where=%s&signature=%s&expires=%s",
			a.UsedBaseURL,
			a.RequestPath,
			a.Params["api_key"],
			url.QueryEscape(val),
			a.Signature,
			a.Params["expires"],
		)
	} else {
		a.FinalURL = fmt.Sprintf(
			"%s%s?api_key=%s&signature=%s&expires=%s",
			a.UsedBaseURL,
			a.RequestPath,
			a.Params["api_key"],
			a.Signature,
			a.Params["expires"],
		)
	}

	for key, value := range a.Params {
		switch key {
		case "user_permission":
			a.FinalURL += fmt.Sprintf("&%s=%s", key, value)
		case "limit":
			a.FinalURL += fmt.Sprintf("&%s=%s", key, value)
		case "page_token":
			a.FinalURL += fmt.Sprintf("&%s=%s", key, value)
		case "include":
			a.FinalURL += fmt.Sprintf("&%s=%s", key, value)
		}
	}
}

// Send Http Request
func (a *OoyalaAPI) sendRequest() error {
	a.GenerateSignature()

	a.UsedBaseURL = a.CacheBaseURL

	if a.HTTPMethod != "GET" {
		a.UsedBaseURL = a.BaseURL
	}

	a.generateFinalURL()

	client := &http.Client{}
	request, err := http.NewRequest(a.HTTPMethod, a.FinalURL, bytes.NewReader([]byte(a.Body)))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Length", strconv.Itoa(len(a.Body)))
	request.Header.Add("Content-Type", "application/json; charset=utf-8")

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	switch response.StatusCode {
	case 200:
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		a.Response = string(contents)

	case 204:
		return errors.New("NO CONTENT")
	case 400:
		return errors.New("BAD REQUEST")
	case 401:
		return errors.New("NOT AUTHORISED")
	case 403:
		return errors.New("FORBIDDEN")
	case 404:
		return errors.New("NOT FOUND")
	case 429:
		return errors.New("INSUFFICIENT API CREDITS")
	}
	return nil
}

// Get or view a resource
func (a *OoyalaAPI) Get() error {
	a.HTTPMethod = "GET"
	a.GenerateSignature()
	return a.send()
}

// Patch or update an existing resource
func (a *OoyalaAPI) Patch() error {
	a.HTTPMethod = "PATCH"
	if a.Body == "" {
		return errors.New("NO DATA TO UPDATE")
	}
	a.GenerateSignature()
	return a.send()
}

// Post or create a new resource
func (a *OoyalaAPI) Post() error {
	a.HTTPMethod = "POST"
	if a.Body == "" {
		return errors.New("NO NEW ASSET DATA")
	}
	a.GenerateSignature()
	// return a.send_request()
	return a.send()
}

// Put or replace an existing reource
func (a *OoyalaAPI) Put() error {
	a.HTTPMethod = "PUT"
	a.GenerateSignature()
	return a.send()
}

// Delete a resource
func (a *OoyalaAPI) Delete() error {
	a.HTTPMethod = "DELETE"
	a.GenerateSignature()
	return a.send()
}

// GenerateSignature Generates the signature for a request
func (a *OoyalaAPI) GenerateSignature() {
	signature := a.Secret + a.HTTPMethod + a.RequestPath
	hash := sha256.New()

	var keys []string

	for k := range a.Params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		signature += k + "=" + a.Params[k]
	}

	signature += a.Body

	_, err := io.WriteString(hash, signature)
	if err != nil {
		log.Fatalln(err)
	}

	signature = base64.StdEncoding.EncodeToString(hash.Sum(nil))[0:43]
	signature = url.QueryEscape(signature)

	a.Signature = signature
}
