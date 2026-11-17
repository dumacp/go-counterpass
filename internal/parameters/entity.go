package parameters

type Configuration struct {
	Port                 string `json:"port"`
	Vendor               string `json:"vendor"`
	Baud                 int    `json:"baud"`
	Type                 int    `json:"type"`
	VerifyOpenDoor       bool   `json:"verifyOpenDoor"`
	IsZeroOpenStateDoor0 bool   `json:"isZeroOpenStateDoor0"`
	IsZeroOpenStateDoor1 bool   `json:"isZeroOpenStateDoor1"`
}
