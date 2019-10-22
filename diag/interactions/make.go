package interactions

import (
	"fmt"

	"github.com/peterhoward42/umli"
	"github.com/peterhoward42/umli/diag/lifeline"
	"github.com/peterhoward42/umli/dsl"
	"github.com/peterhoward42/umli/graphics"
)

/*
Maker knows how to make the interaction lines and when to start/stop
their activity boxes.
*/
type Maker struct {
	dependencies  *MakerDependencies
	graphicsModel *graphics.Model
}

/*
MakerDependencies encapsulates the prior state of the diagram creation
process at the time that the Make method is called. And includes all
the things the Maker needs from the outside to do its job.
*/
type MakerDependencies struct {
	fontHt float64
	spacer *lifeline.Spacing
}

// NewMakerDependencies makes a MakerDependencies ready to use.
func NewMakerDependencies(fontHt float64, spacer *lifeline.Spacing) *MakerDependencies {
	return &MakerDependencies{
		fontHt: fontHt,
		spacer: spacer,
	}
}

/*
NewMaker initialises a Maker ready to use.
*/
func NewMaker(d *MakerDependencies, gm *graphics.Model) *Maker {
	return &Maker{
		dependencies:  d,
		graphicsModel: gm,
	}
}

/*
Scan goes through the DSL statements in order, and works out what graphics are
required to represent interaction lines, and activitiy boxes etc. It advances
the tidemark as it goes, and returns the final resultant tidemark.
*/
func (mkr *Maker) Scan(
	tidemark float64,
	statements []*dsl.Statement) (newTidemark float64, err error) {

	// Build a list of actions to execute depending on the statement
	// keyword.
	actions := []dispatch{}
	for _, s := range statements {
		switch s.Keyword {
		case umli.Dash:
			actions = append(actions, dispatch{mkr.interactionLabel, s})
			actions = append(actions, dispatch{mkr.startToBox, s})
			actions = append(actions, dispatch{mkr.interactionLine, s})
		case umli.Full:
			actions = append(actions, dispatch{mkr.interactionLabel, s})
			actions = append(actions, dispatch{mkr.startFromBox, s})
			actions = append(actions, dispatch{mkr.startToBox, s})
			actions = append(actions, dispatch{mkr.interactionLine, s})
		case umli.Self:
			actions = append(actions, dispatch{mkr.startFromBox, s})
			actions = append(actions, dispatch{mkr.interactionLine, s})
		case umli.Stop:
			actions = append(actions, dispatch{mkr.endBox, s})
		}
	}
	var prevTidemark float64 = tidemark
	var updatedTidemark float64
	for _, action := range actions {
		updatedTidemark, err := action.fn(prevTidemark, action.statement)
		if err != nil {
			return -1, fmt.Errorf("actionFn: %v", err)
		}
		prevTidemark = updatedTidemark
	}
	return updatedTidemark, nil
}

// interactionLabel creates the graphics label that belongs to an interaction
// line.
func (mkr *Maker) interactionLabel(
	tidemark float64, s *dsl.Statement) (newTidemark float64, err error) {
	fromX, toX, err := mkr.LifelineCentres(s)
	x, horizJustification := NewLabelPosn(fromX, toX).Get()
	mkr.graphicsModel.Primitives.RowOfStrings(
		x, tidemark, mkr.dependencies.fontHt, horizJustification, s.LabelSegments)
	/*

		firstRowY := tideMark
		c.rowOfLabels(x, firstRowY, horizJustification, s.LabelSegments)
		tideMark += float64(len(s.LabelSegments))*
			mkr.dependencies.fontHt + mkr.dependencies.sizer.Get("InteractionLineTextPadB")
		c.ilZones.RegisterSpaceClaim(
			sourceLifeline, destLifeline, firstRowY, c.tideMark)
	*/
	return -1, nil
}

// startToBox registers with <something> that an activity box on a lifeline
// should be started ready for an interaction line to arrive at the top of
// it.
func (mkr *Maker) startToBox(
	tidemark float64, s *dsl.Statement) (newTidemark float64, err error) {
	return -1, nil
}

// starFromBox registers with <something> that an activity box on a lifeline
// should be started for an activity line to emenate from it.
func (mkr *Maker) startFromBox(
	tidemark float64, s *dsl.Statement) (newTidemark float64, err error) {
	return -1, nil
}

// interactionLine makes an interaction line (and its arror)
func (mkr *Maker) interactionLine(
	tidemark float64, s *dsl.Statement) (newTidemark float64, err error) {
	return -1, nil
}

// endBox registers with <something> that an activity box that has been
// started, should now be terminated.
func (mkr *Maker) endBox(
	tidemark float64, s *dsl.Statement) (newTidemark float64, err error) {
	return -1, nil
}

// actionFn describes a function that can be called to draw something
// related to statement s. Such a function receives the current tidemark,
// calculates how it should be advanced, and return the updated value.
type actionFn func(
	tideMark float64,
	s *dsl.Statement) (newTidemark float64, err error)

// dispatch is a simple container to hold a binding between  an actionFn and
// the statement to which it refers.
type dispatch struct {
	fn        actionFn
	statement *dsl.Statement
}

/*
LifelineCentres evaluates the X coordinates for the lifelines between which
an interaction line travels.
*/
func (mkr *Maker) LifelineCentres(
	interactionLine *dsl.Statement) (fromX, toX float64, err error) {
	fromCoords, err := mkr.dependencies.spacer.CentreLine(
		interactionLine.ReferencedLifelines[0])
	if err != nil {
		return -1.0, -1.0, fmt.Errorf("space.CentreLine: %v", err)
	}
	toCoords, err := mkr.dependencies.spacer.CentreLine(
		interactionLine.ReferencedLifelines[1])
	if err != nil {
		return -1.0, -1.0, fmt.Errorf("space.CentreLine: %v", err)
	}
	return fromCoords.Centre, toCoords.Centre, nil
}