// +build ingnovus

package device

import (
	"github.com/dumacp/go-ingnovus"
	"time"
)

func NewDevice(port string, speed int) (interface{}, error) {
	dev, err := ingnovus.NewDevice(port, speed, 600*time.Millisecond)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
