package helpers

import (
	"testing"
	"github.com/go-home-io/server/plugins/device/enums"
	"github.com/go-home-io/server/plugins/common"
)

type testData struct {
	input interface{}
	gold  interface{}
	prop  enums.Property
	cmd   enums.Command
}

// Tests properties conversion.
func TestProperties(t *testing.T) {
	data := []testData{
		{
			input: map[interface{}]interface{}{"r": 10, "g": 20, "b": 30},
			gold:  &common.Color{R: 10, G: 20, B: 30,},
			prop:  enums.PropColor,
			cmd:   enums.CmdSetColor,
		},
		{
			input: true,
			gold:  true,
			prop:  enums.PropOn,
			cmd:   -1,
		},
		{
			input: 10,
			gold:  &common.Percent{Value: 10},
			prop:  enums.PropBrightness,
			cmd:   enums.CmdSetBrightness,
		},
	}

	for _, v := range data {
		p, err := PropertyFixYaml(v.input, v.prop)
		if err != nil {
			t.Error("failed fix property " + v.prop.String())
		}

		if !PropertyDeepEqual(p, v.gold, v.prop) {
			t.Error("failed deep equal property " + v.prop.String())
		}

		if -1 == v.cmd {
			continue
		}

		p, err = CommandPropertyFixYaml(v.input, v.cmd)

		if err != nil {
			t.Error("failed fix command " + v.cmd.String())
		}

		if !PropertyDeepEqual(p, v.gold, v.prop) {
			t.Error("failed deep equal command " + v.prop.String())
		}
	}
}
