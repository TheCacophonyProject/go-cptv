// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package cptvframe

import (
	"time"
)

type Telemetry struct {
	TimeOn       time.Duration
	FFCState     string
	FrameCount   int
	FrameMean    uint16
	TempC        float64
	LastFFCTempC float64
	LastFFCTime  time.Duration
}

// Frame represents the thermal readings for a single frame.
type Frame struct {
	Pix    [][]uint16
	Status Telemetry
}

type CameraResolution interface {
	ResX() int
	ResY() int
	// FPS() int
}

func NewFrame(c CameraResolution) *Frame {
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
