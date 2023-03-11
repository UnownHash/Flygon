package golbatapi

var golbatUrl string
var bearerToken string
var apiSecret string

// SetApiUrl sets the API URL to use for the Golbat API calls
func SetApiUrl(url string, token string, secret string) {
	golbatUrl = url
	bearerToken = token
	apiSecret = secret
}
