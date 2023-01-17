package aoscxgo

import (
	"crypto/tls"
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
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	res, err := http_transport.RoundTrip(req)
	if res.Status != "200 OK" {
		log.Fatalf("Got error while connecting to switch %s Error %s", res.Status, err)
		return nil, err
	}

	fmt.Println("Login Successful")
	cookie := res.Cookies()[0]

	return cookie, err
}

// logout performs POST to logout using a cookie from the given URL.
func logout(http_transport *http.Transport, cookie *http.Cookie, url string) *http.Response {
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false

	req.AddCookie(cookie)
	res, err := http_transport.RoundTrip(req)
	//Handle Error
	if res.Status != "200 OK" {
		log.Fatalf("Got error while logging out of switch %s Error %s", res.Status, err)
	}

	fmt.Println("Logout Successful")

	return res
}
