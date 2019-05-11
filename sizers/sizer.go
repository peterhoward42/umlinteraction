package sizers

import (
	"github.com/peterhoward42/umli/dslmodel"
)

// Sizer is the top level sizing component.
type Sizer struct {
	TopMargin float64
	Lanes     *Lanes
}

// NewSizer provides a Sizer structure that has been initialised
// as is ready for use.
func NewSizer(diagramWidth int, fontHeight float64,
	statements []*dslmodel.Statement) *Sizer {
	sizer := &Sizer{}
	sizer.TopMargin = diagramTopMarginK * fontHeight
	sizer.Lanes = NewLanes(diagramWidth, fontHeight, statements)
	return sizer
}
