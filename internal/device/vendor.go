package device

import (
	"flag"
	"fmt"
)

var VendorArray = []string{"beane", "sonar", "extreme", "optocontrol", "optocontrol-v301",
	"ingnovus-st300", "ingnovus-st3310", "ingnovus-st3310H04", "ingnovus"}
var VendorCounter string
var TimeoutSamples int

func init() {
	flag.StringVar(&VendorCounter, "vendor", "",
		fmt.Sprintf("vendor type of counterpass device\n\toptions: %q (default %q)", VendorArray, VendorArray[0]))
	flag.IntVar(&TimeoutSamples, "timeout", 900, "timeout samples in millis")
}

func VerifyVendor(vendor string) error {
	// defer func() {
	// 	fmt.Printf("vendor: %s\n", VendorCounter)
	// }()
	verifyTimeoutSamples(vendor)
	if len(vendor) <= 0 {
		VendorCounter = VendorArray[0]
		return nil
	}
	for _, v := range VendorArray {
		if v == vendor {
			VendorCounter = vendor
			return nil
		}
	}
	return fmt.Errorf("error, choose a supported vendor: %q", VendorArray)
}

func verifyTimeoutSamples(vendor string) {
	if TimeoutSamples > 0 {
		return
	}
	switch vendor {
	case "beane", "optocontrol", "optocontrol-v301":
		TimeoutSamples = 900
	case "extreme", "logirastreo", "ingnovus-st300":
		TimeoutSamples = 3000
	case "ingnovus-st3310", "ingnovus-st3310H04":
		TimeoutSamples = 300
	default:
		TimeoutSamples = 900
	}
}
