// Package common contains shared data available for all plugins.
package common

// Color defines color parameter type.
type Color struct {
	R uint8 `json:"r" validate:"required"`
	G uint8 `json:"g" validate:"required"`
	B uint8 `json:"b" validate:"required"`
}

// Int defines simple integer parameter type.
type Int struct {
	Value int `json:"value" validate:"required"`
}

// String defines simple string parameter type.
type String struct {
	Value string `json:"value" validate:"required"`
}

// Percent defines percent parameter type.
type Percent struct {
	Value uint8 `json:"value" validate:"required,percent"`
}
