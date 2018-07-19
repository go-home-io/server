package helpers

import (
	"testing"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/device"
	"github.com/go-home-io/server/plugins/common"
	"encoding/json"
)

type testExpData struct {
	payload        string
	expression     string
	expectedResult interface{}
	property       enums.Property
}

type testFormatData struct {
	expression     string
	expectedResult interface{}
	params         map[string]interface{}
	property       enums.Property
}

// Tests evaluation of different expressions.
func TestEvaluations(t *testing.T) {
	data := []testExpData{
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
		if err != nil {
			t.Error("failed to compile " + v.expression)
			t.FailNow()
		}

		res, err := exp.Parse(v.payload)
		if err != nil {
			t.Error("failed to parse " + v.expression)
			t.FailNow()
		}
		if !PropertyDeepEqual(res, v.expectedResult, v.property) {
			t.Error("unexpected result " + v.expression)
			t.Fail()
		}
	}
}

// Tests errors in evaluating.
func TestParseErrors(t *testing.T) {
	p := NewParser()

	_, err := p.Compile("_=1")
	if err == nil {
		t.Fail()
	}

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
		if err != nil {
			t.Error("failed to compile " + v)
			t.FailNow()
		}

		_, err = exp.Parse("test")
		if err == nil {
			t.Error("unexpected result " + v)
		}
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

	if val.(map[string]interface{})["T"].(string) != "data" {
		t.Fail()
	}
}

// Tests correct parsing.
func TestParse(t *testing.T) {
	data := []testFormatData{
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
		if err != nil {
			t.Error("failed to compile " + v.expression)
			t.FailNow()
		}

		res, err := exp.Format(v.params)
		if err != nil {
			t.Error("failed to parse " + v.expression)
			t.FailNow()
		}

		if !v.property.IsAProperty() {
			if v.expectedResult != res {
				t.Error("unexpected result " + v.expression)
				t.FailNow()
			}

			continue
		}

		if !PropertyDeepEqual(res, v.expectedResult, v.property) {
			t.Error("unexpected result " + v.expression)
			t.Fail()
		}
	}
}
