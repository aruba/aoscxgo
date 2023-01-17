package aoscxgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// A custom error struct
type RequestError struct {
	StatusCode string

	Err error
}

// A custom error string
func (r *RequestError) Error() string {
	return fmt.Sprintf("status %d: err %v", r.StatusCode, r.Err)
}

// delete performs DELETE to the given URL and returns the response.
func delete(http_transport *http.Transport, cookie *http.Cookie, url string) *http.Response {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}

// get performs GET to the given URL and returns the data response.
func get(http_transport *http.Transport, cookie *http.Cookie, url string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false
	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	if err != nil {
		fmt.Println(err)
	}

	body := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&body)

	return res, body
}

// post performs POST to the given URL and returns the response.
func post(http_transport *http.Transport, cookie *http.Cookie, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("POST", url, json_body)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
		//os.Exit(1)
	}

	return res
}

// put performs PUT to the given URL and returns the response.
func put(http_transport *http.Transport, cookie *http.Cookie, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("PUT", url, json_body)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
		//os.Exit(1)
	}

	return res
}

// patch performs PATCH to the given URL and returns the response.
func patch(http_transport *http.Transport, cookie *http.Cookie, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("PATCH", url, json_body)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}
