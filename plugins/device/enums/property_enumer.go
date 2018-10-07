// Code generated by "enumer -type=Property -transform=snake -trimprefix=Prop -json -text -yaml"; DO NOT EDIT.

package enums

import (
	"encoding/json"
	"fmt"
)

const _PropertyName = "oncolornum_devicestransition_timebrightnessscenespowertemperaturebattery_levelsunrisesunsethumiditypressurevisibilitywind_directionwind_speedclickdouble_clickpresssensor_typevac_statusareadurationfan_speedpicturedistanceuser"

var _PropertyIndex = [...]uint8{0, 2, 7, 18, 33, 43, 49, 54, 65, 78, 85, 91, 99, 107, 117, 131, 141, 146, 158, 163, 174, 184, 188, 196, 205, 212, 220, 224}

func (i Property) String() string {
	if i < 0 || i >= Property(len(_PropertyIndex)-1) {
		return fmt.Sprintf("Property(%d)", i)
	}
	return _PropertyName[_PropertyIndex[i]:_PropertyIndex[i+1]]
}

var _PropertyValues = []Property{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26}

var _PropertyNameToValueMap = map[string]Property{
	_PropertyName[0:2]:     0,
	_PropertyName[2:7]:     1,
	_PropertyName[7:18]:    2,
	_PropertyName[18:33]:   3,
	_PropertyName[33:43]:   4,
	_PropertyName[43:49]:   5,
	_PropertyName[49:54]:   6,
	_PropertyName[54:65]:   7,
	_PropertyName[65:78]:   8,
	_PropertyName[78:85]:   9,
	_PropertyName[85:91]:   10,
	_PropertyName[91:99]:   11,
	_PropertyName[99:107]:  12,
	_PropertyName[107:117]: 13,
	_PropertyName[117:131]: 14,
	_PropertyName[131:141]: 15,
	_PropertyName[141:146]: 16,
	_PropertyName[146:158]: 17,
	_PropertyName[158:163]: 18,
	_PropertyName[163:174]: 19,
	_PropertyName[174:184]: 20,
	_PropertyName[184:188]: 21,
	_PropertyName[188:196]: 22,
	_PropertyName[196:205]: 23,
	_PropertyName[205:212]: 24,
	_PropertyName[212:220]: 25,
	_PropertyName[220:224]: 26,
}

// PropertyString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func PropertyString(s string) (Property, error) {
	if val, ok := _PropertyNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Property values", s)
}

// PropertyValues returns all values of the enum
func PropertyValues() []Property {
	return _PropertyValues
}

// IsAProperty returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Property) IsAProperty() bool {
	for _, v := range _PropertyValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for Property
func (i Property) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Property
func (i *Property) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Property should be a string, got %s", data)
	}

	var err error
	*i, err = PropertyString(s)
	return err
}

// MarshalText implements the encoding.TextMarshaler interface for Property
func (i Property) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface for Property
func (i *Property) UnmarshalText(text []byte) error {
	var err error
	*i, err = PropertyString(string(text))
	return err
}

// MarshalYAML implements a YAML Marshaler for Property
func (i Property) MarshalYAML() (interface{}, error) {
	return i.String(), nil
}

// UnmarshalYAML implements a YAML Unmarshaler for Property
func (i *Property) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	var err error
	*i, err = PropertyString(s)
	return err
}
