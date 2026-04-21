package helpers

import "encoding/json"

// JsonToMap used to convert json object into map[string]
func JsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}
