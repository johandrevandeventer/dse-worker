package dse890

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/johandrevandeventer/dse-worker/internal/workers"
	"github.com/johandrevandeventer/dse-worker/internal/workers/dse_worker/dse890/genset"
	"github.com/johandrevandeventer/dse-worker/internal/workers/types"
	"github.com/johandrevandeventer/kafkaclient/payload"
	"go.uber.org/zap"
)

const (
	DeviceTypeGenset = "genset"
)

func Processor(msg payload.Payload, logger *zap.Logger) (MessageInfo *types.MessageInfo, err error) {
	var data map[string]map[string]map[string]any
	if err := json.Unmarshal(msg.Message, &data); err != nil {
		return MessageInfo, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(data) == 0 {
		return MessageInfo, fmt.Errorf("empty payload")
	}

	// Get first controller ID
	var controllerID string
	var deviceID string
	for k := range data {
		controllerID = k
		break
	}

	logger.Debug("Processing controller", zap.String("controllerID", controllerID))

	ignoredControllers, err := workers.GetIgnoredControllers()
	if err != nil {
		return MessageInfo, fmt.Errorf("error getting ignored controllers: %w", err)
	}

	if slices.Contains(ignoredControllers, controllerID) {
		return MessageInfo, fmt.Errorf("controller is ignored: %s", controllerID)
	}

	deviceID = controllerID

	logger.Debug("Processing device", zap.String("deviceID", deviceID))

	ignoredDevices, err := workers.GetIgnoredDevices()
	if err != nil {
		return MessageInfo, fmt.Errorf("error getting ignored devices: %w", err)
	}

	if slices.Contains(ignoredDevices, deviceID) {
		return MessageInfo, fmt.Errorf("device is ignored: %s", deviceID)
	}

	device, err := workers.GetDevicesByDeviceIdentifier(deviceID)
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return MessageInfo, fmt.Errorf("device not found: %s", deviceID)
		}

		return MessageInfo, fmt.Errorf("error getting device by device ID - %s: %w", deviceID, err)
	}

	deviceType := device.DeviceType
	deviceTypeLower := strings.ToLower(deviceType)
	timestamp := msg.MessageTimestamp

	var devices []types.Device
	var rawData map[string]any
	var processedData map[string]any

	logger.Debug(fmt.Sprintf("%s :: %s", device.Controller, device.DeviceType))

	switch deviceTypeLower {
	// Process Genset devices
	case DeviceTypeGenset:
		rawData, processedData, err = genset.Decoder(data[controllerID])
		if err != nil {
			return MessageInfo, fmt.Errorf("error decoding genset data: %w", err)
		}
	}
	rawData["SerialNo1"] = device.ControllerIdentifier
	processedData["SerialNo1"] = device.ControllerIdentifier

	deviceStruct := &types.Device{
		CustomerID:           device.Site.Customer.ID,
		CustomerName:         device.Site.Customer.Name,
		SiteID:               device.Site.ID,
		SiteName:             device.Site.Name,
		Controller:           device.Controller,
		DeviceType:           device.DeviceType,
		ControllerIdentifier: device.ControllerIdentifier,
		DeviceName:           device.DeviceName,
		DeviceIdentifier:     device.DeviceIdentifier,
		RawData:              rawData,
		ProcessedData:        processedData,
		Timestamp:            timestamp,
	}

	devices = append(devices, *deviceStruct)

	return &types.MessageInfo{
		MessageID: msg.ID.String(),
		Devices:   devices,
	}, nil
}
