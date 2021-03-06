package lifeline

import (
	"testing"

	"github.com/peterhoward42/umli/diag/nogozone"
	"github.com/peterhoward42/umli/dsl"
	"github.com/peterhoward42/umli/geom"
	"github.com/peterhoward42/umli/graphics"
	"github.com/peterhoward42/umli/parser"
	"github.com/peterhoward42/umli/sizer"
	"github.com/stretchr/testify/assert"
)

func TestLifelinesGetDrawnCorrectlyIncludingMakingTheRequiredGaps(t *testing.T) {
	assert := assert.New(t)

	// Our test case will have 3 lifelines from left to right A,B,C,
	// There will be one NoGoZone and one activity box interrupting
	// lifeline B.

	dslScript := `
		life A foo
		life B bar
		life C baz
	`
	dslModel := parser.MustCompileParse(dslScript)
	width := 2000.0
	fontHt := 10.0
	sizer := sizer.NewLiteralSizer(map[string]float64{
		"FrameInternalPadB":          10,
		"IdealLifelineTitleBoxWidth": 300.0,
	})
	lifelines := dslModel.LifelineStatements()
	spacer := NewSpacing(sizer, fontHt, width, lifelines)
	noGoSeg := geom.NewSegment(50, 60)
	zone := nogozone.NewNoGoZone(noGoSeg, lifelines[0], lifelines[2])
	noGoZones := []nogozone.NoGoZone{zone}
	boxes := map[*dsl.Statement]*BoxTracker{}
	for _, ll := range lifelines {
		boxes[ll] = NewBoxTracker()
	}
	boxesForLifeB := boxes[lifelines[1]]
	err := boxesForLifeB.AddStartingAt(80)
	assert.NoError(err)
	err = boxesForLifeB.TerminateAt(90)
	assert.NoError(err)

	minSegLen := 1.0
	lifelineF := NewFinalizer(lifelines, spacer, noGoZones, boxes, sizer)
	top := 10.0
	bottom := 100.0
	primitives := graphics.NewPrimitives()
	err = lifelineF.Finalize(top, bottom, minSegLen, primitives)
	assert.NoError(err)

	// The lines created for lifelines A and C (only) should run from
	// top to bottom uninterrupted.
	for _, i := range []int{0, 2} {
		lifeCoords, err := spacer.CentreLine(lifelines[i])
		assert.NoError(err)
		expectedX := lifeCoords.Centre
		expectedLine := graphics.Line{
			P1:     graphics.NewPoint(expectedX, top),
			P2:     graphics.NewPoint(expectedX, bottom),
			Dashed: true,
		}
		assert.True(primitives.ContainsLine(expectedLine))
	}

	// The lines created for lifeline B should have gaps in.
	lifeCoords, err := spacer.CentreLine(lifelines[1])
	assert.NoError(err)
	expectedX := lifeCoords.Centre
	expectedSegments := []geom.Segment{
		geom.NewSegment(10, 50),
		geom.NewSegment(60, 80),
		geom.NewSegment(90, 100),
	}
	for _, seg := range expectedSegments {
		expectedLine := graphics.Line{
			P1:     graphics.NewPoint(expectedX, seg.Start),
			P2:     graphics.NewPoint(expectedX, seg.End),
			Dashed: true,
		}
		assert.True(primitives.ContainsLine(expectedLine))
	}
}
