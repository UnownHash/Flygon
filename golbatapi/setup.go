package golbatapi

var golbatUrl string
var apiSecret string

// SetApiUrl sets the API URL to use for the Golbat API calls
func SetApiUrl(url string, secret string) {
	golbatUrl = url
	apiSecret = secret
}
