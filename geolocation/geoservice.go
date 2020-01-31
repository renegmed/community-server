package geolocation

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kelvins/geocoder/structs"
)

const (
	geocodeApiUrl = "https://maps.googleapis.com/maps/api/geocode/json?address="
)

type GoogleGeoCoder struct {
	apiKey string
	url    string
}

func NewGoogleGeoCoder(key string, apiUrl string) (GoogleGeoCoder, error) {

	if key == "" {
		fmt.Println("Api key  is empty.")
		err := errors.New("Error: Api key is empty.")
		return GoogleGeoCoder{}, err
	}

	if apiUrl == "" {
		apiUrl = geocodeApiUrl
	}

	return GoogleGeoCoder{apiKey: key, url: apiUrl}, nil
}

// httpRequest function send the HTTP request, decode the JSON
// and return a Results structure
func (g *GoogleGeoCoder) GeoLocationRequest(address string) (structs.Results, error) {

	var results structs.Results

	formattedAddress := strings.Replace(address, " ", "+", -1)
	url := fmt.Sprintf("%s%s&key=%s", g.url, formattedAddress, g.apiKey)

	fmt.Println(url)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return results, err
	}

	// For control over HTTP client headers, redirect policy, and other settings, create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		return results, err
	}

	// Callers should close resp.Body when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Use json.Decode for reading streams of JSON data
	err = json.NewDecoder(resp.Body).Decode(&results)
	if err != nil {
		return results, err
	}
	// The "OK" status indicates that no error has occurred, it means
	// the address was analyzed and at least one geographic code was returned
	if strings.ToUpper(results.Status) != "OK" {
		// If the status is not "OK" check what status was returned
		switch strings.ToUpper(results.Status) {
		case "ZERO_RESULTS":
			err = errors.New("No results found.")
			break
		case "OVER_QUERY_LIMIT":
			err = errors.New("You are over your quota.")
			break
		case "REQUEST_DENIED":
			err = errors.New("Your request was denied.")
			break
		case "INVALID_REQUEST":
			err = errors.New("Probably the query is missing.")
			break
		case "UNKNOWN_ERROR":
			err = errors.New("Server error. Please, try again.")
			break
		default:
			break
		}
	}

	return results, err
}

// func (g *GoogleGeoCoder) GetGeoLocation(address Address) (Coordinate, error) {
// 	gAddress := geocoder.Address{
// 		Street:     address.Street,
// 		Number:     address.Number,
// 		City:       address.City,
// 		State:      address.State,
// 		Country:    address.Country,
// 		PostalCode: address.PostalCode,
// 	}

// 	geocoder.ApiKey = g.apiKey
// 	location, err := geocoder.Geocoding(gAddress)
// 	if err != nil {
// 		return Coordinate{}, err
// 	}

// 	return Coordinate{
// 		Lat: location.Latitude,
// 		Lon: location.Longitude,
// 	}, nil
// }
