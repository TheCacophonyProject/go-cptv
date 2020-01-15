// Copyright 2020 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package cptvframe

// Frame represents the thermal readings for a single frame.
type Frame struct {
	Pix    [][]uint16
	Status Telemetry
}

// Creates a new frame sized for the provided camera implementation
func NewFrame(c CameraSpec) *Frame {
	frame := new(Frame)
	frame.Pix = make([][]uint16, c.ResY())
	for i := range frame.Pix {
		frame.Pix[i] = make([]uint16, c.ResX())
	}
	return frame
}

// Copy sets current frame as other frame
func (fr *Frame) Copy(orig *Frame) {
	fr.Status = orig.Status
	for y, row := range orig.Pix {
		copy(fr.Pix[y][:], row)
	}
}
