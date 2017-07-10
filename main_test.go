package main

import (
	"testing"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigNoHostUrl(t *testing.T) {
	configs := configsModel{stfAccessToken:"test"}
	require.Error(t, configs.validate())
}

func TestValidateConfigNoAccessToken(t *testing.T) {
	configs := configsModel{stfHostURL:"http://test.test"}
	require.Error(t, configs.validate())
}

func TestValidateConfigNoErrors(t *testing.T) {
	configs := configsModel{stfHostURL:"http://test.test", stfAccessToken:"test"}
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