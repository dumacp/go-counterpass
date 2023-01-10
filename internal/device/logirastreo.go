package device

import (
	"time"

	logirastreo "github.com/dumacp/go-logirastreo/v1"
)

func NewDeviceLogirastreo(port string, speed int) (Device, error) {
	dev := logirastreo.NewDevice(port, speed, 3*time.Second)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
