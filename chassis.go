package aoscxgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Chassis represents the result from the chassis subsystem on v10.09.
type Chassis struct {
	AclInitStatus string `json:"acl_init_status"`
	AdminState    string `json:"admin_state"`
	AsicInfo      struct {
	} `json:"asic_info"`
	Buttons      string   `json:"buttons"`
	Capabilities []string `json:"capabilities"`
	Capacities   struct {
		PsuSlots int `json:"psu_slots"`
	} `json:"capacities"`
	Daemons                    string `json:"daemons"`
	DataPlaneConnectivityState struct {
	} `json:"data_plane_connectivity_state"`
	DataPlaneConnectivityTargetState struct {
	} `json:"data_plane_connectivity_target_state"`
	DataPlaneError struct {
	} `json:"data_plane_error"`
	DataPlaneState             interface{} `json:"data_plane_state"`
	DataPlaneTargetState       interface{} `json:"data_plane_target_state"`
	DataPlanes                 string      `json:"data_planes"`
	DiagTestResults            string      `json:"diag_test_results"`
	DiagnosticDisable          bool        `json:"diagnostic_disable"`
	DiagnosticLastRunTimestamp int         `json:"diagnostic_last_run_timestamp"`
	DiagnosticPerformed        int         `json:"diagnostic_performed"`
	DiagnosticRequested        int         `json:"diagnostic_requested"`
	EntityState                struct {
	} `json:"entity_state"`
	FanConfigurationState interface{} `json:"fan_configuration_state"`
	Fans                  string      `json:"fans"`
	Interfaces            struct {
	} `json:"interfaces"`
	Leds             string      `json:"leds"`
	MacsRemaining    int         `json:"macs_remaining"`
	Name             string      `json:"name"`
	NextMacAddress   string      `json:"next_mac_address"`
	PacGbpInitStatus string      `json:"pac_gbp_init_status"`
	PartNumberCfg    interface{} `json:"part_number_cfg"`
	PoePower         struct {
		AvailablePower              int `json:"available_power"`
		FailoverPower               int `json:"failover_power"`
		PowerStatusRefreshTimestamp int `json:"power_status_refresh_timestamp"`
		PowerStatusUpdateTimestamp  int `json:"power_status_update_timestamp"`
		RedundantPower              int `json:"redundant_power"`
	} `json:"poe_power"`
	PolicyInitStatus string      `json:"policy_init_status"`
	PowerConsumed    interface{} `json:"power_consumed"`
	PowerSupplies    string      `json:"power_supplies"`
	ProductInfo      struct {
		BaseMacAddress     string `json:"base_mac_address"`
		DeviceVersion      string `json:"device_version"`
		Instance           string `json:"instance"`
		NumberOfMacs       string `json:"number_of_macs"`
		PartNumber         string `json:"part_number"`
		ProductDescription string `json:"product_description"`
		ProductName        string `json:"product_name"`
		SerialNumber       string `json:"serial_number"`
		Vendor             string `json:"vendor"`
	} `json:"product_info"`
	PsuRedundancyOper string `json:"psu_redundancy_oper"`
	PsuRedundancySet  string `json:"psu_redundancy_set"`
	RebootStatistics  struct {
		Configuration int `json:"configuration"`
		Error         int `json:"error"`
		Hotswap       int `json:"hotswap"`
		Isp           int `json:"isp"`
		Thermal       int `json:"thermal"`
		User          int `json:"user"`
	} `json:"reboot_statistics"`
	ResetsPerformed struct {
	} `json:"resets_performed"`
	ResetsRequested struct {
	} `json:"resets_requested"`
	ResourceCapacity struct {
	} `json:"resource_capacity"`
	ResourceReservationPerFeature struct {
	} `json:"resource_reservation_per_feature"`
	ResourceUnreserved struct {
	} `json:"resource_unreserved"`
	ResourceUtilization struct {
	} `json:"resource_utilization"`
	ResourceUtilizationPerFeature struct {
	} `json:"resource_utilization_per_feature"`
	ResourceWidthPerFeature struct {
	} `json:"resource_width_per_feature"`
	Selftest struct {
		Status string `json:"status"`
	} `json:"selftest"`
	SelftestDisable bool `json:"selftest_disable"`
	SoftwareImages  struct {
	} `json:"software_images"`
	State   string `json:"state"`
	Storage struct {
	} `json:"storage"`
	TempSensors  string `json:"temp_sensors"`
	ThermalState string `json:"thermal_state"`
	Type         string `json:"type"`
	UsbStatus    string `json:"usb_status"`
}

// GetChassis returns the chassis information by the given id. The first chassis has id 1.
func (c *Client) GetChassis(id int) (Chassis, error) {
	url := fmt.Sprintf("https://%s/rest/%s/system/subsystems/chassis,%v", c.Hostname, c.Version, id)
	var result Chassis
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("accept", "*/*")
	req.Close = false
	req.AddCookie(c.Cookie)
	res, err := c.Transport.RoundTrip(req)
	if err == nil {
		err = json.NewDecoder(res.Body).Decode(&result)
	}
	return result, err
}
