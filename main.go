package main

import (
	"fmt"
	"net/http"
	"time"
	"log"
	"strings"
	"errors"
	"os"
	"encoding/json"
)

type configsModel struct {
	stfHostURL     string
	stfAccessToken string
	deviceSerials  []string
}

const userDevicesEndpoint = "/api/v1/user/devices"

var client = &http.Client{Timeout: time.Second * 10}

func main() {
	configs, err := createConfigsModelFromEnvs()
	if err != nil {
		log.Fatalf("Could not create config, error: %s", err)
	}
	configs.dump()
	if err := configs.validate(); err != nil {
		log.Fatalf("Could not validate config, error: %s", err)
	}
	for _, serial := range configs.deviceSerials {
		log.Printf("Releasing device %s", serial)
		if err := removeDeviceFromControl(configs, serial); err != nil {
			log.Fatalf("Could not remove device from control, error: %s", err)
		}
	}
}

func createConfigsModelFromEnvs() (configsModel, error) {
	serials, err := parseJSONStringArraySafely(os.Getenv("stf_device_serial_list"))
	if err != nil {
		return configsModel{}, err
	}
	return configsModel{
		stfHostURL:        os.Getenv("stf_host_url"),
		stfAccessToken:    os.Getenv("stf_access_token"),
		deviceSerials:     serials,
	}, nil
}

func parseJSONStringArraySafely(raw string) ([]string, error) {
	var array []string
	if raw == "" {
		return []string{}, nil
	}
	if err := json.Unmarshal([]byte(raw), &array); err != nil {
		return nil, fmt.Errorf("Input %s cannot be deserialized, error %s", raw, err)
	}
	return array, nil
}

func (configs configsModel) dump() {
	log.Println("Config:")
	log.Printf("STF host: %s", configs.stfHostURL)
	log.Printf("Device serials: %s", configs.deviceSerials)
}

func (configs *configsModel) validate() error {
	if !strings.HasPrefix(configs.stfHostURL, "http") {
		return fmt.Errorf("Invalid STF host: %s", configs.stfHostURL)
	}
	if configs.stfAccessToken == "" {
		return errors.New("STF access token cannot be empty")
	}
	return nil
}

func removeDeviceFromControl(configs configsModel, serial string) error {
	req, err := http.NewRequest("DELETE", configs.stfHostURL + userDevicesEndpoint + "/" + serial, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer " + configs.stfAccessToken)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	if err := response.Body.Close(); err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("Request failed, status: %s", response.Status)
	}
	return nil
}
