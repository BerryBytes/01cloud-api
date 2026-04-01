package models

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type Manifest struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Spec       map[string]interface{} `json:"spec"`
	Metadata   map[string]interface{} `json:"metadata"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

func ToJson(data interface{}) string {
	jsonString, _ := json.Marshal(data)
	return string(jsonString)
}

func ToMap(data interface{}) map[string]interface{} {
	jsonString, _ := json.Marshal(data)
	mp := map[string]interface{}{}
	err := json.Unmarshal(jsonString, &mp)
	if err != nil {
		log.Error(err)
		return nil
	}
	return mp
}

func FromJson(input []byte) map[string]interface{} {
	mp := map[string]interface{}{}
	err := json.Unmarshal(input, &mp)
	if err != nil {
		log.Error(err)
		return nil
	}
	return mp
}

func ValidateUserVariable(userVariables []map[string]string) bool {
	mySet := map[string]bool{}
	for _, variable := range userVariables {
		if _, ok := mySet[variable["key"]]; ok {
			return false
		}
		mySet[variable["key"]] = true
	}
	return true
}

func GetValuesJson(inputMap map[string]interface{}, outputMap map[string]interface{}, key string) interface{} {
	if inputMap["type"].(string) == "string" {
		if _, ok := inputMap["hidden"]; ok {
			if _, ok := inputMap["hidden"].(bool); ok {
				outputMap[key] = uuid.New().String()
			} else {
				outputMap[key] = inputMap["value"]
			}
		} else {
			outputMap[key] = inputMap["value"]
		}
	} else if inputMap["type"].(string) == "object" {
		newMap := map[string]interface{}{}
		for k, v := range inputMap["properties"].(map[string]interface{}) {
			GetValuesJson(v.(map[string]interface{}), newMap, k)
		}
		log.Debug("new value json map :: ", newMap)
		if key != "" {
			outputMap[key] = newMap
		}
	}
	return outputMap
}
