//go:build ingnovus
// +build ingnovus

package device

import (
	"github.com/dumacp/go-ingnovus"
	"time"
)

func NewDevice(port string, speed int) (interface{}, error) {
	dev := ingnovus.NewDevice(port, speed, 1000*time.Millisecond)

	if err := dev.Open(); err != nil {
		return nil, err
	}

	return dev, nil
}
