package device

import "github.com/dumacp/go-levis"

func NewDeviceBaene(port string, speed int) (Device, error) {

	dev, err := levis.NewDevice(port, speed)
	if err != nil {
		return nil, err
	}

	return dev, nil
}
