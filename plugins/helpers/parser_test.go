package helpers

import (
	"encoding/json"
	"testing"

	"github.com/go-home-io/server/plugins/common"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests evaluation of different expressions.
func TestEvaluations(t *testing.T) {
	data := []struct {
		payload        string
		expression     string
		expectedResult interface{}
		property       enums.Property
	}{
		{
			expression:     "payload=='on'",
			payload:        "on",
			expectedResult: true,
			property:       enums.PropOn,
		},
		{
			expression:     "jq(payload, '.status') == 'off'",
			payload:        "{ \"status\": \"off\" }",
			expectedResult: true,
			property:       enums.PropOn,
		},
		{
			expression:     "jq(payload, '.status.value') == 'true'",
			payload:        "{ \"status\": { \"value\": true } }",
			expectedResult: true,
			property:       enums.PropOn,
		},
		{
			expression:     "num(payload)",
			payload:        "10",
			expectedResult: 10.0,
			property:       enums.PropBrightness,
		},
		{
			expression:     "num(payload) == 10.0",
			payload:        "10.0",
			expectedResult: true,
			property:       enums.PropOn,
		},
	}

	p := NewParser()

	for _, v := range data {
		exp, err := p.Compile(v.expression)
		require.NoError(t, err, "compile %s", v.expression)
		res, err := exp.Parse(v.payload)
		require.NoError(t, err, "parse %s", v.expression)
		assert.True(t, PropertyDeepEqual(res, v.expectedResult, v.property), "equal %s", v.expression)
	}
}

// Tests errors in evaluating.
func TestParseErrors(t *testing.T) {
	p := NewParser()

	_, err := p.Compile("_=1")
	assert.Error(t, err, "compile workerd")

	data := []string{
		"num()",
		"num(1,2)",
		"num('int')",
		"str()",
		"str(1,3)",
		"fmt()",
		"fmt()",
		"jq()",
		"jq(1)",
		"jq('ok', 1)",
		"jq('ok', '=<)')",
		"jq('ok', '.data')",
		"jq('ok', '1', '1')",
	}

	for _, v := range data {
		exp, err := p.Compile(v)
		require.NoError(t, err, "compile %s", v)
		_, err = exp.Parse("test")
		assert.Error(t, err, "parse %s", v)
	}
}

// Tests JSON conversion.
func TestJsonConvert(t *testing.T) {
	p := NewParser()

	tp := struct {
		T string
	}{T: "data"}
	d, _ := json.Marshal(tp)
	exp, _ := p.Compile("jq(payload)")
	val, _ := exp.Parse(string(d))
	assert.Equal(t, "data", val.(map[string]interface{})["T"].(string))
}

// Tests correct parsing.
func TestParse(t *testing.T) {
	data := []struct {
		expression     string
		expectedResult interface{}
		params         map[string]interface{}
		property       enums.Property
	}{
		{
			expression:     "str('on')",
			params:         nil,
			property:       123,
			expectedResult: "on",
		},
		{
			expression:     "str(1)",
			params:         nil,
			property:       123,
			expectedResult: "1",
		},
		{
			expression: "(state.On == true) ? 'on' : 'off'",
			params: map[string]interface{}{"state": &device.LightState{
				On: true,
			}},
			property:       111,
			expectedResult: "on",
		},
		{
			expression:     "'generic'",
			params:         map[string]interface{}{},
			property:       111,
			expectedResult: "generic",
		},
		{
			expression: "fmt('r:%v,g:%v,b:%v', state.Color.R, state.Color.G, state.Color.B)",
			params: map[string]interface{}{"state": &device.LightState{
				On: true,
				Color: common.Color{
					R: 10,
					G: 20,
					B: 30,
				},
			}},
			property:       111,
			expectedResult: "r:10,g:20,b:30",
		},
		{
			expression: "fmt(state.Color.R)",
			params: map[string]interface{}{"state": &device.LightState{
				On: true,
				Color: common.Color{
					R: 10,
					G: 20,
					B: 30,
				},
			}},
			property:       111,
			expectedResult: "10",
		},
	}
	p := NewParser()

	for _, v := range data {
		exp, err := p.Compile(v.expression)
		require.NoError(t, err, "compile %s", v.expression)
		res, err := exp.Format(v.params)
		require.NoError(t, err, "parse %s", v.expression)

		if !v.property.IsAProperty() {
			assert.Equal(t, v.expectedResult, res, "not property %s", v.expression)
			continue
		}

		assert.True(t, PropertyDeepEqual(res, v.expectedResult, v.property), v.expression)
	}
}
