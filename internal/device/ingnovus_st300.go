package device

import (
	"time"

	"github.com/dumacp/go-ingnovus/st300"
)

func NewDeviceSt300(port string, speed int) (Device, error) {
	dev := st300.NewDevice(port, speed, 600*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
