// Package common contains shared data available for all plugins.
package common

// Color defines color parameter type.
type Color struct {
	R uint8 `json:"r" validate:"required" yaml:"r"`
	G uint8 `json:"g" validate:"required" yaml:"g"`
	B uint8 `json:"b" validate:"required" yaml:"b"`
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
