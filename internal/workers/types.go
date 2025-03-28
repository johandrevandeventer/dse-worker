package workers

import (
	"time"

	"github.com/google/uuid"
)

type IgnoredControllersAndDevices struct {
	IgnoredControllers []string `json:"ignored_controllers"`
	IgnoredDevices     []string `json:"ignored_devices"`
}

type DataStruct struct {
	State                string
	CustomerID           uuid.UUID
	CustomerName         string
	SiteID               uuid.UUID
	SiteName             string
	Gateway              string
	Controller           string
	DeviceType           string
	ControllerIdentifier string
	DeviceName           string
	DeviceIdentifier     string
	Data                 map[string]any
	Timestamp            time.Time
}
