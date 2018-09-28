package device

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"strings"

	"github.com/corona10/goimagehash"
	"github.com/disintegration/imaging"
	"github.com/go-home-io/server/plugins/device/enums"
	"gopkg.in/yaml.v2"
)

const (
	// Default distance to use.
	defaultCameraDistance = 15
	// Default JPEG quality to use.
	defaultImageQuality = 50
	// Desired width of the final image.
	defaultCameraWidth = 800
)

// Device settings.
type cameraSettings struct {
	Distance int `yaml:"distance"`
	Quality  int `yaml:"quality"`
	Width    int `yaml:"width"`
}

// Post-processor for a camera device.
type cameraProcessor struct {
	distance int
	quality  int
	width    int
	prevHash *goimagehash.ImageHash
}

// Constructs a new camera processor.
func newCameraProcessor(rawConfig string) IProcessor {
	s := &cameraSettings{}
	err := yaml.Unmarshal([]byte(rawConfig), s)
	if err != nil {
		return &cameraProcessor{distance: defaultCameraDistance, quality: defaultImageQuality, width: defaultCameraWidth}
	}

	if s.Distance < 1 {
		s.Distance = defaultCameraDistance
	}

	if s.Quality > 100 || s.Quality < 0 {
		s.Quality = defaultImageQuality
	}

	if s.Width < 0 || s.Width > 2000 {
		s.Width = defaultCameraWidth
	}

	return &cameraProcessor{distance: s.Distance, quality: s.Quality, width: s.Width}
}

// IsPropertyGood determines whether property has to be processed by worker and sent back to master.
func (p *cameraProcessor) IsPropertyGood(prop enums.Property, val interface{}) (bool, interface{}) {
	if prop != enums.PropPicture {
		return true, val
	}

	reader := strings.NewReader(val.(string))
	img, err := jpeg.Decode(reader)
	if err != nil {
		return false, nil
	}

	hash, err := goimagehash.AverageHash(img)
	if err != nil {
		return false, nil
	}

	if nil == p.prevHash {
		p.prevHash = hash
		return p.resizeImage(img)
	}

	distance, err := p.prevHash.Distance(hash)
	p.prevHash = hash
	if err != nil {
		return false, nil
	}

	if distance >= p.distance {
		return p.resizeImage(img)
	}

	return false, nil
}

// Performs image resizing.
func (p *cameraProcessor) resizeImage(original image.Image) (bool, interface{}) {
	dst := imaging.Resize(original, p.width, 0, imaging.Lanczos)
	buf := bytes.NewBuffer(make([]byte, 0))
	err := jpeg.Encode(buf, dst, &jpeg.Options{Quality: p.quality})
	if err != nil {
		return false, nil
	}

	return true, base64.StdEncoding.EncodeToString(buf.Bytes())
}
