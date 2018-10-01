package device

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/corona10/goimagehash"
	"github.com/go-home-io/server/plugins/device/enums"
)

func getCamera() IProcessor {
	return newDeviceProcessor(enums.DevCamera, `
width: 760
quality: 50`)
}

func getImage() (string, *goimagehash.ImageHash) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, G: 255, B: 255})
	buf := bytes.NewBuffer(make([]byte, 0))
	hash, _ := goimagehash.AverageHash(img)
	jpeg.Encode(buf, img, &jpeg.Options{Quality: 100})

	return string(buf.Bytes()), hash
}

// Tests camera processor constructor.
func TestCameraProcessorCtor(t *testing.T) {
	c := getCamera()

	if c.(*cameraProcessor).width != 760 || c.(*cameraProcessor).quality != 50 ||
		c.(*cameraProcessor).distance != defaultCameraDistance {
		t.Fail()
	}
}

// Tests regular property validation.
func TestCameraProcessorRegularProperty(t *testing.T) {
	c := getCamera()
	ok, out := c.IsPropertyGood(enums.PropArea, 20)
	if !ok || len(out) != 1 || out[enums.PropArea] != 20 {
		t.Fail()
	}
}

// Tests passing incorrect data.
func TestCorruptedPicture(t *testing.T) {
	c := getCamera()
	ok, out := c.IsPropertyGood(enums.PropPicture, "wrong data")
	if ok || out != nil {
		t.Fail()
	}
}

// Tests passing a new image.
func TestNewImage(t *testing.T) {
	c := getCamera()
	img, _ := getImage()

	ok, out := c.IsPropertyGood(enums.PropPicture, img)

	if !ok || out == nil {
		t.Error("Validation failed")
		t.FailNow()
	}

	outB, err := base64.StdEncoding.DecodeString(out[enums.PropPicture].(string))
	if err != nil {
		t.Error("Received wrong base64 data")
		t.FailNow()
	}

	reader := bytes.NewReader(outB)
	jp, err := jpeg.Decode(reader)
	if err != nil {
		t.Error("Didn't receive image: " + err.Error())
		t.FailNow()
	}

	if jp.Bounds().Max.X != 760 || jp.Bounds().Max.Y != 760 {
		t.Error("Received image is incorrect")
		t.Fail()
	}
}

// Tests passing same image twice.
func TestSameImage(t *testing.T) {
	c := getCamera()
	img, _ := getImage()
	ok, _ := c.IsPropertyGood(enums.PropPicture, img)
	if !ok {
		t.Error("Validation failed")
		t.FailNow()
	}

	ok, _ = c.IsPropertyGood(enums.PropPicture, img)
	if ok {
		t.Error("Second validation failed")
		t.Fail()
	}
}

// Tests whether all extra properties correctly processed.
func TestExtraProperties(t *testing.T){
	c := getCamera()
	for _, v := range c.GetExtraSupportPropertiesSpec() {
		if !c.IsExtraProperty(v) {
			t.Error("Failed extra property " + v.String())
			t.Fail()
		}
	}
}
