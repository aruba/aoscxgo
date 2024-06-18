package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
)

type Interface struct {

	// Connection properties.
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	AdminState       string                 `json:"admin"`
	InterfaceDetails map[string]interface{} `json:"details"`
	materialized     bool                   `json:"materialized"`
	uri              string                 `json:"uri"`
}

// checkName validates if interface Name is valid or not
func checkName(name string) bool {
	re := "\\d+/\\d+/\\d+"

	found, err := regexp.MatchString(re, name)

	if found && err == nil {
		return true
	}
	return false
}

// checkValues validates if interface Name and AdminState are valid or not
func (i *Interface) checkValues() error {
	if !checkName(i.Name) {
		return &RequestError{
			StatusCode: "Invalid Required Value: Name",
			Err:        errors.New("Create Error"),
		}
	}

	status_str := "Invalid Required Value: AdminState - valid options are 'up' or 'down' received: " + i.AdminState

	if i.AdminState != "down" && i.AdminState != "up" {
		return &RequestError{
			StatusCode: status_str,
			Err:        errors.New("Create Error"),
		}
	}
	return nil
}

// Create performs POST to create Interface configuration on the given Client object.
func (i *Interface) Create(c *Client) error {
	base_uri := "system/interfaces"
	url_str := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri

	int_str := url.PathEscape(i.Name)

	i.uri = "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	err := i.checkValues()
	if err != nil {
		return err

	}

	postMap := map[string]interface{}{
		"name":        i.Name,
		"description": i.Description,
		"admin":       i.AdminState,
	}

	if i.AdminState == "down" {
		postMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if i.AdminState == "up" {
		postMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	postBody, _ := json.Marshal(postMap)

	json_body := bytes.NewBuffer(postBody)

	res := post(c, url_str, json_body)

	if res.Status != "201 Created" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Create Error"),
		}
	}

	i.materialized = true

	return nil
}

// Update performs PATCH to update Interface configuration on the given Client object.
func (i *Interface) Update(c *Client) error {
	base_uri := "system/interfaces"
	err := i.checkValues()
	if err != nil {
		return err

	}

	int_str := url.PathEscape(i.Name)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	patchMap := map[string]interface{}{
		"description": i.Description,
		"admin":       i.AdminState,
	}

	if i.AdminState == "down" {
		patchMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	}
	if i.AdminState == "up" {
		patchMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	patchBody, _ := json.Marshal(patchMap)

	json_body := bytes.NewBuffer(patchBody)

	res := patch(c, url, json_body)

	if res.Status != "204 No Content" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Update Error"),
		}
	}

	return nil
}

// Delete performs PUT to remove/default Interface configuration from the given Client object.
func (i *Interface) Delete(c *Client) error {
	base_uri := "system/interfaces"
	int_str := url.PathEscape(i.Name)

	putMap := map[string]interface{}{}

	putBody, _ := json.Marshal(putMap)

	json_body := bytes.NewBuffer(putBody)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str
	//res := delete(c,  url)

	//need logic for handling interfaces between platforms

	res := put(c, url, json_body)

	if res.Status != "204 No Content" && res.Status != "200 OK" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Delete Error"),
		}
	}

	return nil
}

// Get performs GET to retrieve Interface configuration from the given Client object.
func (i *Interface) Get(c *Client) error {
	base_uri := "system/interfaces"
	int_str := url.PathEscape(i.Name)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + ""

	res, body := get(c, url)

	if res.Status != "200 OK" {
		i.materialized = false
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Retrieval Error"),
		}
	}

	if i.InterfaceDetails == nil {
		i.InterfaceDetails = map[string]interface{}{}
	}

	for key, value := range body {
		i.InterfaceDetails[key] = value
		if key == "description" && value != nil {
			i.Description = value.(string)
		}

		if key == "admin" && value != nil {
			i.AdminState = value.(string)
		}

	}

	i.materialized = true

	return nil
}

// GetStatus returns True if Interface exists on Client object or False if not.
func (i *Interface) GetStatus() bool {
	return i.materialized
}
