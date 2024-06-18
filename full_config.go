package aoscxgo

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/go-cmp/cmp"
)

type FullConfig struct {

	// Connection properties.
	FileName string `json:"filename"`
	//Hash     hash.Hash `json:"hash"`
	Config string `json:"config"`
	uri    string `json:"uri"`
}

// Create performs POST to create VLAN configuration on the given Client object.
func (fc *FullConfig) Create(c *Client) (*http.Response, error) {
	if fc.FileName == "" {
		return nil, &RequestError{
			StatusCode: "Missing FileName",
			Err:        errors.New("Create Error"),
		}
	}

	config_str, err := fc.ReadConfigFile(fc.FileName)

	if err != nil {
		return nil, &RequestError{
			StatusCode: "Error in reading config file",
			Err:        err,
		}
	}

	res, body := fc.ValidateConfig(c, config_str)

	if body == nil {
		return res, &RequestError{
			StatusCode: res.Status,
			Err:        errors.New("Validation Error"),
		}

	} else if body["state"] == "success" {
		res2, body2 := fc.ApplyConfig(c, config_str)
		if body2["state"] != "success" {
			errors_dict := body2["errors"].([]interface{})
			error_str := convert_errors(errors_dict)

			return res2, &RequestError{
				StatusCode: "Error in applying config error : \n" + error_str,
				Err:        errors.New("Apply Error"),
			}
		} else {
			log.Println("New Config Applied Successfully")
			fc.Get(c)
			return res2, nil
		}
	} else if res != nil && body != nil {
		errors_dict := body["errors"].([]interface{})
		errors_dict = errors_dict
		error_str := convert_errors(errors_dict)

		return res, &RequestError{
			StatusCode: "Error in validating config error : \n" + error_str,
			Err:        errors.New("Apply Error"),
		}

	}

	return res, &RequestError{
		StatusCode: res.Status,
		Err:        errors.New("Validation Error"),
	}
}

// Get performs GET to retrieve Running configuration for the given Client object.
func (fc *FullConfig) Get(c *Client) error {
	base_uri := "configs/running-config"
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri
	res, _ := get_accept_text(c, url)

	if res.Status != "200 OK" {
		return &RequestError{
			StatusCode: res.Status + url,
			Err:        errors.New("Retrieval Error"),
		}
	}

	// Read the content
	var bodyBytes []byte
	if res.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(res.Body)
	}
	// Restore the io.ReadCloser to its original state
	res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	bodyString := string(bodyBytes)

	// config := bytes.NewBuffer(nil)
	// config, _ = ioutil.ReadAll(res.Body)
	fc.Config = bodyString

	return nil
}

func (fc *FullConfig) ReadConfigFile(filename string) (string, error) {
	config_contents, err := ioutil.ReadFile(filename)

	if err != nil {
		err_str := "Unable to read file " + filename
		return "", &RequestError{
			StatusCode: err_str,
			Err:        err,
		}
	}
	config_str := string(config_contents)
	return config_str, nil

}

// Formats the errors provided by dryrun for user
func convert_errors(errors_list []interface{}) string {
	errors_str := ""
	for _, error_dict := range errors_list {
		tmp_dict := error_dict.(map[string]interface{})
		line_num := tmp_dict["line"]
		line_float := line_num.(float64)
		line_int := int(line_float)
		errors_str += "line "
		line_num_str := fmt.Sprintf("%d", line_int)
		errors_str += line_num_str
		errors_str += " | "
		errors_str += tmp_dict["message"].(string)
		errors_str += "\n"
	}
	return errors_str
}

// Sets Config Attribute
func (fc *FullConfig) SetConfig(config string) error {
	fc.Config = config

	return nil
}

// Compares supplied string Config to stored Object
func (fc *FullConfig) CompareConfig(new_config string) string {

	return cmp.Diff(fc.Config, new_config)
}

// Compares supplied string Config to stored Object
func (fc *FullConfig) DownloadConfig(c *Client, filename string) error {
	base_uri := "configs/running-config"
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri
	res, _ := get_accept_text(c, url)

	if res.Status != "200 OK" {
		return &RequestError{
			StatusCode: res.Status + url,
			Err:        errors.New("Retrieval Error"),
		}
	}

	// Read the content
	var bodyBytes []byte
	if res.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(res.Body)
	}
	// Restore the io.ReadCloser to its original state
	res.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	// Use the content
	bodyString := string(bodyBytes)

	err := ioutil.WriteFile(filename, []byte(bodyString), 0644)
	if err != nil {
		panic(err)
	}

	return err
}

// Validates supplied CLI configuration as string using dryrun
func (fc *FullConfig) ValidateConfig(c *Client, config string) (*http.Response, map[string]interface{}) {
	base_uri := "configs/running-config"
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri
	dryrun_url := url + "?dryrun=validate"

	json_body := bytes.NewBufferString(config)

	res := post(c, dryrun_url, json_body)

	if res.Status != "200 OK" && res.Status != "202 Accepted" {
		return res, nil
	}

	dryrun_url = url + "?dryrun"

	res2, body := get(c, dryrun_url)

	iterations := 10

	for iterations > 0 {
		iterations -= 1
		if body["state"] == "success" || body["state"] == "error" {
			break
		}
		time.Sleep(2 * time.Second)
		res2, body = get(c, dryrun_url)
	}

	return res2, body
}

func (fc *FullConfig) ApplyConfig(c *Client, config string) (*http.Response, map[string]interface{}) {
	base_uri := "configs/running-config"
	url := "https://" + c.Hostname + "/rest/" + c.Version + "/" + base_uri
	dryrun_url := url + "?dryrun=apply"

	json_body := bytes.NewBufferString(config)

	res := post(c, dryrun_url, json_body)

	if res.Status != "200 OK" && res.Status != "202 Accepted" {
		return res, nil
	}

	dryrun_url = url + "?dryrun"

	res2, body := get(c, dryrun_url)

	iterations := 10

	for iterations > 0 {
		iterations -= 1
		if body["state"] == "success" || body["state"] == "error" {
			break
		}
		time.Sleep(2 * time.Second)
		res2, body = get(c, dryrun_url)
	}

	return res2, body
}
