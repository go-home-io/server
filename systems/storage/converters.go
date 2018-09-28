package storage

import (
	"encoding/json"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device/enums"
)

// PropertySave converts actual property before storing into the database.
func PropertySave(property enums.Property, value interface{}) (interface{}, error) {
	switch property {
	case enums.PropScenes, enums.PropSensorType:
		return nil, nil
	case enums.PropOn, enums.PropClick, enums.PropDoubleClick, enums.PropPress, enums.PropPicture:
		return value, nil
	case enums.PropBrightness, enums.PropBatteryLevel, enums.PropFanSpeed:
		return value.(common.Percent).Value, nil
	case enums.PropDuration:
		return value.(common.Int).Value, nil
	case enums.PropColor:
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}

		return string(data), nil
	case enums.PropVacStatus:
		return value, nil

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
			return nil, err
		}

		return data, nil
	}

	return value, nil
}
