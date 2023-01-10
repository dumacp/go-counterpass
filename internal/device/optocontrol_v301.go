package device

import (
	"time"

	"github.com/dumacp/go-optocontrol"
)

func NewDeviceOptocontrol_v301(port string, speed int) (Device, error) {
	dev := optocontrol.NewV301(port, speed, 1*time.Second)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
