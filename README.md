aoscxgo
========================

aoscxgo is a golang package that allows users to connect to and configure AOS-CX switches using REST API. The minimum supported firmware version is 10.09.

Using aoscxgo
===========

To login to the switch and create a client connection:

```go
package main

import (
	"log"

	"github.com/aruba/aoscxgo"
)

func main() {
	sw, err := aoscxgo.Connect(
		&aoscxgo.Client{
			Hostname:          "10.0.0.1",
			Username:          "admin",
			Password:          "admin",
			VerifyCertificate: false,
		},
	)

	if (sw.Cookie == nil) || (err != nil) {
		log.Printf("Failed to login to switch: %s", err)
		return
	}
	log.Printf("Login Success")

}

```

This will login to the switch and create a cookie to use for authentication in further calls. This cookie is stored within the aoscxgo.Client object that will be passed into configuration modules like so:

```go
	vlan100 := aoscxgo.Vlan{
		VlanId:      100,
		Name:        "uplink VLAN",
		Description: "uplink VLAN",
		AdminState:  "up",
	}

	// if the vlan exists use
	// err = vlan100.Update(sw)
	err = vlan100.Create(sw)

	if err != nil {
		log.Printf("Error in creating VLAN 100: %s", err)
		return
	}

	log.Printf("VLAN Create Success")
```

Each API resource will have the following functions (exceptions may vary):

  * `Create()`
  * `Update()`
  * `Get()`
  * `GetStatus()`
  * `Delete()`
