package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type configsModel struct {
	stfHostURL     string
	stfAccessToken string
	deviceSerials  []string
}

const userDevicesEndpoint = "/api/v1/user/devices"

var adbDeviceLineRegex = regexp.MustCompile(`([\w.:]+)\s+device$`)
var client = &http.Client{Timeout: time.Second * 30}

func main() {
	configs, err := createConfigsModelFromEnvs()
	if err != nil {
		log.Errorf("Could not create config, error: %s", err)
		os.Exit(1)
	}
	configs.dump()
	if err := configs.validate(); err != nil {
		log.Errorf("Could not validate config, error: %s", err)
		os.Exit(1)
	}

	adbDevicesList := getAdbDevicesList()

	adbDevices := extractDevicesListFromAdbOutput(adbDevicesList)
	connectedDevices := mapSerialsToAdbDevices(adbDevices)

	for _, serialToDisconnect := range configs.deviceSerials {
		log.Printf("Releasing device %s", serialToDisconnect)

		device := connectedDevices[serialToDisconnect]
		if err := disconnectDevice(device); err != nil {
			log.Printf("Could not disconnect device from ADB: %s", err)
		}

		if err := removeDeviceFromControl(configs, serialToDisconnect); err != nil {
			log.Printf("Could not remove device from control, error: %s", err)
		}
	}
}

func createConfigsModelFromEnvs() (configsModel, error) {
	serials, err := parseJSONStringArraySafely(os.Getenv("stf_device_serial_list"))
	if err != nil {
		return configsModel{}, err
	}
	return configsModel{
		stfHostURL:     os.Getenv("stf_host_url"),
		stfAccessToken: os.Getenv("stf_access_token"),
		deviceSerials:  serials,
	}, nil
}

func parseJSONStringArraySafely(raw string) ([]string, error) {
	var array []string
	if raw == "" {
		return []string{}, nil
	}
	if err := json.Unmarshal([]byte(raw), &array); err != nil {
		return nil, fmt.Errorf("input %s cannot be deserialized, error %s", raw, err)
	}
	return array, nil
}

func (configs configsModel) dump() {
	log.Infof("Config:")
	log.Infof("STF host: %s", configs.stfHostURL)
	log.Infof("Device serials: %s", configs.deviceSerials)
}

func (configs *configsModel) validate() error {
	if !strings.HasPrefix(configs.stfHostURL, "http") {
		return fmt.Errorf("invalid STF host: %s", configs.stfHostURL)
	}
	if configs.stfAccessToken == "" {
		return errors.New("STF access token cannot be empty")
	}
	return nil
}

func removeDeviceFromControl(configs configsModel, serial string) error {
	req, err := http.NewRequest("DELETE", configs.stfHostURL+userDevicesEndpoint+"/"+serial, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+configs.stfAccessToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if err := response.Body.Close(); err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("request failed, status: %s", response.Status)
	}
	return nil
}

func disconnectDevice(device string) error {
	output, err := command.RunCommandAndReturnCombinedStdoutAndStderr("adb", "disconnect", device)
	if err != nil {
		return fmt.Errorf("%s, error: %s", output, err)
	}
	return nil
}

func getAdbDevicesList() string {
	output, err := command.RunCommandAndReturnStdout("adb", "devices")
	if err != nil {
		log.Warnf("Could not get ADB devices, error: %s", err)
	}
	return output
}

func extractDevicesListFromAdbOutput(adbDevicesOutput string) []string {
	lines := strings.Split(adbDevicesOutput, "\n")

	var devices []string
	for _, line := range lines {
		if submatch := adbDeviceLineRegex.FindStringSubmatch(line); submatch != nil {
			devices = append(devices, submatch[1])
		}
	}
	return devices
}

func mapSerialsToAdbDevices(adbDevices []string) map[string]string {
	serialsToDevicesMap := make(map[string]string)
	for _, device := range adbDevices {
		serial, err := command.RunCommandAndReturnStdout("adb", "-s", device, "shell", "getprop ro.serialno")
		if err != nil {
			log.Warnf("Could not get serial for device %s, error: %s", device, err)
		} else {
			serialsToDevicesMap[serial] = device
		}
	}
	return serialsToDevicesMap
}
