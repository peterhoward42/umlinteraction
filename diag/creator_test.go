package diag

import (
	"testing"
	"strings"
	"bufio"

	"github.com/peterhoward42/umlinteraction/parser"
	"github.com/peterhoward42/umlinteraction/graphics"
	"github.com/stretchr/testify/assert"
)

func TestToTeaseOutAPIDuringDevelopment(t *testing.T) {
	assert := assert.New(t)

	reader := strings.NewReader(parser.ReferenceInput)
	p := parser.NewParser()
	statements, err := p.Parse(bufio.NewScanner(reader))
	assert.Nil(err)

	// These widths and heights are chosen to be similar to the size
	// of A4 paper (in mm), to help think about the sizing abstractions.
	width := 200
	fontHeight := 3.0
	creator := NewCreator(width, fontHeight, statements)
    created := creator.Create()

    assert.IsType(&graphics.Model{}, created)
}
