package aoscxgo

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Client struct {
	// Connection properties.
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
	Version  string `json:"version"`
	// Generated after Connect
	Cookie *http.Cookie `json:"cookie"`
	// HTTP transport options.  Note that the VerifyCertificate setting is
	// only used if you do not specify a HTTP transport yourself.
	VerifyCertificate bool            `json:"verify_certificate"`
	Transport         *http.Transport `json:"-"`
}

// Connect creates connection to given Client object.
func Connect(c *Client) (*Client, error) {
	var err error

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !c.VerifyCertificate},
	}

	if c.Version == "" || (c.Version != "v10.09" && c.Version == "v10.10") {
		c.Version = "v10.09"
	}

	if c.Transport == nil {
		c.Transport = tr
	}

	cookie, err := login(c.Transport, c.Hostname, c.Version, c.Username, c.Password)

	if err != nil {
		return nil, err
	}
	c.Cookie = cookie

	return c, err
}

// login performs POST to create a cookie for authentication to the given IP with the provided credentials.
func login(http_transport *http.Transport, ip string, rest_version string, username string, password string) (*http.Cookie, error) {
	url := fmt.Sprintf("https://%s/rest/%s/login?username=%s&password=%s", ip, rest_version, username, password)
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	res, err := http_transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("Got error while connecting to switch %s Error %s\n", res.Status, err)
		return nil, err
	}
	cookie := res.Cookies()[0]

	return cookie, err
}

// logout performs POST to logout using a cookie from the given URL.
func logout(http_transport *http.Transport, cookie *http.Cookie, url string) *http.Response {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if res.StatusCode != http.StatusOK {
		log.Printf("Got error while logging out of switch %s Error %s\n", res.Status, err)
	}
	return res
}

// Logout calls the logout endpoint to clear the session.
func (c *Client) Logout() error {
	if c == nil {
		return errors.New("nil value to Logout")
	}
	url := fmt.Sprintf("https://%s/rest/%s/logout", c.Hostname, c.Version)
	resp := logout(c.Transport, c.Cookie, url)
	if resp.StatusCode != http.StatusOK {
		return errors.New(resp.Status)
	}
	return nil
}
