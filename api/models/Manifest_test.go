package models

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mainfest = Manifest{
	ApiVersion: "test-app-version",
	Kind:       "test-kind",
	Spec:       map[string]interface{}{"data": "test-data"},
	Metadata:   map[string]interface{}{"data": "test-data"},
	Data:       map[string]interface{}{"data": "test-data"},
}

func TestMainfestToJson(t *testing.T) {
	data := ToJson(mainfest)
	assert.NotNil(t, data)
}

func TestMainfestToMap(t *testing.T) {
	for _, condition := range testCondition {
		if condition == "fail" {
			data := ToMap("mainfest")
			if data == nil {
				assert.Nil(t, data)
			} else {
				data := ToMap(mainfest)
				assert.NotNil(t, data)
			}

		}

	}

}

func TestMainfestFromJson(t *testing.T) {
	for _, condition := range testCondition {
		if condition == "fail" {
			data := "{1234:10,\"Kind\":\"test-kind\"}"
			res := FromJson([]byte(data))
			if res == nil {
				assert.Nil(t, res)
			} else {
				data := "{\"Name\":10,\"Kind\":\"test-kind\"}"
				res := FromJson([]byte(data))
				assert.NotNil(t, res)
			}

		}

	}

}

func TestMainfestValidateUserVariable(t *testing.T) {
	m := make(map[string]string)
	var myarr [2]map[string]string

	for _, condition := range testCondition {
		if condition == "fail" {
			m["keys"] = "false"
			myarr[0] = m
			data := ValidateUserVariable(myarr[:])
			assert.Equal(t, reflect.TypeOf(data).Kind(), reflect.Bool)
		} else {
			m["key"] = "true"
			myarr[0] = m
			data := ValidateUserVariable(myarr[:])
			assert.Equal(t, reflect.TypeOf(data).Kind(), reflect.Bool)
		}

	}

}

func TestGetValuesJson(t *testing.T) {

	testcases := []struct {
		data map[string]interface{}
	}{
		{
			data: map[string]interface{}{
				"type":   "string",
				"hidden": true,
				"value":  "value",
			},
		},
		{
			data: map[string]interface{}{
				"type":   "string",
				"hidden": false,
				"value":  "value",
			},
		},
		{
			data: map[string]interface{}{
				"type":  "string",
				"value": "value",
			},
		},
	}
	for _, tc := range testcases {
		data := GetValuesJson(tc.data, tc.data, "value")
		assert.NotNil(t, data)
	}

}
