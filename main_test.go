package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
)

type MockHTTPClient struct {
	Response *http.Response
	Err      error
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.Response, m.Err
}

func TestParseFlightRequest(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput *FlightRequest
		expectedError  string
	}{
		{"{\"icao\":\"AB1234\"}", &FlightRequest{ICAO: "AB1234"}, ""},
		{"{\"icao\":\"1234\"}", nil, "Invalid ICAO format. ICAO should consist of 2 to 4 letters followed by 1 to 4 numbers. Your input was: 1234"},
		{"{\"icao\":\"ABCDE1234\"}", nil, "Invalid ICAO format. ICAO should consist of 2 to 4 letters followed by 1 to 4 numbers. Your input was: ABCDE1234"},
		{"{}", nil, "Invalid ICAO format. ICAO should consist of 2 to 4 letters followed by 1 to 4 numbers. Your input was: "},
	}

	for _, test := range tests {
		output, err := parseFlightRequest(test.input)
		if err != nil && err.Error() != test.expectedError {
			t.Errorf("For input %s, expected error %s, but got %s", test.input, test.expectedError, err.Error())
		}
		if err == nil && test.expectedError != "" {
			t.Errorf("For input %s, expected error %s, but got nil", test.input, test.expectedError)
		}
		if test.expectedOutput != nil && *output != *test.expectedOutput {
			t.Errorf("For input %s, expected output %v, but got %v", test.input, test.expectedOutput, output)
		}
	}
}
func TestGetFlightData(t *testing.T) {
	tests := []struct {
		mockResponse *http.Response
		mockError    error
		expectedErr  string
	}{
		{
			&http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"data\":[]}"))),
			},
			nil,
			"",
		},
		{
			&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("Internal Server Error"))),
			},
			nil,
			"Error from Aviation Stack: Internal Server Error",
		},
		{
			nil,
			errors.New("mock error"),
			"Failed to fetch flight data: mock error",
		},
	}

	originalSendRequest := sendRequest
	defer func() {
		sendRequest = originalSendRequest
	}()

	for _, test := range tests {
		sendRequest = func(url string) (*http.Response, error) {
			return test.mockResponse, test.mockError
		}

		_, err := getFlightData(FlightRequest{ICAO: "AB1234"}, "fakeAPIKey")
		if err != nil && err.Error() != test.expectedErr {
			t.Errorf("Expected error %s, but got %s", test.expectedErr, err.Error())
		}
		if err == nil && test.expectedErr != "" {
			t.Errorf("Expected error %s, but got nil", test.expectedErr)
		}
	}
}
