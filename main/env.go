package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dumacp/go-counterpass/internal/device"
)

const baud_defaut = 19200

func getENV() {
	if len(socket) <= 0 {
		port_env := os.Getenv("COUNTERPASS_PORT")
		if len(port_env) <= 0 {
			socket = "/dev/ttyS2"
		} else {
			socket = os.Getenv("COUNTERPASS_PORT")
		}
	}
	if baudRate <= 0 {
		baud_env := os.Getenv("COUNTERPASS_BAUD")
		if v, err := strconv.Atoi(baud_env); err == nil && v > 0 {
			baudRate = v
		} else {
			baudRate = baud_defaut
		}
	}

	vendor_env := device.VendorCounter
	if len(vendor_env) <= 0 {
		vendor_env = os.Getenv("COUNTERPASS_VENDOR")
	}
	if err := device.VerifyVendor(vendor_env); err != nil {
		log.Fatalln(err)
	}

	if typeCounterDoor <= 0 && !strings.Contains(fmt.Sprintf("%s", os.Args), "typeCounterDoor") {
		temp := os.Getenv("COUNTERPASS_DOORTYPE")
		// fmt.Printf("COUNTERPASS_DOORTYPE: %s\n", temp)
		if v, err := strconv.Atoi(temp); err == nil && v > 0 {
			typeCounterDoor = v
		} else {
			typeCounterDoor = 0
		}
	}
	if err := device.VerifyVendor(vendor_env); err != nil {
		log.Fatalln(err)
	}
}
