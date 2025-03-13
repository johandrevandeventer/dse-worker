package gensetworker

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/johandrevandeventer/dse-worker/internal/workers"
	dse890_decoder "github.com/johandrevandeventer/dse-worker/internal/workers/dse_worker/genset/dse890"
	"github.com/johandrevandeventer/kafkaclient/payload"
	"go.uber.org/zap"
)

const (
	ControllerTypeDSE890 = "dse890"
)

func GensetWorker(msg payload.Payload, logger *zap.Logger) (*workers.DataStruct, *workers.DataStruct, error) {
	logger.Debug("Decoding genset data")
	gensetData := map[string]any{}
	rawDataStruct := &workers.DataStruct{}
	processedDataStruct := &workers.DataStruct{}

	// Parse the JSON string into a map
	err := json.Unmarshal([]byte(msg.Message), &gensetData)
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("error unmarshalling genset data: %w", err)
	}

	// Get the controller ID and payload
	var controllerId string
	var payload map[string]any

	for key := range gensetData {
		controllerId = key
		payload = gensetData[key].(map[string]any)
		break
	}

	logger.Debug("Processing controller", zap.String("controllerId", controllerId))

	ignoredControllers, err := workers.GetIgnoredControllers()
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("error getting ignored controllers: %w", err)
	}

	if slices.Contains(ignoredControllers, controllerId) {
		logger.Warn("Controller is ignored", zap.String("controller_id", controllerId))
		return rawDataStruct, processedDataStruct, nil
	}

	logger.Debug("Fetching devices by controller ID", zap.String("controllerId", controllerId))

	devices, err := workers.GetDevicesByControllerSerialNumber(controllerId)
	if err != nil {
		return rawDataStruct, processedDataStruct, fmt.Errorf("error getting devices by controller ID: %w", err)
	}

	if len(devices) == 0 {
		logger.Warn("No devices found for controller", zap.String("controller_id", controllerId))
	}

	for _, device := range devices {
		if device.ControllerSerialNumber != controllerId {
			continue
		}

		timestamp := msg.MessageTimestamp
		controller := device.Controller
		controllerLower := strings.ToLower(controller)

		ignoredDevices, err := workers.GetIgnoredDevices()
		if err != nil {
			return rawDataStruct, processedDataStruct, fmt.Errorf("error getting ignored devices: %w", err)
		}

		for _, ignoredDevice := range ignoredDevices {
			if device.DeviceSerialNumber == ignoredDevice {
				logger.Warn("Device is ignored", zap.String("device_serial_number", device.DeviceSerialNumber))
				continue
			}
		}

		switch controllerLower {
		case ControllerTypeDSE890:
			rawData, processedData, err := dse890_decoder.DSE890Decoder(payload)
			if err != nil {
				return rawDataStruct, processedDataStruct, fmt.Errorf("error decoding UC100 data: %w", err)
			}

			if len(rawData) == 0 {
				logger.Warn("No raw data found", zap.String("controller", controllerId))
				continue
			}

			if len(processedData) == 0 {
				logger.Warn("No processed data found", zap.String("controller", controllerId))
				continue
			}

			rawDataStruct = &workers.DataStruct{
				State:                  "Pre",
				CustomerID:             device.Site.Customer.ID,
				CustomerName:           device.Site.Customer.Name,
				SiteID:                 device.Site.ID,
				SiteName:               device.Site.Name,
				Gateway:                device.Gateway,
				Controller:             device.Controller,
				DeviceType:             device.DeviceType,
				ControllerSerialNumber: device.ControllerSerialNumber,
				DeviceName:             device.DeviceName,
				DeviceSerialNumber:     device.DeviceSerialNumber,
				Data:                   processedData,
				Timestamp:              timestamp,
			}

			processedDataStruct = &workers.DataStruct{
				State:                  "Post",
				CustomerID:             device.Site.Customer.ID,
				CustomerName:           device.Site.Customer.Name,
				SiteID:                 device.Site.ID,
				SiteName:               device.Site.Name,
				Gateway:                device.Gateway,
				Controller:             device.Controller,
				DeviceType:             device.DeviceType,
				ControllerSerialNumber: device.ControllerSerialNumber,
				DeviceName:             device.DeviceName,
				DeviceSerialNumber:     device.DeviceSerialNumber,
				Data:                   processedData,
				Timestamp:              timestamp,
			}

			return rawDataStruct, processedDataStruct, nil
		default:
			logger.Warn("Controller not supported", zap.String("controller", controller))
		}
	}

	return rawDataStruct, processedDataStruct, nil
}
