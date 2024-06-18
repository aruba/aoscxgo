package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/exp/slices"
)

type VlanInterface struct {

	// Connection properties.
	Vlan             Vlan                   `json:"vlan"`
	Description      string                 `json:"description"`
	Ipv4             []interface{}          `json:"ipv4"`
	Ipv6             []interface{}          `json:"ipv6"`
	Vrf              string                 `json:"vrf"`
	InterfaceDetails map[string]interface{} `json:"details"`
	materialized     bool                   `json:"materialized"`
}

// Create performs POST to create VlanInterface configuration on the given Client object.
func (v *VlanInterface) Create(c *Client) error {
	base_uri := "system/interfaces"

	vlan_interface_id := fmt.Sprintf("vlan%d", v.Vlan.VlanId)

	postMap := map[string]interface{}{}

	if v.Vlan.VlanId == 0 {
		return &RequestError{
			StatusCode: "Missing Required Values VlanId",
			Err:        errors.New("Create Error"),
		}
	}

	// Retrieve VLAN from sw if existing
	tmp_vlan := Vlan{
		VlanId: v.Vlan.VlanId,
	}

	err := tmp_vlan.Get(c)

	if err != nil {
		return &RequestError{
			StatusCode: fmt.Sprintf("Missing VLAN %d - Create Vlan before VlanInterface", v.Vlan.VlanId),
			Err:        errors.New("Create Error"),
		}
	}

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri

	postMap["description"] = v.Description
	postMap["admin"] = v.Vlan.AdminState
	postMap["name"] = vlan_interface_id
	postMap["type"] = "vlan"
	postMap["interfaces"] = []string{fmt.Sprintf("/rest/%s/system/vlans/%s", c.Version, strconv.Itoa(v.Vlan.VlanId))}

	if v.Vrf == "" {
		postMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + "default"
	} else {
		postMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + v.Vrf
	}

	//check if it's ipv6
	// Validate ipv4 address

	if len(v.Ipv4) == 0 {
		postMap["ip4_address"] = nil
		postMap["ip4_address_secondary"] = nil
	} else if len(v.Ipv4) == 1 {
		str_ipv4_1 := fmt.Sprintf("%v", v.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			postMap["ip4_address"] = str_ipv4_1
			postMap["ip4_address_secondary"] = nil
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in same ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Create Error"),
			}
		}
	} else if len(v.Ipv4) > 1 {
		str_ipv4_1 := fmt.Sprintf("%v", v.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			postMap["ip4_address"] = v.Ipv4[0]
		} else {
			fmt.Println("THIS IS THE VALUE vvv")
			fmt.Println(str_ipv4_1)
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Create Error"),
			}
		}

		var tmp_splice []string
		for index := 1; index < len(v.Ipv4); index++ {
			str_ipv4_tmp := fmt.Sprintf("%v", v.Ipv4[index])
			fmt.Println("THIS IS THE VALUE vvv")
			fmt.Println(str_ipv4_tmp)
			fmt.Println(v.Ipv4)
			if checkIPAddress(str_ipv4_tmp) {
				tmp_splice = append(tmp_splice, str_ipv4_tmp)
			} else {
				status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_tmp
				return &RequestError{
					StatusCode: status_str,
					Err:        errors.New("Create Error"),
				}
			}
		}
		postMap["ip4_address_secondary"] = tmp_splice
	}

	if v.Vlan.AdminState == "down" || v.Vlan.AdminState == "" {
		postMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if v.Vlan.AdminState == "up" {
		postMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	postBody, _ := json.Marshal(postMap)

	json_body := bytes.NewBuffer(postBody)

	res := post(c, url, json_body)

	if res.StatusCode != http.StatusCreated {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Create Error"),
		}
	}

	// string to track ipv6 Create success
	failed_ipv6 := ""

	if len(v.Ipv6) == 0 {
		// What are default values when no ipv6 but routing enabled
		postMap["ip6_addresses"] = nil
	} else if len(v.Ipv6) > 0 {
		// two values for ipv6? how is that affected
		// first create POST to add an ipv6 address Object
		//
		for _, ip_address := range v.Ipv6 {
			if _, ok := ip_address.(string); ok {
				str_ipv6 := fmt.Sprintf("%v", ip_address)
				if checkIPAddress(str_ipv6) {
					ipv6Map := map[string]interface{}{}
					ipv6Map["address"] = str_ipv6
					ipv6Map["type"] = "global-unicast"
					ipv6Map["preferred_lifetime"] = 604800
					ipv6Map["valid_lifetime"] = 2592000
					ipv6Map["node_address"] = true
					ipv6Map["ra_prefix"] = true
					ipv6Map["ra_route"] = false

					ipv6body, _ := json.Marshal(ipv6Map)

					json_body := bytes.NewBuffer(ipv6body)

					ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id + "/" + "ip6_addresses"

					res := post(c, ip6_url, json_body)

					if res.StatusCode != http.StatusCreated {
						failed_ipv6 = fmt.Sprintf("\nip6_addresses failed to create %v\nstatus code %v", v.Ipv6, res.Status)
						return &RequestError{
							StatusCode: failed_ipv6,
							Err:        errors.New("Create Error"),
						}
					}

				} else {
					status_str := "Invalid Required Value: Ipv6 - ensure addresses are in ipv6 address/mask format:" +
						ip_address.(string)
					return &RequestError{
						StatusCode: status_str,
						Err:        errors.New("Create Error"),
					}
				}
			}
		}

	}

	v.materialized = true

	return nil
}

// Update performs PATCH or PUT to update VlanInterface configuration on the given Client object.
func (v *VlanInterface) Update(c *Client, use_put bool) error {
	base_uri := "system/interfaces"

	updateMap := map[string]interface{}{}

	vlan_interface_id := fmt.Sprintf("vlan%d", v.Vlan.VlanId)

	tmp_vlan_int := VlanInterface{Vlan: v.Vlan}

	if use_put {
		err := tmp_vlan_int.Get(c)
		if err != nil {
			error_str := "Missing VlanInterface - " + vlan_interface_id
			return &RequestError{
				StatusCode: error_str,
				Err:        errors.New("Update Error"),
			}
		}
		for key, value := range tmp_vlan_int.InterfaceDetails {
			updateMap[key] = value
		}
	}

	if v.Vrf == "" {
		updateMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + "default"
	} else {
		updateMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + v.Vrf
	}

	//check if it's ipv6
	// Validate ipv4 address

	if len(v.Ipv4) == 0 {
		updateMap["ip4_address"] = nil
		updateMap["ip4_address_secondary"] = nil
	} else if len(v.Ipv4) == 1 {
		str_ipv4_1 := fmt.Sprintf("%v", v.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			updateMap["ip4_address"] = str_ipv4_1
			updateMap["ip4_address_secondary"] = nil
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in poop ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

	} else if len(v.Ipv4) > 1 {
		str_ipv4_1 := fmt.Sprintf("%v", v.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			updateMap["ip4_address"] = v.Ipv4[0]
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in butt ipv4 format: \n" + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

		var tmp_splice []string
		for index := 1; index < len(v.Ipv4); index++ {
			str_ipv4_tmp := fmt.Sprintf("%v", v.Ipv4[index])
			fmt.Println("THIS IS THE VALUE vvv")
			fmt.Println(str_ipv4_tmp)
			fmt.Println(v.Ipv4)
			if checkIPAddress(str_ipv4_tmp) {
				tmp_splice = append(tmp_splice, str_ipv4_tmp)
			} else {
				status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 fdafdas format: " + str_ipv4_tmp
				return &RequestError{
					StatusCode: status_str,
					Err:        errors.New("Update Error"),
				}
			}
		}
		updateMap["ip4_address_secondary"] = tmp_splice
	}

	if len(v.Ipv6) == 0 {
		// What are default values when no ipv6 but routing enabled
		updateMap["ip6_addresses"] = nil
		// retrieve what's existing on the switch and remove
		ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id + "/" + "ip6_addresses"
		res, body := get(c, ip6_url)

		if res.StatusCode != http.StatusOK {
			return &RequestError{
				StatusCode: res.Status,
				Err:        errors.New("Retrieval Error"),
			}
		}
		for key, _ := range body {
			tmp_ip6_str := url.QueryEscape(key)
			del_ip6_url := ip6_url + "/" + tmp_ip6_str
			res := delete(c, del_ip6_url)

			if res.StatusCode != http.StatusNoContent {
				return &RequestError{
					StatusCode: res.Status,
					Err:        errors.New("Delete Error"),
				}
			}
		}
	} else if len(v.Ipv6) > 0 {
		// two values for ipv6? how is that affected
		// first create POST to add an ipv6 address Object
		//
		ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id + "/" + "ip6_addresses"

		// execute GET to retrieve current IPs
		// iterate over the list of IPs
		// delete the ones that aren't provided

		res, body := get(c, ip6_url)

		if res.StatusCode != http.StatusOK {
			return &RequestError{
				StatusCode: res.Status,
				Err:        errors.New("Retrieval Error"),
			}
		}

		var ipv6_slice []string
		// convert i.Ipv6 from []interface{} to []string
		for _, ip_address := range v.Ipv6 {
			if _, ok := ip_address.(string); ok {
				ipv6_slice = append(ipv6_slice, ip_address.(string))
			}
		}

		var get_ipv6_slice []string
		for key, _ := range body {
			if !slices.Contains(ipv6_slice, key) {
				tmp_ip6_str := url.QueryEscape(key)
				del_ip6_url := ip6_url + "/" + tmp_ip6_str
				res := delete(c, del_ip6_url)

				if res.StatusCode != http.StatusNoContent {
					return &RequestError{
						StatusCode: res.Status,
						Err:        errors.New("Delete Error"),
					}
				}

			} else {
				get_ipv6_slice = append(get_ipv6_slice, key)
			}
		}
		for _, str_ipv6 := range ipv6_slice {
			// Only POST ipv6 addresses not existing
			if checkIPAddress(str_ipv6) && !slices.Contains(get_ipv6_slice, str_ipv6) {
				ipv6Map := map[string]interface{}{}
				ipv6Map["address"] = str_ipv6
				ipv6Map["type"] = "global-unicast"
				ipv6Map["preferred_lifetime"] = 604800
				ipv6Map["valid_lifetime"] = 2592000
				ipv6Map["node_address"] = true
				ipv6Map["ra_prefix"] = true
				ipv6Map["ra_route"] = false

				ipv6body, _ := json.Marshal(ipv6Map)

				json_body := bytes.NewBuffer(ipv6body)

				res := post(c, ip6_url, json_body)

				// include logic to check if address is existing?
				if res.StatusCode != http.StatusCreated {
					status_str := "ip6_addresses failed to update " +
						str_ipv6 + " status " + res.Status
					return &RequestError{
						StatusCode: status_str,
						Err:        errors.New("Update Error"),
					}
				}

			} else if !checkIPAddress(str_ipv6) {
				status_str := "Invalid Required Value: Ipv6 - ensure addresses are in ipv6 address/mask format:" +
					str_ipv6
				return &RequestError{
					StatusCode: status_str,
					Err:        errors.New("Update Error"),
				}
			}
		}
	}

	if v.Vlan.AdminState == "down" || v.Vlan.AdminState == "" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if v.Vlan.AdminState == "up" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id

	updateMap["description"] = v.Description
	updateMap["admin"] = v.Vlan.AdminState
	updateMap["routing"] = true
	updateMap["vlan_mode"] = nil
	updateMap["vlan_tag"] = nil

	if v.Vlan.AdminState == "down" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	}
	if v.Vlan.AdminState == "up" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	updateBody, _ := json.Marshal(updateMap)

	json_body := bytes.NewBuffer(updateBody)

	if use_put {
		res := put(c, url, json_body)
		if res.StatusCode != http.StatusOK {
			status_str := url + "||" + string(updateBody) + "PUT status code = " + res.Status
			return &RequestError{
				//StatusCode: res.Status,
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

	} else {
		res := patch(c, url, json_body)
		if res.StatusCode != http.StatusNoContent {
			status_str := string(updateBody) + "PATCH status code = " + res.Status
			return &RequestError{
				//StatusCode: res.Status,
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}
	}

	v.materialized = true

	return nil
}

// Delete performs DELETE to remove VlanInterface configuration from the given Client object.
func (v *VlanInterface) Delete(c *Client) error {
	base_uri := "system/interfaces"
	if v.Vlan.VlanId == 0 {
		return &RequestError{
			StatusCode: "Missing VlanId unable to configure VlanInterface",
			Err:        errors.New("Delete Error"),
		}
	}
	vlan_interface_id := fmt.Sprintf("vlan%d", v.Vlan.VlanId)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id

	res := delete(c, url)

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Delete Error"),
		}
	}

	return nil
}

// Get performs GET to retrieve VlanInterface configuration from the given Client object.
func (v *VlanInterface) Get(c *Client) error {
	base_uri := "system/interfaces"
	if v.Vlan.VlanId == 0 {
		return &RequestError{
			StatusCode: "Missing VlanId unable to configure VlanInterface",
			Err:        errors.New("Get Error"),
		}
	}
	vlan_interface_id := fmt.Sprintf("vlan%d", v.Vlan.VlanId)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id + "?selector=writable"

	res, body := get(c, url)

	if res.StatusCode != http.StatusOK || len(body) <= 1 {
		v.materialized = false
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Retrieval Error"),
		}
	}

	if v.InterfaceDetails == nil {
		v.InterfaceDetails = map[string]interface{}{}
	}

	for key, value := range body {
		v.InterfaceDetails[key] = value
		if key == "description" && value != nil {
			v.Description = value.(string)
			v.Vlan.Description = value.(string)
		}

		if key == "admin" && value != nil {
			v.Vlan.AdminState = value.(string)
		}

		if key == "ip4_address" && value != nil {
			var tmp_splice []interface{}
			tmp_splice = append(tmp_splice, value.(string))
			if val, ok := body["ip4_address_secondary"]; ok {
				tmp_addr := val.([]interface{})
				if len(tmp_addr) > 0 {
					for index := 0; index < len(tmp_addr); index++ {
						tmp_splice = append(tmp_splice, tmp_addr[index])
					}
				}
			}
			v.Ipv4 = tmp_splice
		}

		if key == "vrf" && value != nil {
			for key, _ := range value.(map[string]interface{}) {
				v.Vrf = key
			}
		}

	}

	// Include a GET for ip6 and populate .ipv6 attribute

	ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + vlan_interface_id + "/" + "ip6_addresses"

	res, body = get(c, ip6_url)

	if res.StatusCode != http.StatusOK {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Retrieval Error"),
		}
	}

	var ipv6_slice []string

	for key, _ := range body {
		if key != "" {
			ipv6_slice = append(ipv6_slice, key)
		}
	}

	ip6_addresses := make([]interface{}, len(ipv6_slice))

	if len(ipv6_slice) > 0 {
		for index := 0; index < len(ipv6_slice); index++ {
			ip6_addresses[index] = ipv6_slice[index]
		}
	}

	v.Ipv6 = ip6_addresses

	v.materialized = true

	return nil
}

// GetStatus returns True if VlanInterface exists on Client object or False if not.
func (i *VlanInterface) GetStatus() bool {
	return i.materialized
}
