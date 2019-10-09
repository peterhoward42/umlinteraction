/*
Package sizer is concerned with sizing diagram elements.

Not only how big things are, but also how far apart they should be.

E.g. the size of a lifeline title box, the mark-space settings for
dashed lines, arrow sizing, the margins or padding required for each thing etc.
*/
package sizer

// Sizer defines the contract for a thing that can provide sizes.
type Sizer interface {

	// Get returns the size specified by propertyName, or panics
	// if the property is not recognized. (It is by definition a programming
	// error if this happens)
	Get(propertyName string) (size float64)
}