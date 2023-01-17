package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
)

type Vlan struct {

	// Connection properties.
	VlanId       int                    `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	AdminState   string                 `json:"admin_state"`
	VlanDetails  map[string]interface{} `json:"details"`
	materialized bool                   `json:"materialized"`
	uri          string                 `json:"uri"`
}

// Create performs POST to create VLAN configuration on the given Client object.
func (v *Vlan) Create(c *Client) error {
	base_uri := "system/vlans"
	vlan_str := strconv.Itoa(v.VlanId)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri
	v.uri = "/rest/" + c.Version + "/" + base_uri + "/" + vlan_str

	if v.VlanId == 0 || v.Name == "" {
		return &RequestError{
			StatusCode: "Missing Required Values VlanId & Name",
			Err:        errors.New("Create Error"),
		}
	}

	postMap := map[string]interface{}{
		"id":   v.VlanId,
		"name": v.Name,
		"type": "static", //default value
	}

	if v.Description != "" {
		postMap["description"] = v.Description
	}
	if v.AdminState != "" {
		postMap["admin"] = v.AdminState
	}

	postBody, _ := json.Marshal(postMap)

	json_body := bytes.NewBuffer(postBody)

	res := post(c.Transport, c.Cookie, url, json_body)

	if res.Status != "201 Created" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Create Error"),
		}
	}

	v.materialized = true

	return nil
}

// Update performs PATCH to update VLAN configuration on the given Client object.
func (v *Vlan) Update(c *Client) error {
	base_uri := "system/vlans"
	vlan_str := strconv.Itoa(v.VlanId)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_str

	if v.VlanId == 0 || v.Name == "" {
		return &RequestError{
			StatusCode: "Missing Required Values VlanId & Name",
			Err:        errors.New("Update Error"),
		}
	}
	patchMap := map[string]interface{}{
		"name":        v.Name,
		"description": v.Description,
		"admin":       v.AdminState,
		"type":        "static", //default value
	}

	patchBody, _ := json.Marshal(patchMap)

	json_body := bytes.NewBuffer(patchBody)

	res := patch(c.Transport, c.Cookie, url, json_body)

	if res.Status != "204 No Content" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Update Error"),
		}
	}

	return nil
}

// Delete performs DELETE to remove VLAN configuration from the given Client object.
func (v *Vlan) Delete(c *Client) error {
	base_uri := "system/vlans"
	vlan_str := strconv.Itoa(v.VlanId)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_str
	res := delete(c.Transport, c.Cookie, url)

	if res.Status != "204 No Content" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Delete Error"),
		}
	}

	return nil
}

// Get performs GET to retrieve VLAN configuration for the given Client object.
func (v *Vlan) Get(c *Client) error {
	base_uri := "system/vlans"
	vlan_str := strconv.Itoa(v.VlanId)
	v.uri = "/rest/" + c.Version + "/" + base_uri + "/" + vlan_str

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_str + ""

	res, body := get(c.Transport, c.Cookie, url)

	if res.Status != "200 OK" {
		v.materialized = false
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Retrieval Error"),
		}
	}

	if v.VlanDetails == nil {
		v.VlanDetails = map[string]interface{}{}
	}

	for key, value := range body {
		v.VlanDetails[key] = value
		if key == "name" && value != nil {
			v.Name = value.(string)
		}
		if key == "description" && value != nil {
			v.Description = value.(string)
		}

		if key == "admin" {
			v.AdminState = value.(string)
		}

	}

	v.materialized = true

	return nil
}

// GetStatus returns True if VLAN exists on Client object or False if not.
func (v *Vlan) GetStatus() bool {
	return v.materialized
}

// GetURI returns URI of VLAN.
func (v *Vlan) GetURI() string {
	return v.uri
}
