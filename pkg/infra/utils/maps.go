package utils

import "reflect"

// AreMapsHaveTheSameKeys checks if two maps have the same keys - check also nested maps.
func AreMapsHaveTheSameKeys(map1 map[string]interface{}, map2 map[string]interface{}) bool {
	return AreAllKeysOfFirstMapExistsInSecondMap(map1, map2) && AreAllKeysOfFirstMapExistsInSecondMap(map2, map1)
}

// AreAllKeysOfFirstMapExistsInSecondMap checks if all the keys of the first map are exist in the second map - check also nested maps.
func AreAllKeysOfFirstMapExistsInSecondMap(map1 map[string]interface{}, map2 map[string]interface{}) bool {
	if map1 == nil || map2 == nil {
		return false
	}

	// Iterate over all items of the first map
	for k, v := range map1 {
		// First check that k exists in map2 and extract its value.
		v2, ok := map2[k]
		if !ok {
			return false
		}

		// Second check if there is nested map.
		typ := reflect.TypeOf(v).Kind()
		if typ == reflect.Map {
			typ2 := reflect.TypeOf(v2).Kind()
			if typ2 != reflect.Map {
				return false
			}

			vAsMap, ok := v.(map[string]interface{})
			if !ok {
				return false
			}

			v2AsMap, ok := v2.(map[string]interface{})
			if !ok {
				return false
			}

			// Recursive call.
			if !AreAllKeysOfFirstMapExistsInSecondMap(vAsMap, v2AsMap) {
				return false
			}
		}
	}
	return true
}
