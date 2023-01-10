package device

import (
	"time"

	"github.com/dumacp/go-ingnovus/st3310H04"
)

func NewDeviceSt3310_H04(port string, speed int) (Device, error) {
	dev := st3310H04.NewDevice(port, speed, 600*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
