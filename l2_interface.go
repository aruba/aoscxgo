package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
)

type L2Interface struct {

	// Connection properties.
	Interface        Interface              `json:"interface"`
	Description      string                 `json:"description"`
	VlanMode         string                 `json:"vlan_mode"`
	VlanIds          []interface{}          `json:"vlan_ids"`
	VlanTag          int                    `json:"vlan_tag"`
	TrunkAllowedAll  bool                   `json:"trunk_allowed_all"`
	NativeVlanTag    bool                   `json:"native_vlan_tag"`
	InterfaceDetails map[string]interface{} `json:"details"`
	materialized     bool                   `json:"materialized"`
}

// Create performs PATCH to update L2Interface configuration on the given Client object.
func (i *L2Interface) Create(c *Client) error {
	base_uri := "system/interfaces"

	patchMap := map[string]interface{}{}

	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L2Interface",
			Err:        errors.New("Create Error"),
		}
	}

	err := i.Interface.checkValues()
	if err != nil {
		return err

	}

	int_str := url.PathEscape(i.Interface.Name)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	if i.VlanMode == "access" || i.VlanMode == "" {
		if i.VlanTag == 0 {
			i.VlanTag = 1
		}

		tmp_vlan := Vlan{
			VlanId: i.VlanTag,
		}

		err := tmp_vlan.Get(c)

		if err != nil && !tmp_vlan.materialized {
			err = tmp_vlan.Create(c)
			if err != nil && !tmp_vlan.materialized {
				return &RequestError{
					StatusCode: "Vlan Not found unable to configure L2Interface",
					Err:        errors.New("Create Error"),
				}
			}
		}
		patchMap["vlan_tag"] = map[string]interface{}{strconv.Itoa(i.VlanTag): tmp_vlan.GetURI()}
		patchMap["vlan_mode"] = "access"

	} else if i.VlanMode == "trunk" || i.VlanMode == "native-untagged" || i.VlanMode == "native-tagged" {

		if i.NativeVlanTag {
			i.VlanMode = "native-tagged"
		} else {
			i.VlanMode = "native-untagged"
		}

		if i.VlanTag == 1 {
			patchMap["vlan_tag"] = nil
		} else {
			tmp_vlan := Vlan{
				VlanId: i.VlanTag,
			}

			err := tmp_vlan.Get(c)

			if err != nil && !tmp_vlan.materialized {
				return &RequestError{
					StatusCode: "Vlan Not found unable to configure L2Interface",
					Err:        errors.New("Create Error"),
				}
			}
			patchMap["vlan_tag"] = map[string]interface{}{
				strconv.Itoa(tmp_vlan.VlanId): tmp_vlan.GetURI(),
			}
		}

		vlan_trunks := map[string]interface{}{}

		if !i.TrunkAllowedAll {
			// Test what is behavior of List being empty or not
			for _, item := range i.VlanIds {
				tmp_vlan_obj := Vlan{VlanId: item.(int)}
				err = tmp_vlan_obj.Get(c)
				if err == nil {
					vlan_trunks[strconv.Itoa(tmp_vlan_obj.VlanId)] = tmp_vlan_obj.GetURI()
				}

			}

		} else {
			vlan_trunks = map[string]interface{}{}
		}
		patchMap["vlan_trunks"] = vlan_trunks
		patchMap["vlan_mode"] = i.VlanMode
	} else {
		status_str := "Invalid Required Value: VlanMode - valid options are 'access' or 'trunk' received: " + i.VlanMode
		return &RequestError{
			StatusCode: status_str,
			Err:        errors.New("Create Error"),
		}
	}

	// Make sure Interface exists in table before patching L2 attributes
	tmp_int := Interface{
		Name: i.Interface.Name,
	}

	err = tmp_int.Get(c)

	if err != nil {
		return err

	} else if !tmp_int.materialized {
		err = i.Interface.Create(c)

		if err != nil {
			return err

		}
	}

	patchMap["description"] = i.Description
	patchMap["admin"] = i.Interface.AdminState
	patchMap["routing"] = false

	if i.Interface.AdminState == "down" {
		patchMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if i.Interface.AdminState == "up" {
		patchMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	patchBody, _ := json.Marshal(patchMap)

	json_body := bytes.NewBuffer(patchBody)

	res := patch(c.Transport, c.Cookie, url, json_body)

	if res.Status != "204 No Content" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Create Error"),
		}
	}

	i.materialized = true

	return nil
}

// Update performs PATCH or PUT to update L2Interface configuration on the given Client object.
func (i *L2Interface) Update(c *Client, use_put bool) error {
	base_uri := "system/interfaces"

	updateMap := map[string]interface{}{}

	if use_put {
		tmp_l2_int := L2Interface{Interface: i.Interface}
		err := tmp_l2_int.Get(c)
		if err != nil {
			return err
		}

		for key, value := range tmp_l2_int.InterfaceDetails {
			updateMap[key] = value
		}
	}

	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L2Interface",
			Err:        errors.New("Create Error"),
		}
	}

	err := i.Interface.checkValues()
	if err != nil {
		return err

	}

	int_str := url.PathEscape(i.Interface.Name)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	if i.VlanMode == "access" || i.VlanMode == "" {
		if i.VlanTag == 0 {
			i.VlanTag = 1
		}

		tmp_vlan := Vlan{
			VlanId: i.VlanTag,
		}

		err := tmp_vlan.Get(c)

		if err != nil && !tmp_vlan.materialized {
			err_str := "Vlan Not found unable to configure L2Interface Stats " + err.(*RequestError).StatusCode + " | " + strconv.FormatBool(tmp_vlan.materialized)
			return &RequestError{
				StatusCode: err_str,
				Err:        errors.New("Create Error"),
			}
		}
		updateMap["vlan_tag"] = map[string]interface{}{strconv.Itoa(i.VlanTag): tmp_vlan.GetURI()}
		updateMap["vlan_mode"] = "access"

	} else if i.VlanMode == "trunk" || i.VlanMode == "native-untagged" || i.VlanMode == "native-tagged" {

		if i.NativeVlanTag {
			i.VlanMode = "native-tagged"
		} else {
			i.VlanMode = "native-untagged"
		}

		if i.VlanTag == 1 || i.VlanTag == 0 {
			updateMap["vlan_tag"] = nil
		} else {
			tmp_vlan := Vlan{
				VlanId: i.VlanTag,
			}

			err := tmp_vlan.Get(c)

			if err != nil && !tmp_vlan.materialized {
				err_str := "Vlan Not found unable to configure L2Interface Stats " + err.(*RequestError).StatusCode + " | " + strconv.Itoa(i.VlanTag)
				return &RequestError{
					StatusCode: err_str,
					Err:        errors.New("Create Error"),
				}
			}
			updateMap["vlan_tag"] = map[string]interface{}{
				strconv.Itoa(tmp_vlan.VlanId): tmp_vlan.GetURI(),
			}
		}

		vlan_trunks := map[string]interface{}{}

		if !i.TrunkAllowedAll {
			// Test what is behavior of List being empty or not
			for _, item := range i.VlanIds {
				tmp_vlan_obj := Vlan{VlanId: item.(int)}
				err = tmp_vlan_obj.Get(c)
				if err == nil {
					vlan_trunks[strconv.Itoa(tmp_vlan_obj.VlanId)] = tmp_vlan_obj.GetURI()
				}

			}
		} else {
			vlan_trunks = map[string]interface{}{}
		}
		updateMap["vlan_trunks"] = vlan_trunks
		updateMap["vlan_mode"] = i.VlanMode

	} else {
		status_str := "Invalid Required Value: VlanMode - valid options are 'access' or 'trunk' received: " + i.VlanMode
		return &RequestError{
			StatusCode: status_str,
			Err:        errors.New("Create Error"),
		}
	}

	updateMap["description"] = i.Description
	updateMap["admin"] = i.Interface.AdminState
	updateMap["routing"] = false

	if i.Interface.AdminState == "down" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	}
	if i.Interface.AdminState == "up" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	updateBody, _ := json.Marshal(updateMap)

	json_body := bytes.NewBuffer(updateBody)

	if use_put {
		res := put(c.Transport, c.Cookie, url, json_body)
		if res.Status != "200 OK" {
			status_str := string(updateBody) + "PUT status code = " + res.Status
			return &RequestError{
				//StatusCode: res.Status,
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

	} else {
		res := patch(c.Transport, c.Cookie, url, json_body)
		if res.Status != "204 No Content" {
			status_str := string(updateBody) + "PATCH status code = " + res.Status
			return &RequestError{
				//StatusCode: res.Status,
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}
	}

	i.materialized = true

	return nil
}

// Delete performs PUT to remove/default L2Interface configuration from the given Client object.
func (i *L2Interface) Delete(c *Client) error {
	base_uri := "system/interfaces"
	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L2Interface",
			Err:        errors.New("Create Error"),
		}
	}
	int_str := url.PathEscape(i.Interface.Name)

	putMap := map[string]interface{}{}

	putBody, _ := json.Marshal(putMap)

	json_body := bytes.NewBuffer(putBody)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	//need logic for handling interfaces between platforms

	res := put(c.Transport, c.Cookie, url, json_body)

	if res.Status != "204 No Content" && res.Status != "200 OK" {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Delete Error"),
		}
	}

	return nil
}

// Get performs GET to retrieve L2Interface configuration from the given Client object.
func (i *L2Interface) Get(c *Client) error {
	base_uri := "system/interfaces"
	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L2Interface",
			Err:        errors.New("Create Error"),
		}
	}
	int_str := url.PathEscape(i.Interface.Name)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "?selector=writable"

	res, body := get(c.Transport, c.Cookie, url)

	if res.Status != "200 OK" {
		i.materialized = false
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Retrieval Error"),
		}
	}

	if i.Interface.InterfaceDetails == nil {
		i.Interface.InterfaceDetails = map[string]interface{}{}
	}

	for key, value := range body {
		i.Interface.InterfaceDetails[key] = value
		if key == "description" && value != nil {
			i.Description = value.(string)
			i.Interface.Description = value.(string)
		}

		if key == "vlan_mode" && value != nil {
			i.VlanMode = value.(string)
			if i.VlanMode == "native-tagged" {
				i.NativeVlanTag = true
			} else {
				i.NativeVlanTag = false
			}
		}

		if key == "admin" && value != nil {
			i.Interface.AdminState = value.(string)
		}

		if key == "vlan_tag" && value != nil {
			// convert json to value singular
			// "vlan_tag": {
			// 	"42": "/rest/v10.09/system/vlans/42"
			//   },
			for key, _ := range value.(map[string]interface{}) {
				vlan_int, _ := strconv.Atoi(key)
				i.VlanTag = vlan_int
			}

		}

		if key == "vlan_trunks" && len(value.(map[string]interface{})) != 0 {
			// convert json to value singular
			// "vlan_trunks": {
			// 	"42": "/rest/v10.09/system/vlans/42"
			//   },
			tmp_splice := []interface{}{}
			for key, _ := range value.(map[string]interface{}) {
				vlan_int, _ := strconv.Atoi(key)
				tmp_splice = append(tmp_splice, vlan_int)

			}

			i.VlanIds = tmp_splice
			i.TrunkAllowedAll = false

		} else if key == "vlan_trunks" && len(value.(map[string]interface{})) == 0 {
			i.TrunkAllowedAll = true
			i.VlanIds = []interface{}{}
		}

	}

	i.materialized = true

	return nil
}

// GetStatus returns True if L2Interface exists on Client object or False if not.
func (i *L2Interface) GetStatus() bool {
	return i.materialized
}
