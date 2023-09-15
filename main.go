package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

// Request structure
type FlightRequest struct {
	ICAO string `json:"icao"`
}

type Flight struct {
	FlightDate string   `json:"flight_date"`
	Departure  Location `json:"departure"`
	Arrival    Location `json:"arrival"`
	Airline    Airline  `json:"airline"`
}

type Location struct {
	Airport string `json:"airport"`
}

type Airline struct {
	Name string `json:"name"`
}

// Response structure
type FlightResponse struct {
	Data []Flight `json:"data"`
}

func main() {

	inputJSON := os.Args[1]

	// Print the input JSON to stdout
	fmt.Println(inputJSON)

	apiKey, exists := os.LookupEnv("AVIATION_KEY")
	if !exists {
		log.Fatal("AVIATION_KEY does not exist.")
	}

	if len(os.Args) < 2 {
		log.Fatal("Expected JSON input as the first argument.")
	}

	flightRequest, err := parseFlightRequest(inputJSON)

	if err != nil {
		log.Fatal(err)
	}

	flightResponse, err := getFlightData(*flightRequest, apiKey)

	if err != nil {
		log.Fatal(err)
	}

	if len(flightResponse.Data) == 0 {
		log.Fatal("No flights found.")
	}

	flight := flightResponse.Data[0]

	fmt.Printf("Flight %s from %s to %s operated by %s on %s\n", flightRequest.ICAO, flight.Departure.Airport, flight.Arrival.Airport, flight.Airline.Name, flight.FlightDate)
}

func parseFlightRequest(inputJSON string) (*FlightRequest, error) {
	var flightRequest FlightRequest
	err := json.Unmarshal([]byte(inputJSON), &flightRequest)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse JSON: %v", err)
	}

	// Regex to match 2 to 4 letters followed by 1 to 4 numbers
	validICAO := regexp.MustCompile(`^[A-Za-z]{2,4}\d{1,4}$`)

	// Validate ICAO
	if !validICAO.MatchString(flightRequest.ICAO) {
		return nil, fmt.Errorf("Invalid ICAO format. ICAO should consist of 2 to 4 letters followed by 1 to 4 numbers. Your input was: %s", flightRequest.ICAO)
	}

	return &flightRequest, nil
}

var httpClient = &http.Client{}

var sendRequest = func(url string) (*http.Response, error) {
	return httpClient.Get(url)
}

func getFlightData(flightRequest FlightRequest, apiKey string) (*FlightResponse, error) {

	baseURL := "http://api.aviationstack.com/v1/flights"
	finalURL := fmt.Sprintf("%s?access_key=%s&flight_icao=%s", baseURL, apiKey, url.QueryEscape(flightRequest.ICAO))

	resp, err := sendRequest(finalURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch flight data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read flight data: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error from Aviation Stack: %s", body)
	}

	var flightResponse FlightResponse
	err = json.Unmarshal(body, &flightResponse)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode flight data: %v", err)
	}

	return &flightResponse, nil
}
