package diag

/*
This module contains the *Creator* type which provides the public interface
to clients of the *diag* package.

The module then provides the high-level implementation for Create() and
expresses the essential creation algorithm - delegating much of its work to
code in other modules in the package.

See todo for an explanation of the diagram creation algorithm.
*/

import (
	"github.com/peterhoward42/umli"
	"github.com/peterhoward42/umli/dsl"
	"github.com/peterhoward42/umli/graphics"
	"github.com/peterhoward42/umli/sizer"
)

/*
Creator is the top level type for the diag package and provides the API
to create diagrams.
*/
type Creator struct {
	/*
	   *width* is the width of the entire diagram. This is a, private and
	   arbitrary, working-width that serves only to provide us with a fixed,
	   private, abstract coordinate system to build the model in. It is expected
	   that model renderers will need / want to scale it to a coordinate system
	   that suits them at render time.
	*/
	width float64
	// Font height is used as the root for all sizing decisions.
	fontHeight float64
	// Parsed DSL script.
	model *dsl.Model
	// Owns the horizontal sizing and layout for lifelines
	lifelineGeomH *lifelineGeomH
	// In charge of making the outer frame and title.
	frameMaker *frameMaker
	// Keeps track of activity box top and bottom coordinates.
	activityBoxes map[*dsl.Statement]*lifelineBoxes
	// Keeps track of the space taken up by interaction lines.
	ilZones *InteractionLineZones
	// A delegate to make the lifelines dashed line segments.
	lifelineMaker *lifelineMaker
	// The output.
	graphicsModel *graphics.Model
	// Knows how to size everything.
	sizer *sizer.Sizer
	// Gradually moves down the page during creation.
	tideMark float64
}

/*
Create is the main API method which work out what the diagram should look like.
It orchestrates a multi-pass creation process which accumulates the graphics
primitives required in its graphicsModel and then returns that model.
*/
func (c *Creator) Create(model *dsl.Model) *graphics.Model {
	c.initializeTheCreator(model)
	c.initializeTheGraphicsModel()
	c.createGraphicsAnchoredToTopOfDiagram()
	c.processDSLWorkingDownThePage()
	c.finalizeActivityBoxes()
	c.finalizeLifelines()
	c.frameMaker.finalizeFrame()
	c.finalizeDiagramHeight()
	return c.graphicsModel
}

/*
initializeTheCreator initialises the Creator structure, incuding composing
a set of helper objects to which it can delegate.
*/
func (c *Creator) initializeTheCreator(m *dsl.Model) {
	c.model = m
	c.setWidthAndFontHeight()
	c.addLettersToLifelineTitles()
	c.activityBoxes = map[*dsl.Statement]*lifelineBoxes{}
	for _, s := range c.model.LifelineStatements() {
		c.activityBoxes[s] = newlifelineBoxes()
	}
	c.sizer = sizer.NewSizer(c.width, c.fontHeight)
	c.lifelineGeomH = newLifelineGeomH(c.width, c.fontHeight, c.sizer,
		c.model)
	c.frameMaker = newFrameMaker(c)
	c.ilZones = NewInteractionLineZones(c)
	c.lifelineMaker = newLifelineMaker(c)
}

/*
setWidthAndFontHeight sets the modelled diagram width and sets the font height
as as a ratio of the width. The ratio is taken from the DSL when present, or
otherwise a default.
*/
func (c *Creator) setWidthAndFontHeight() {
	c.width = 2000.0 // Arbitrary but human-relatable to support debugging.
	const defaultTextHeightRatio = 1.0 / 100.0
	textHeightRatio := defaultTextHeightRatio
	s, ok := c.model.FirstStatementOfType(umli.TextSize)
	if ok {
		if s.TextSize == 0 {
			panic("Developer error, this must not be allowed to happen")
		}
		// 5  signifies 1:200  0.005
		// 10 signifies 1:100  0.010
		// 20 signifies 1:50   0.020
		textHeightRatio = s.TextSize / 1000.0
	}
	c.fontHeight = c.width * textHeightRatio
}

/*
addLettersToLifelineTitles looks for a *showletters* statement, and sets the creator's
flag attribute accordingly, or to the behaviour default of true.
*/
func (c *Creator) addLettersToLifelineTitles() {
	showLetters := true // default behaviour
	s, ok := c.model.FirstStatementOfType(umli.ShowLetters)
	if ok && s.ShowLetters == false {
		showLetters = false
	}
	// To show the lifeline letters we add them to the lifeline
	// title box labels.
	if showLetters {
		for _, s := range c.model.LifelineStatements() {
			letter := s.LifelineName
			addition := []string{"", letter}
			s.LabelSegments = append(s.LabelSegments, addition...)
		}
	}
}

/*
initializeTheGraphicsModel constructs a graphics.Model parameterized by
width, height and font height and attaches it to the creator.
*/
func (c *Creator) initializeTheGraphicsModel() {
	diagHeight := 0.0 // Set later to accomodate contents once known.
	c.graphicsModel = graphics.NewModel(
		c.width, diagHeight, c.fontHeight,
		c.sizer.DashLineDashLen, c.sizer.DashLineDashGap)
}

/*
createGraphicsAnchoredToTopOfDiagram generates the graphics that must be
produced at the top of the diagram E.g the frame and title box, and the
lifelines with their title boxes at the top of each.
*/
func (c *Creator) createGraphicsAnchoredToTopOfDiagram() {
	c.tideMark = c.sizer.DiagramPadT
	// Quite complex - so delegate.
	c.frameMaker.initFrameAndMakeTitleBox()
	c.lifelineTitleBoxes()
}

/*
processDSLWorkingDownThePage takes each parsed statement from the DSL script in
turn, to generate the sequence-dependent primitives.  This includes for example
the interaction lines and labels. But it excludes the generation of primitives
that can only be dimensioned once the interaction line Y coordinates are known;
for example the activity boxes that sit on lifelines.
*/
func (c *Creator) processDSLWorkingDownThePage() {
	graphicalEvents := newScanner().Scan(c.model.Statements())
	// Outer loop is per DSL statement
	for _, statement := range c.model.Statements() {
		statementEvents := graphicalEvents[statement]
		// Inner loop is for the (multiple) graphical events called for
		// by that statement.
		for _, evt := range statementEvents {
			// Evaluate and add the graphics primitives required.
			c.graphicsForDrawingEvent(evt, statement)
		}
	}
}

/*
finalizeDiagramHeight sets the graphics model's Height attribute to just
large enough to accomodate the final tide mark.
*/
func (c *Creator) finalizeDiagramHeight() {
	c.graphicsModel.Height = c.tideMark + c.sizer.DiagramPadB
}

/*
graphicsForDrawingEvent synthesizes the lines and label strings primititives
required to render a single diagram element drawing event. It also advances
c.tideMark, to accomodate the space taken up by what it generated.
*/
func (c *Creator) graphicsForDrawingEvent(evt eventType,
	statement *dsl.Statement) {

	switch evt {
	case EndBox:
		c.endBox(statement)
	case InteractionLine:
		c.interactionLine(statement)
	case InteractionLabel:
		c.interactionLabel(statement)
	case SelfInteractionLines:
		c.selfInteractionLines(statement)
	case PotentiallyStartFromBox:
		c.potentiallyStartFromBox(statement)
	case PotentiallyStartToBox:
		c.potentiallyStartToBox(statement)
	}
}

/*
lifelineTitleBoxes generates the lines and text to draw the title boxes at
the top of all the lifelines. Then advances the tide mark corresponding to the
depth they occupy.
*/
func (c *Creator) lifelineTitleBoxes() {
	top := c.tideMark
	bot := top + c.lifelineMaker.titleBoxHeight()
	c.lifelineMaker.titleBoxTopAndBottom = &segment{top, bot}

	for _, lifeline := range c.model.LifelineStatements() {
		centre := c.lifelineGeomH.CentreLine(lifeline)
		left := centre - 0.5*c.lifelineGeomH.TitleBoxWidth
		right := centre + 0.5*c.lifelineGeomH.TitleBoxWidth
		c.graphicsModel.Primitives.AddRect(left, top, right, bot)

		n := len(lifeline.LabelSegments)
		firstRowY := bot - float64(n)*c.fontHeight - c.sizer.TitleBoxLabelPadB
		c.rowOfLabels(centre, firstRowY, graphics.Centre, lifeline.LabelSegments)
	}

	c.tideMark += c.lifelineMaker.titleBoxTopAndBottom.Length()
	c.tideMark += c.sizer.TitleBoxPadB
}

/*
interactionLabel generates the labels that sit above one of the horizontal
interaction lines. It then claims the vertical space it has consumed for
itself by advancing the tide mark. And registers this space claim with
the creator's InteractionLineZones component.
*/
func (c *Creator) interactionLabel(
	statement *dsl.Statement) {
	sourceLifeline := statement.ReferencedLifelines[0]
	destLifeline := statement.ReferencedLifelines[1]
	x, horizJustification := c.lifelineGeomH.InteractionLabelPosition(
		sourceLifeline, destLifeline)
	firstRowY := c.tideMark
	c.rowOfLabels(x, firstRowY, horizJustification, statement.LabelSegments)
	c.tideMark += float64(len(statement.LabelSegments))*
		c.fontHeight + c.sizer.InteractionLineTextPadB
	c.ilZones.RegisterSpaceClaim(
		sourceLifeline, destLifeline, firstRowY, c.tideMark)
}

/*
rowOfLabels is a (DRY) helper function to make the graphics.Primitives
objects for the set of strings representing a multi-row label. It hard-codes
the vertical justification (to top), but takes a parameter to specify
horizontal justification. It does not advance the tideMark.
*/
func (c *Creator) rowOfLabels(x float64, firstRowY float64,
	horizJustification graphics.Justification, labelSegments []string) {
	for i, labelSeg := range labelSegments {
		y := firstRowY + float64(i)*c.fontHeight
		c.graphicsModel.Primitives.AddLabel(labelSeg, c.fontHeight,
			x, y, horizJustification, graphics.Top)
	}
}

/*
interactionLine generates the horizontal line and arrow head.  It then claims
the vertical space it needs for itself by advancing the tide mark.  And
registers this space claim with the creator's InteractionLineZones component.
*/
func (c *Creator) interactionLine(
	statement *dsl.Statement) {
	sourceLifeline := statement.ReferencedLifelines[0]
	destLifeline := statement.ReferencedLifelines[1]
	x1, x2 := c.lifelineGeomH.InteractionLineEndPoints(
		sourceLifeline, destLifeline)
	y := c.tideMark
	c.graphicsModel.Primitives.AddLine(x1, y, x2, y,
		statement.Keyword == umli.Dash)
	arrowVertices := makeArrow(x1, x2, y, c.sizer.ArrowLen,
		c.sizer.ArrowHeight)
	c.graphicsModel.Primitives.AddFilledPoly(arrowVertices)
	c.tideMark += 0.5*c.sizer.ArrowHeight + c.sizer.InteractionLinePadB
	c.ilZones.RegisterSpaceClaim(
		sourceLifeline, destLifeline, y, c.tideMark)
}

/*
selfInteractionLines generates a set of *self* interaction lines (i.e. a loop),
including the arrow head and labels. It then claims the vertical space it
has occupied by advancing the tide mark.
*/
func (c *Creator) selfInteractionLines(
	statement *dsl.Statement) {
	theLifeline := statement.ReferencedLifelines[0]

	// First the labels
	centre := c.lifelineGeomH.CentreLine(theLifeline)
	labelX := centre + c.sizer.InteractionLineLabelIndent
	firstRowY := c.tideMark
	c.rowOfLabels(labelX, firstRowY, graphics.Left, statement.LabelSegments)
	n := len(statement.LabelSegments)
	c.tideMark += float64(n) * c.fontHeight
	c.tideMark += c.sizer.InteractionLineTextPadB

	left, right := c.lifelineGeomH.SelfInteractionLineXCoords(theLifeline)
	top := c.tideMark
	bot := c.tideMark + c.sizer.SelfLoopHeight

	prims := c.graphicsModel.Primitives
	prims.AddLine(left, top, right, top, false)
	prims.AddLine(right, top, right, bot, false)
	prims.AddLine(right, bot, left, bot, false)
	arrowVertices := makeArrow(right, left, bot,
		c.sizer.ArrowLen, c.sizer.ArrowHeight)
	prims.AddFilledPoly(arrowVertices)

	c.tideMark = bot + c.sizer.InteractionLinePadB
}

/*
potentiallyStartToBox works out if the Creator has already started a
lifeline activity box for the lifeline that this interaction line is
going to, and if it hasn't it registers where it should start.
*/
func (c *Creator) potentiallyStartToBox(
	statement *dsl.Statement) {
	behindTidemarkDelta := 0.0
	c.potentiallyStartActivityBox(statement.ReferencedLifelines[1],
		behindTidemarkDelta)
}

/*
potentiallyStartFromBox works out if the Creator has already started a
lifeline activity box when an interactionline is about to be emitted
from it. If it hasn't it registers where it should start.
Note it is atypical because it renders behind the tidemark, to position the
start of the box a little before the interaction line, but then leaves the
tidemark unchanged, so that the interaction line that follows, stays in contact
with its label (which has already been emitted).
*/
func (c *Creator) potentiallyStartFromBox(
	statement *dsl.Statement) {
	behindTidemarkDelta := c.sizer.ActivityBoxVerticalOverlap
	c.potentiallyStartActivityBox(statement.ReferencedLifelines[0],
		behindTidemarkDelta)
}

// potentiallyStartActivityBox is a DRY helper to (potentially) note the
// top edge of a lifeline's activity box in c.activityBoxes.
func (c *Creator) potentiallyStartActivityBox(
	lifeline *dsl.Statement, behindTidemarkDelta float64) {
	// Already a box in progress?
	if c.activityBoxes[lifeline].inProgress() {
		return
	}
	y := c.tideMark - behindTidemarkDelta
	c.activityBoxes[lifeline].startBoxAt(y)
}

/*
endBox finishes off a lifeline activity box in response to an
explicit *stop* instruction in the DSL. It then advances the
tide mark to a little beyond the bottom of the box.
*/
func (c *Creator) endBox(
	endBoxStatement *dsl.Statement) {
	lifeline := endBoxStatement.ReferencedLifelines[0]
	bottom := c.tideMark
	c.finalizeActivityBox(lifeline, bottom)
}

// finalizeActivityBoxes identifies lifeline activity boxes that
// have been started, but not *stopped*, and draws them now that
// their size and position can be determined.
func (c *Creator) finalizeActivityBoxes() {
	bottom := c.tideMark + c.sizer.ActivityBoxVerticalOverlap
	for lifeline := range c.activityBoxes {
		c.finalizeActivityBox(lifeline, bottom)
	}
	c.tideMark = bottom + c.sizer.FinalizedActivityBoxesPadB
}

/*
finalizeActivityBox is a DRY helper that draws a single lifeline activity box -
based on the top Y coordinate stored in c.activityBoxes and the given bottom Y
coordinate. It then advances the tide mark to the bottom value provided.
*/
func (c *Creator) finalizeActivityBox(
	lifeline *dsl.Statement, bottom float64) {
	// Skip those that have been stopped earlier explicitly with a *stop*
	// statement.
	if !c.activityBoxes[lifeline].inProgress() {
		return
	}
	top := c.activityBoxes[lifeline].mostRecent().extent.start
	left, _, right := c.lifelineGeomH.ActivityBoxXCoords(lifeline)
	c.graphicsModel.Primitives.AddRect(left, top, right, bottom)
	c.tideMark = bottom
	c.activityBoxes[lifeline].terminateInProgressBoxAt(bottom)
}

// todo
func (c *Creator) finalizeLifelines() {
	// Quite complex - so delegate.
	c.lifelineMaker.produceLifelines()
}
