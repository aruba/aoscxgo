package aoscxgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/exp/slices"
)

type L3Interface struct {

	// Connection properties.
	Interface        Interface              `json:"interface"`
	Description      string                 `json:"description"`
	Ipv4             []interface{}          `json:"ipv4"`
	Ipv6             []interface{}          `json:"ipv6"`
	Vrf              string                 `json:"vrf"`
	InterfaceDetails map[string]interface{} `json:"details"`
	materialized     bool                   `json:"materialized"`
}

// Create performs PATCH to update L3Interface configuration on the given Client object.
func (i *L3Interface) Create(c *Client) error {
	base_uri := "system/interfaces"

	createMap := map[string]interface{}{}

	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L3Interface",
			Err:        errors.New("Create Error"),
		}
	}

	err := i.Interface.checkValues()
	if err != nil {
		return err

	}

	int_str := url.PathEscape(i.Interface.Name)
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	createMap["description"] = i.Description
	createMap["admin"] = i.Interface.AdminState
	createMap["routing"] = true

	if i.Vrf == "" {
		createMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + "default"
	} else {
		createMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + i.Vrf
	}

	//check if it's ipv6
	// Validate ipv4 address

	if len(i.Ipv4) == 0 {
		createMap["ip4_address"] = nil
		createMap["ip4_address_secondary"] = nil
	} else if len(i.Ipv4) == 1 {
		str_ipv4_1 := fmt.Sprintf("%v", i.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			createMap["ip4_address"] = str_ipv4_1
			createMap["ip4_address_secondary"] = nil
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Create Error"),
			}
		}

	} else if len(i.Ipv4) > 1 {
		str_ipv4_1 := fmt.Sprintf("%v", i.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			createMap["ip4_address"] = i.Ipv4[0]
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Create Error"),
			}
		}

		var tmp_splice []string
		for index := 1; index < len(i.Ipv4); index++ {
			str_ipv4_tmp := fmt.Sprintf("%v", i.Ipv4[index])
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
		createMap["ip4_address_secondary"] = tmp_splice
	}

	// string to track ipv6 Create success
	failed_ipv6 := ""

	if len(i.Ipv6) == 0 {
		// What are default values when no ipv6 but routing enabled
		createMap["ip6_addresses"] = nil
	} else if len(i.Ipv6) > 0 {
		// two values for ipv6? how is that affected
		// first create POST to add an ipv6 address Object
		//
		for _, ip_address := range i.Ipv6 {
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

					ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "/" + "ip6_addresses"

					res := post(c, ip6_url, json_body)

					if res.StatusCode != http.StatusCreated {
						failed_ipv6 = fmt.Sprintf("\nip6_addresses failed to create %v\nstatus code %v", i.Ipv6, res.Status)
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

	if i.Interface.AdminState == "down" || i.Interface.AdminState == "" {
		createMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if i.Interface.AdminState == "up" {
		createMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	// Make sure Interface exists in table before patching L3 attributes
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

	patchBody, _ := json.Marshal(createMap)

	json_body := bytes.NewBuffer(patchBody)

	res := patch(c, url, json_body)

	if res.StatusCode != http.StatusNoContent && failed_ipv6 != "" {
		// Combine error messages
		status_str := "L3Interface Create Failed\nstatus " + res.Status
		status_str = status_str + failed_ipv6
		return &RequestError{
			StatusCode: status_str,
			Err:        errors.New("Create Error"),
		}
	} else if res.StatusCode != http.StatusNoContent {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Create Error"),
		}
	} else if failed_ipv6 != "" {
		return &RequestError{
			StatusCode: failed_ipv6,
			Err:        errors.New("Create Error"),
		}
	}

	i.materialized = true

	return nil
}

// Update performs PATCH or PUT to update L3Interface configuration on the given Client object.
func (i *L3Interface) Update(c *Client, use_put bool) error {
	base_uri := "system/interfaces"

	updateMap := map[string]interface{}{}
	int_str := url.PathEscape(i.Interface.Name)

	if use_put {
		tmp_l3_int := L3Interface{Interface: i.Interface}
		err := tmp_l3_int.Get(c)
		if err != nil {
			return err
		}

		for key, value := range tmp_l3_int.InterfaceDetails {
			updateMap[key] = value
		}
	}

	if i.Vrf == "" {
		updateMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + "default"
	} else {
		updateMap["vrf"] = "/rest/" + c.Version + "/system/vrfs/" + i.Vrf
	}

	//check if it's ipv6
	// Validate ipv4 address

	if len(i.Ipv4) == 0 {
		updateMap["ip4_address"] = nil
		updateMap["ip4_address_secondary"] = nil
	} else if len(i.Ipv4) == 1 {
		str_ipv4_1 := fmt.Sprintf("%v", i.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			updateMap["ip4_address"] = str_ipv4_1
			updateMap["ip4_address_secondary"] = nil
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

	} else if len(i.Ipv4) > 1 {
		str_ipv4_1 := fmt.Sprintf("%v", i.Ipv4[0])
		if checkIPAddress(str_ipv4_1) {
			updateMap["ip4_address"] = i.Ipv4[0]
		} else {
			status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_1
			return &RequestError{
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

		var tmp_splice []string
		for index := 1; index < len(i.Ipv4); index++ {
			str_ipv4_tmp := fmt.Sprintf("%v", i.Ipv4[index])
			if checkIPAddress(str_ipv4_tmp) {
				tmp_splice = append(tmp_splice, str_ipv4_tmp)
			} else {
				status_str := "Invalid Required Value: Ipv4 - ensure addresses are in ipv4 format: " + str_ipv4_tmp
				return &RequestError{
					StatusCode: status_str,
					Err:        errors.New("Update Error"),
				}
			}
		}
		updateMap["ip4_address_secondary"] = tmp_splice
	}

	if len(i.Ipv6) == 0 {
		// What are default values when no ipv6 but routing enabled
		updateMap["ip6_addresses"] = nil
		// retrieve what's existing on the switch and remove
		ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "/" + "ip6_addresses"
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
	} else if len(i.Ipv6) > 0 {
		// two values for ipv6? how is that affected
		// first create POST to add an ipv6 address Object
		//
		ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "/" + "ip6_addresses"

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
		for _, ip_address := range i.Ipv6 {
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

	if i.Interface.AdminState == "down" || i.Interface.AdminState == "" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "down"}
	} else if i.Interface.AdminState == "up" {
		updateMap["user_config"] = map[string]interface{}{
			"admin": "up"}
	}

	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface.Name unable to configure L3Interface",
			Err:        errors.New("Update Error"),
		}
	}

	err := i.Interface.checkValues()
	if err != nil {
		return err

	}

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	updateMap["description"] = i.Description
	updateMap["admin"] = i.Interface.AdminState
	updateMap["routing"] = true
	updateMap["vlan_mode"] = nil
	updateMap["vlan_tag"] = nil

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
		res := put(c, url, json_body)
		if res.StatusCode != http.StatusOK {
			status_str := string(updateBody) + "\nPUT status code = " + res.Status
			return &RequestError{
				//StatusCode: res.Status,
				StatusCode: status_str,
				Err:        errors.New("Update Error"),
			}
		}

	} else {
		res := patch(c, url, json_body)
		if res.StatusCode != http.StatusNoContent {
			status_str := string(updateBody) + "\nPATCH status code = " + res.Status
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

// Delete performs PUT to remove/default L3Interface configuration from the given Client object.
func (i *L3Interface) Delete(c *Client) error {
	base_uri := "system/interfaces"
	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to delete L3Interface",
			Err:        errors.New("Delete Error"),
		}
	}
	int_str := url.PathEscape(i.Interface.Name)

	putMap := map[string]interface{}{}

	putBody, _ := json.Marshal(putMap)

	json_body := bytes.NewBuffer(putBody)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str

	res := put(c, url, json_body)

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Delete Error"),
		}
	}

	return nil
}

// Get performs GET to retrieve L3Interface configuration from the given Client object.
func (i *L3Interface) Get(c *Client) error {
	base_uri := "system/interfaces"
	if i.Interface.Name == "" {
		return &RequestError{
			StatusCode: "Missing Interface unable to configure L3Interface",
			Err:        errors.New("Create Error"),
		}
	}
	int_str := url.PathEscape(i.Interface.Name)

	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "?selector=writable"

	res, body := get(c, url)

	if res.StatusCode != http.StatusOK {
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

		if key == "admin" && value != nil {
			i.Interface.AdminState = value.(string)
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
			i.Ipv4 = tmp_splice
		}

		if key == "vrf" && value != nil {
			for key, _ := range value.(map[string]interface{}) {
				i.Vrf = key
			}
		}

	}

	// Include a GET for ip6 and populate .ipv6 attribute

	ip6_url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri + "/" + int_str + "/" + "ip6_addresses"

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

	i.Ipv6 = ip6_addresses

	i.materialized = true

	return nil
}

// GetStatus returns True if L3Interface exists on Client object or False if not.
func (i *L3Interface) GetStatus() bool {
	return i.materialized
}

func checkIPAddress(ip string) bool {
	if strings.Contains(ip, "/") {
		_, _, err := net.ParseCIDR(ip)
		if err == nil {
			return true
		} else {
			return false
		}
	}
	if net.ParseIP(ip) == nil {
		return false
	} else {
		return true
	}
}
