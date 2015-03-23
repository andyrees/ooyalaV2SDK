package main

import (
	"fmt"
	"log"

	"github.com/andyrees/ooyalaV2SDK"
)

func createNewAPIInstance() *ooyalaV2SDK.OoyalaApi {
	apiSecret := "<Your Secret Key>"
	apiKey := "<Your API Key>"
	var expires int64 = 15 // set query to expire in 15 seconds

	return ooyalaV2SDK.NewApi(apiKey, apiSecret, expires)
}

func main() {
	api := createNewAPIInstance()
	api.Request_path = "/v2/assets"
	api.Params["include"] = "metadata"         // include metadata
	api.Params["where"] = "asset_type='video'" // filter by asset type

	// get first set of results
	err := api.Get()
	if err != nil {
		log.Fatalln(err.Error())
	}

	fmt.Printf("%+v\n", api.Response)
}
