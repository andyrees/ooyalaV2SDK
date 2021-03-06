Ooyala-v2-sdk
==============

This Go SDK is simple re-implementation of the Ooyala client class sdks. As with the other sdks, it allows you to do GET, POST, PUT, PATCH and DELETE requests to the API as in the simple example below:

```Go
package main

import (
	"fmt"
	"github.com/andyrees/ooyalaV2SDK"
	"log"
)

func createNewApiInstance() *ooyalaV2SDK.OoyalaApi {
	api_secret := "<Your Secret Key>"
	api_key := "<Your API Key>"
	var expires int64 = 15 // set query to expire in 15 seconds

	return ooyalaV2SDK.NewApi(api_key, api_secret, expires)
}

func main() {
	api := createNewApiInstance()
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
```
