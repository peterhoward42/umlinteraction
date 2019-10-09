package lifeline

import (
	"fmt"
	"errors"

	"github.com/peterhoward42/umli/dsl"
	"github.com/peterhoward42/umli/sizer"
)

/*
Spacing holds the knowledge about the horizontal pitch and geometry
of lifelines. For example, how to space them out across the page, and
providing the left and right edge coordinates for the activity boxes on each
etc.

It makes all the title boxes the same width, and distributes these equally
across the width of the diagram. It uses the same gap (gutter) between these
boxes and as margins at the left and right edge of the diagram.
*/
type Spacing struct {
	sizer     sizer.Sizer
	lifelines []*dsl.Statement
}

// NewSpacing  provides a Spacing  ready to use.
func NewSpacing(sizer sizer.Sizer, lifelines []*dsl.Statement) *Spacing {
	return &Spacing{
		sizer:     sizer,
		lifelines: lifelines,
	}
}

/*
CentreLine provides the X coordinate for the centreline of lifeline.
*/
func (s Spacing) CentreLine(lifeline *dsl.Statement) (float64, error) {
	dv := s.calcDrivingValues()
	_ = dv
	num, err := s.lifelineNumber(lifeline)
	_ = num
	if err != nil {
		return 0, fmt.Errorf("lifelineNumber: %v", err)
	}
	/*
		return (n+1)*drivingValues.TitleBoxGutter + (n+0.5)*drivingValues.TitleBoxWidth
	*/
	return 0.0, nil
}

/*
The spacing algorithm is dependent on a few key driving values.
For example the chosen width of the lifeline title boxes, and the gap between
them. This function decides what they should be.
*/
type drivingValues struct {
	titleBoxWidth  float64
	titleBoxGutter float64
}

func (s *Spacing) calcDrivingValues() drivingValues {
	var dv drivingValues
	dv.titleBoxWidth = s.sizer.Get("IdealLifelineTitleBoxWidth")
	n := len(s.lifelines)
	diagWidth := s.sizer.Get("DiagWidth")
	spaceAvail := diagWidth - dv.titleBoxWidth*float64(n)
	nGuttersRequired := n + 1
	dv.titleBoxGutter = spaceAvail / float64(nGuttersRequired)

	// But if that has that made the gutter too small, or even negative,
	// make the boxes less wide to preserve a minimum gutter equal to
	// one font height.
	fontHt := s.sizer.Get("FontHt")
	if dv.titleBoxGutter < fontHt {
		dv.titleBoxGutter = fontHt
		dv.titleBoxWidth = diagWidth - float64(n+1)*dv.titleBoxGutter/float64(n)
	}
	return dv
}

func (s *Spacing) lifelineNumber(lifeline *dsl.Statement) (int, error) {
	for num, registered := range s.lifelines {
		if registered == lifeline {
			return num, nil
		}
	}
	return -1, errors.New("lifeline is not registered")
}