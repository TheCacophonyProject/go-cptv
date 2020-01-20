// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package cptvframe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestCamera struct {
}

func NewTestCamera() *TestCamera {
	return new(TestCamera)
}
func (cam *TestCamera) ResX() int {
	return 160
}
func (cam *TestCamera) ResY() int {
	return 320
}
func (cam *TestCamera) FPS() int {
	return 9
}
func TestFrameCopy(t *testing.T) {

	camera := new(TestCamera)
	frame := NewFrame(camera)
	// Pixel values.
	frame.Pix[0][0] = 1
	frame.Pix[9][7] = 2
	frame.Pix[camera.ResY()-1][0] = 3
	frame.Pix[0][camera.ResX()-1] = 4
	frame.Pix[camera.ResY()-1][camera.ResX()-1] = 5
	// Status values.
	frame.Status.TimeOn = 10 * time.Second
	frame.Status.FrameCount = 123
	frame.Status.TempC = 23.1

	frame2 := NewFrame(camera)
	frame2.Copy(frame)

	assert.Equal(t, 1, int(frame2.Pix[0][0]))
	assert.Equal(t, 2, int(frame2.Pix[9][7]))
	assert.Equal(t, 3, int(frame2.Pix[camera.ResY()-1][0]))
	assert.Equal(t, 4, int(frame2.Pix[0][camera.ResX()-1]))
	assert.Equal(t, 5, int(frame2.Pix[camera.ResY()-1][camera.ResX()-1]))
	assert.Equal(t, frame.Status, frame2.Status)
}
