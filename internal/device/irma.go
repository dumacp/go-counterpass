package device

type IrmaDevice struct {
	Url string
}

func (dev *IrmaDevice) Close() error {
	return nil
}

func NewIrma(url string) (*IrmaDevice, error) {
	return &IrmaDevice{Url: url}, nil
}
