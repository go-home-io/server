// Package common contains shared data available for all plugins.
package common

import "image/color"

// Color defines color parameter type.
type Color struct {
	R uint8 `json:"r" yaml:"r"`
	G uint8 `json:"g" yaml:"g"`
	B uint8 `json:"b" yaml:"b"`
}

// NewColor creates a color from a system color
func NewColor(color color.Color) Color {
	r, g, b, _ := color.RGBA()
	return Color{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
	}
}

// Color returns system color
func (c *Color) Color() color.Color {
	return color.RGBA{
		R: c.R,
		G: c.G,
		B: c.B,
		A: 0,
	}
}

// Int defines simple integer parameter type.
type Int struct {
	Value int `json:"value" validate:"required"`
}

// Float defines simple float parameter type.
type Float struct {
	Value float64 `json:"value" validate:"required"`
}

// String defines simple string parameter type.
type String struct {
	Value string `json:"value" validate:"required"`
}

// Percent defines percent parameter type.
type Percent struct {
	Value uint8 `json:"value" validate:"required,percent"`
}
