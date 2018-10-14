package storage

import (
	"encoding/json"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/helpers"
	"github.com/pkg/errors"
)

// PropertySave converts actual property before storing into the database.
func PropertySave(property enums.Property, value interface{}) (interface{}, error) {
	// Something we don't care to store
	if property == enums.PropScenes || property == enums.PropSensorType {
		return nil, nil
	}

	switch helpers.GetPropertyType(property) {
	case helpers.PropEnum, helpers.PropBool, helpers.PropString, helpers.PropStringSlice:
		return value, nil
	case helpers.PropPercent:
		return value.(common.Percent).Value, nil
	case helpers.PropInt:
		return value.(common.Int).Value, nil
	case helpers.PropColor:
		data, err := json.Marshal(value)
		if err != nil {
			return nil, errors.Wrap(err, "json marshal failed")
		}

		return string(data), nil
	default:
		return value.(common.Float).Value, nil
	}
}

// PropertyLoad restores actual property from the database.
func PropertyLoad(property enums.Property, value interface{}) (interface{}, error) {
	switch property {
	case enums.PropColor:
		data := common.Color{}
		err := json.Unmarshal([]byte(value.(string)), &data)
		if err != nil {
			return nil, errors.Wrap(err, "json un-marshal failed")
		}

		return data, nil
	}

	return value, nil
}
