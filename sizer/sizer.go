/*
Package sizers is the single source of truth for sizing all the elements
that make up the diagram. Not only how big things are, but also how far
apart they should be.

E.g. the coordinates for each lifeline title box,
the mark-space settings for dashed lines, arrow sizing, the margins or
padding required for each thing etc.

It is encapsulated in this dedicated package, to remove this responsibility
from the umli.diag package, so that umli.diag can deal only with the
algorithmic part of diagram creation.

It provides the top level *Sizer* type, along with some subordinate types
it delegates to. For example: sizing.Lifelines.
*/
package sizers

import (
	"github.com/peterhoward42/umli/dslmodel"
)

// Naming conventions:
// Consider the name: <XxxPadT>
// This should be read as the padding required by
// thing Xxxx at the top (T). Where T is from {LRTB}

// Sizer is the top level sizing component.
type Sizer struct {
	DiagramPadT                float64
	DiagramPadB                float64
	DiagPadL                   float64
	FrameTitleTextPadT         float64
	FrameTitleTextPadB         float64
	FrameTitleTextPadL         float64
	FrameTitleBoxWidth         float64
	FrameTitleRectPadB         float64
	Lifelines                  *Lifelines // Delegated sizing
	InteractionLinePadB        float64
	InteractionLineTextPadB    float64
	InteractionLineLabelIndent float64
	ArrowLen                   float64
	ArrowHeight                float64
	DashLineDashLen            float64
	DashLineDashGap            float64
	SelfLoopHeight             float64
	ActivityBoxVerticalOverlap float64
	FinalizedActivityBoxesPadB float64
	MinLifelineSegLength       float64
}

// NewSizer provides a Sizer structure that has been initialised
// as is ready for use.
func NewSizer(diagramWidth int, fontHeight float64,
	lifelineStatements []*dslmodel.Statement) *Sizer {
	sizer := &Sizer{}

	// The requested font height is used as a datum reference,
	// and nearly everything is sized in proportion to this, using
	// settings from settings.go.

	// These settings values typically end with the letter
	// K (e.g. diagramPadTK) to indicate they are proportion coefficients.

	fh := fontHeight

	sizer.DiagramPadT = diagramPadTK * fh
	sizer.DiagramPadB = diagramPadBK * fh
	sizer.DiagPadL = diagPadLK * fh

	sizer.FrameTitleTextPadT = frameTitleTextPadTK * fh
	sizer.FrameTitleTextPadT = frameTitleTextPadBK * fh
	sizer.FrameTitleBoxWidth = frameTitleBoxWidthK * float64(diagramWidth)
	sizer.FrameTitleRectPadB = frameTitleRectPadBK * fh
	sizer.FrameTitleTextPadL = frameTitleTextPadLK * fh

	sizer.Lifelines = NewLifelines(diagramWidth, fh, lifelineStatements)

	sizer.ArrowLen = arrowLenK * fh
	sizer.ArrowHeight = arrowAspectRatio * sizer.ArrowLen
	sizer.InteractionLinePadB = interactionLinePadBK * fh
	sizer.InteractionLineTextPadB = interactionLineTextPadBK * fh
	sizer.InteractionLineLabelIndent = sizer.ArrowLen +
		interactionLineLabelIndent*fh
	sizer.DashLineDashLen = dashLineDashLenK * fh
	sizer.DashLineDashGap = dashLineDashGapK * fh
	sizer.SelfLoopHeight = sizer.Lifelines.SelfLoopWidth * selfLoopAspectRatio
	sizer.ActivityBoxVerticalOverlap = activityBoxVerticalOverlapK * fh
	sizer.FinalizedActivityBoxesPadB = finalizedActivityBoxesPadB * fh
	sizer.MinLifelineSegLength = minLifelineSegLengthK * fh

	return sizer
}
