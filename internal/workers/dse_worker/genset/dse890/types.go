package dse890_decoder

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
