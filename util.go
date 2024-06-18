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
	return fmt.Sprintf("Status : %v\nError %v", r.StatusCode, r.Err)
}

// delete performs DELETE to the given URL and returns the response.
func delete(client *Client, url string) *http.Response {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}

// get performs GET to the given URL and returns the data response.
func get(client *Client, url string) (*http.Response, map[string]interface{}) {
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("accept", "*/*")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false
	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	if err != nil {
		fmt.Println(err)
	}

	body := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&body)

	return res, body
}

// get performs GET to the given URL and returns the data response.
func get_accept_text(client *Client, url string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "text/plain")
	// req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false
	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	if err != nil {
		fmt.Println(err)
	}

	return res, err
}

// get performs GET to the given URL and returns the data response.
func post(client *Client, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("POST", url, json_body)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false

	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}

// put performs PUT to the given URL and returns the response.
func put(client *Client, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("PUT", url, json_body)
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false

	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}

// patch performs PATCH to the given URL and returns the response.
func patch(client *Client, url string, json_body *bytes.Buffer) *http.Response {
	req, _ := http.NewRequest("PATCH", url, json_body)
	req.Header.Set("accept", "*/*")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false

	req.AddCookie(client.Cookie)
	res, err := client.Transport.RoundTrip(req)
	//Handle Error
	if err != nil {
		log.Fatalf("An Error Occured %v %v", err, res.Status)
	}

	return res
}
