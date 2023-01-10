package device

import (
	"time"

	"github.com/dumacp/go-ingnovus"
)

func NewDeviceIngnovus(port string, speed int) (Device, error) {
	dev := ingnovus.NewDevice(port, speed, 1000*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
