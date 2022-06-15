//go:build st3310
// +build st3310

package device

import (
	"time"

	"github.com/dumacp/go-ingnovus/st3310"
)

func NewDevice(port string, speed int) (Device, error) {
	dev := st3310.NewDevice(port, speed, 600*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
