aoscxgo
========================

aoscxgo is a golang package that allows users to connect to and configure AOS-CX switches using REST API. The minimum supported firmware version is 10.09.

This package is forked from [Arubas own aoscxgo](https://github.com/aruba/aoscxgo) with some improvements.

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

This will login to the switch and create a cookie to use for authentication in further calls. This cookie is stored within the aoscxgo.

### Work with VLAN's

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

### Get some switch information

```go
    result, err := sw.GetChassis(1)
    if err == nil {
        log.Printf("Model %s, serial %s\n", result.ProductInfo.ProductName, result.ProductInfo.SerialNumber)
    }
```