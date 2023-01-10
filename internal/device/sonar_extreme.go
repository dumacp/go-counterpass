package device

import (
	"time"

	"github.com/dumacp/sonar/ins50"
)

func NewDeviceExtreme(port string, speed int) (Device, error) {
	dev := ins50.NewDevice(port, speed, 1*time.Second)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
