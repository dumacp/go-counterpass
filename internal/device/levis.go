//go:build (beane || !ingnovus) && (beane || !optocontrol) && (beane || !sonar) && (beane || !extreme) && (beane || !st300) && (beane || !logirastreo) && (beane || !st3310)
// +build beane !ingnovus
// +build beane !optocontrol
// +build beane !sonar
// +build beane !extreme
// +build beane !st300
// +build beane !logirastreo
// +build beane !st3310

package device

import "github.com/dumacp/go-levis"

func NewDevice(port string, speed int) (Device, error) {
	dev, err := levis.NewDevice(port, speed)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
