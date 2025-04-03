package genset

import (
	"fmt"
	"math"
	"reflect"

	coreutils "github.com/johandrevandeventer/dse-worker/utils"
)

type P004 struct {
	R000 float64 `json:"R000"` // OilPressure
	R001 float64 `json:"R001"` // CoolantTemp
	R002 float64 `json:"R002"` // OilTemp
	R003 float64 `json:"R003"` // Fuel
	R004 float64 `json:"R004"` // AlternatorV
	R005 float64 `json:"R005"` // BatV
	R006 float64 `json:"R006"` // Rpm
	R007 float64 `json:"R007"` // GenFreq
	R008 float64 `json:"R008"` // GenL1
	R010 float64 `json:"R010"` // GenL2
	R012 float64 `json:"R012"` // GenL3
}

type P006 struct {
	R000 float64 `json:"R000"` // GenTotalP
	R008 float64 `json:"R008"` // GenTotalS
	R022 float64 `json:"R022"` // LoadPercentage
	R114 float64 `json:"R114"` // AvgVoltage
	R130 float64 `json:"R130"` // AvgCurrent
}

type P007 struct {
	R002 float64 `json:"R002"` // NextService
	R006 float64 `json:"R006"` // RunTime
	R008 float64 `json:"R008"` // GenkWh
	R012 float64 `json:"R012"` // GenkVAh
	R016 float64 `json:"R016"` // TotalStart
	R034 float64 `json:"R034"` // FuelUsed
}

type P003 struct {
	R004 float64 `json:"R004"` // Mode1
}

type P166 struct {
	R000 float64 `json:"R000"` // ComAlarm
	R002 float64 `json:"R002"` // FailStart
	R004 float64 `json:"R004"` // MainFail
	R006 float64 `json:"R006"` // Maintanance
	R008 float64 `json:"R008"` // Estop
	R010 float64 `json:"R010"` // AutoMode
}

type P005 struct {
	R117 float64 `json:"R117"` // FuelTrip
}

// Main struct combining all the smaller structs
type DSE890Data struct {
	P004 P004 `json:"P004"`
	P006 P006 `json:"P006"`
	P007 P007 `json:"P007"`
	P003 P003 `json:"P003"`
	P166 P166 `json:"P166"`
	P005 P005 `json:"P005"`
}

var ratioMap = map[string]float64{
	"P004.R000": 1.0,         // OilPressure
	"P004.R001": 1.0,         // CoolantTemp
	"P004.R002": 1.0,         // OilTemp
	"P004.R003": 1.0,         // Fuel
	"P004.R004": 0.1,         // AlternatorV
	"P004.R005": 0.1,         // BatV
	"P004.R006": 1.0,         // Rpm
	"P004.R007": 0.1,         // GenFreq
	"P004.R008": 0.1,         // GenL1
	"P004.R010": 0.1,         // GenL2
	"P004.R012": 0.1,         // GenL3
	"P006.R000": 0.001,       // GenTotalP
	"P006.R008": 1.0,         // GenTotalS
	"P006.R022": 0.1,         // LoadPercentage
	"P006.R114": 0.1,         // AvgVoltage
	"P006.R130": 0.1,         // AvgCurrent
	"P007.R002": 0.000277778, // NextService
	"P007.R006": 0.000277778, // RunTime
	"P007.R008": 0.1,         // GenkWh
	"P007.R012": 0.1,         // GenkVAh
	"P007.R016": 1.0,         // TotalStart
	"P007.R034": 0.1,         // FuelUsed
	"P003.R004": 1.0,         // Mode1
	"P166.R000": 1.0,         // ComAlarm
	"P166.R002": 1.0,         // FailStart
	"P166.R004": 1.0,         // MainFail
	"P166.R006": 1.0,         // Maintanance
	"P166.R008": 1.0,         // Estop
	"P166.R010": 1.0,         // AutoMode
	"P005.R117": 1.0,         // FuelTrip
}

func Decoder(payload map[string]map[string]any) (rawData, processedData map[string]any, err error) {
	var dse890Data DSE890Data

	// Decode map into struct
	err = coreutils.DecodeMapToStruct(payload, &dse890Data)
	if err != nil {
		return rawData, processedData, fmt.Errorf("error decoding DSE890 data: %w", err)
	}

	rawData = createDataMap(dse890Data)

	resetOutOfRangeValues(&dse890Data)
	applyRatios(&dse890Data)

	processedData = createDataMap(dse890Data)

	return rawData, processedData, nil
}

func createDataMap(pm DSE890Data) map[string]any {
	dataMap := map[string]any{
		"Oil_Pressure":   pm.P004.R000,
		"CoolantTemp":    pm.P004.R001,
		"OilTemp":        pm.P004.R002,
		"Fuel":           pm.P004.R003,
		"AlternatorV":    pm.P004.R004,
		"BatV":           pm.P004.R005,
		"Rpm":            pm.P004.R006,
		"Gen_Freq":       pm.P004.R007,
		"Gen_L1":         pm.P004.R008,
		"Gen_L2":         pm.P004.R010,
		"Gen_L3":         pm.P004.R012,
		"GenTotalP":      pm.P006.R000,
		"GenTotalS":      pm.P006.R008,
		"Loadpercentage": pm.P006.R022,
		"Avg_Voltage":    pm.P006.R114,
		"Avg_Current":    pm.P006.R130,
		"Next_Service":   pm.P007.R002,
		"RunTime":        pm.P007.R006,
		"GenkWh":         pm.P007.R008,
		"GenkVAh":        pm.P007.R012,
		"TotalStart":     pm.P007.R016,
		"Fuel_Used":      pm.P007.R034,
		"Mode1":          pm.P003.R004,
		"Estop":          pm.P166.R008,
		"MainFail":       pm.P166.R002,
		"ComAlarm":       pm.P166.R000,
		"FailStart":      pm.P166.R002,
		"Maintanance":    pm.P166.R004,
		"AutoMode":       pm.P166.R010,
		"FuelTrip":       pm.P005.R117,
	}

	return dataMap
}

func resetOutOfRangeValues(pm *DSE890Data) {
	outOfRangeValues := []float64{32763, 2147483643, 2147483644, 2147483645, 2147483646, 2147483647, 2147483648}

	pmValue := reflect.ValueOf(pm).Elem()

	for i := 0; i < pmValue.NumField(); i++ {
		point := pmValue.Field(i)

		for j := 0; j < point.NumField(); j++ {
			fieldValue := point.Field(j).Float() // Assuming the field type is int

			// Check if the field value is out of range
			for _, outOfRange := range outOfRangeValues {
				if fieldValue == outOfRange {
					point.Field(j).SetFloat(0) // Set to 0 if it's out of range
					break                      // No need to check other out of range values
				}
			}
		}
	}
}

func applyRatios(pm *DSE890Data) {
	// Get the reflection value of the PointMap
	pmValue := reflect.ValueOf(pm).Elem()

	// Iterate over the fields in PointMap (P004, P006, etc.)
	for i := 0; i < pmValue.NumField(); i++ {
		point := pmValue.Type().Field(i).Name
		pointValue := pmValue.Field(i)

		// Iterate over the fields within each struct (R000, R001, etc.)
		for j := 0; j < pointValue.NumField(); j++ {
			field := pointValue.Type().Field(j).Name
			fieldValue := pointValue.Field(j).Float() // Get the integer value of the field

			// Create the key for the ratioMap lookup (e.g., "P004.R000")
			ratioKey := fmt.Sprintf("%s.%s", point, field)

			// Check if there is a ratio for the key
			if ratio, ok := ratioMap[ratioKey]; ok {
				// Apply the ratio and store it as float64
				newValue := math.Round(float64(fieldValue)*ratio*100) / 100
				pointValue.Field(j).Set(reflect.ValueOf(newValue)) // Set the modified float64 value
			}
		}
	}
}
