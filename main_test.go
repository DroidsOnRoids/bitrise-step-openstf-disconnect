package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateConfigNoHostUrl(t *testing.T) {
	configs := configsModel{stfAccessToken: "test"}
	require.Error(t, configs.validate())
}

func TestValidateConfigNoAccessToken(t *testing.T) {
	configs := configsModel{stfHostURL: "http://test.test"}
	require.Error(t, configs.validate())
}

func TestValidateConfigNoErrors(t *testing.T) {
	configs := configsModel{stfHostURL: "http://test.test", stfAccessToken: "test"}
	require.NoError(t, configs.validate())
}

func TestParseJSONStringArraySafelyEmptyString(t *testing.T) {
	jsonArray, err := parseJSONStringArraySafely("")
	require.NoError(t, err)
	require.Equal(t, []string{}, jsonArray)
}

func TestParseJSONStringArraySafelyNonEmptyValidString(t *testing.T) {
	jsonArray, err := parseJSONStringArraySafely(`["test", "test2"]`)
	require.NoError(t, err)
	require.Equal(t, []string{"test", "test2"}, jsonArray)
}

func TestParseJSONStringArraySafelyNonEmptyInvalidString(t *testing.T) {
	jsonArray, err := parseJSONStringArraySafely("test")
	require.Error(t, err)
	require.Nil(t, jsonArray)
}

func TestExtractNonEmptyDevicesListFromAdbOutput(t *testing.T) {
	devices := extractDevicesListFromAdbOutput(`List of devices attached
123456789               device
1.2.3.4:123456          device
`)
	require.Len(t, devices, 2)
	require.Equal(t, "123456789", devices[0])
	require.Equal(t, "1.2.3.4:123456", devices[1])
}

func TestExtractEmptyDevicesListFromAdbOutput(t *testing.T) {
	devices := extractDevicesListFromAdbOutput(`List of devices attached`)
	require.Empty(t, devices)
}
