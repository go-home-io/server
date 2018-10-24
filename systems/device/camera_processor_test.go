package device

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/corona10/goimagehash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-home.io/x/server/plugins/device/enums"
)

func getCamera() IProcessor {
	return newDeviceProcessor(enums.DevCamera, `
width: 760
quality: 50`)
}

//noinspection GoUnhandledErrorResult
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
	assert.Equal(t, 760, c.(*cameraProcessor).width, "width")
	assert.Equal(t, 50, c.(*cameraProcessor).quality, "quality")
	assert.Equal(t, defaultCameraDistance, c.(*cameraProcessor).distance, "distance")
}

// Tests regular property validation.
func TestCameraProcessorRegularProperty(t *testing.T) {
	c := getCamera()
	ok, out := c.IsPropertyGood(enums.PropArea, 20)
	assert.True(t, ok, "not ok")
	require.Equal(t, 1, len(out), "length")
	assert.Equal(t, 20, out[enums.PropArea], "area")
}

// Tests passing incorrect data.
func TestCorruptedPicture(t *testing.T) {
	c := getCamera()
	ok, out := c.IsPropertyGood(enums.PropPicture, "wrong data")
	assert.False(t, ok)
	assert.Nil(t, out)
}

// Tests passing a new image.
func TestNewImage(t *testing.T) {
	c := getCamera()
	img, _ := getImage()

	ok, out := c.IsPropertyGood(enums.PropPicture, img)
	require.True(t, ok)
	require.NotNil(t, out)

	outB, err := base64.StdEncoding.DecodeString(out[enums.PropPicture].(string))
	require.NoError(t, err, "not a base64")

	reader := bytes.NewReader(outB)
	jp, err := jpeg.Decode(reader)
	require.NoError(t, err, "not an image")
	assert.Equal(t, 760, jp.Bounds().Max.X, "x")
	assert.Equal(t, 760, jp.Bounds().Max.Y, "y")
}

// Tests passing same image twice.
func TestSameImage(t *testing.T) {
	c := getCamera()
	img, _ := getImage()
	ok, _ := c.IsPropertyGood(enums.PropPicture, img)
	assert.True(t, ok, "first validation")

	ok, _ = c.IsPropertyGood(enums.PropPicture, img)
	assert.False(t, ok, "second validation")
}

// Tests whether all extra properties correctly processed.
func TestExtraProperties(t *testing.T) {
	c := getCamera()
	for _, v := range c.GetExtraSupportPropertiesSpec() {
		assert.True(t, c.IsExtraProperty(v), v.String())
	}
}
