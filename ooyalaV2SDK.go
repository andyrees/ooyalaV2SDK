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

// The NewApi Factory function returns a new api instance
func NewApi(api_key, api_secret string, expires int64) *OoyalaApi {
	api := OoyalaApi{}
	params := make(map[string]string)
	params["api_key"] = api_key
	params["expires"] = fmt.Sprint(int64(time.Now().Unix() + expires))
	api.Params = params

	api.BaseUrl = "https://api.ooyala.com"
	api.CacheBaseUrl = "https://cdn-api.ooyala.com"
	api.Secret = api_secret

	return &api
}

// Api implements the Ooyala Api
type OoyalaApi struct {
	Params          map[string]string
	BaseUrl         string
	CacheBaseUrl    string
	UsedBaseUrl     string
	Secret          string
	Body            string
	Http_method     string
	Request_path    string
	Response        string
	ResponseHeaders string
	Filter          string
	Signature       string
	FinalUrl        string
}

// Request retry wrapper ensures success
// eliminating transient errors
func (a *OoyalaApi) send() error {
	currentRetry := 0
	retryCount := 3

	for {
		err := a.sendRequest()
		if err != nil {
			currentRetry += 1
			if currentRetry > retryCount {
				return err
			}
			continue
		}
		return err
	}
}

func (a *OoyalaApi) generateFinalUrl() {
	if val, ok := a.Params["where"]; ok {
		a.FinalUrl = fmt.Sprintf(
			"%s%s?api_key=%s&where=%s&signature=%s&expires=%s",
			a.UsedBaseUrl,
			a.Request_path,
			a.Params["api_key"],
			url.QueryEscape(val),
			a.Signature,
			a.Params["expires"],
		)
	} else {
		a.FinalUrl = fmt.Sprintf(
			"%s%s?api_key=%s&signature=%s&expires=%s",
			a.UsedBaseUrl,
			a.Request_path,
			a.Params["api_key"],
			a.Signature,
			a.Params["expires"],
		)
	}

	for key, value := range a.Params {
		switch key {
		case "user_permission":
			a.FinalUrl += fmt.Sprintf("&%s=%s", key, value)
		case "limit":
			a.FinalUrl += fmt.Sprintf("&%s=%s", key, value)
		case "page_token":
			a.FinalUrl += fmt.Sprintf("&%s=%s", key, value)
		case "include":
			a.FinalUrl += fmt.Sprintf("&%s=%s", key, value)
		}
	}
}

// Send Http Request
func (a *OoyalaApi) sendRequest() error {
	a.GenerateSignature()

	a.UsedBaseUrl = a.CacheBaseUrl

	if a.Http_method != "GET" {
		a.UsedBaseUrl = a.BaseUrl
	}

	a.generateFinalUrl()

	client := &http.Client{}
	request, err := http.NewRequest(a.Http_method, a.FinalUrl, bytes.NewReader([]byte(a.Body)))
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

// View a resource
func (a *OoyalaApi) Get() error {
	a.Http_method = "GET"
	a.GenerateSignature()
	return a.send()
}

// update or modify an existing resource
func (a *OoyalaApi) Patch() error {
	a.Http_method = "PATCH"
	if a.Body == "" {
		return errors.New("NO DATA TO UPDATE")
	}
	a.GenerateSignature()
	return a.send()
}

// create a new resource
func (a *OoyalaApi) Post() error {
	a.Http_method = "POST"
	if a.Body == "" {
		return errors.New("NO NEW ASSET DATA")
	}
	a.GenerateSignature()
	// return a.send_request()
	return a.send()
}

// replace an existing reource
func (a *OoyalaApi) Put() error {
	a.Http_method = "PUT"
	a.GenerateSignature()
	return a.send()
}

// delete a resource
func (a *OoyalaApi) Delete() error {
	a.Http_method = "DELETE"
	a.GenerateSignature()
	return a.send()
}

// Generates the signature for a request
func (a *OoyalaApi) GenerateSignature() {
	signature := a.Secret + a.Http_method + a.Request_path
	hash := sha256.New()

	var keys []string

	for k, _ := range a.Params {
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
