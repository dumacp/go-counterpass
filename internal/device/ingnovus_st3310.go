package device

import (
	"time"

	"github.com/dumacp/go-ingnovus/st3310"
)

func NewDeviceSt3310(port string, speed int) (Device, error) {
	dev := st3310.NewDevice(port, speed, 600*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
