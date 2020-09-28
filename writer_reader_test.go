// Copyright 2018 The Cacophony Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cptv

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripHeaderDefaults(t *testing.T) {
	camera := new(TestCamera)

	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes, camera)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	assert.Equal(t, 2, r.Version())
	assert.True(t, time.Since(r.Timestamp()) < time.Minute) // "now" was used
	assert.Equal(t, "", r.DeviceName())
	assert.Equal(t, "<unknown>", r.FirmwareVersion())
	assert.Equal(t, 0, r.SerialNumber())
	assert.Equal(t, 0, r.DeviceID())
	assert.Equal(t, 0, r.PreviewSecs())
	assert.Equal(t, lepton3.Brand, r.BrandName())
	assert.Equal(t, lepton3.Model, r.ModelName())
	assert.Equal(t, camera.ResX(), r.ResX())
	assert.Equal(t, camera.ResY(), r.ResY())
	assert.Equal(t, lepton3.FramesHz, r.FPS())

	assert.Equal(t, "", r.MotionConfig())
	assert.Equal(t, float32(0.0), r.Latitude())
	assert.Equal(t, float32(0.0), r.Longitude())
	assert.True(t, r.LocTimestamp().IsZero()) // time.Time zero value was used
	assert.Equal(t, float32(0.0), r.Altitude())
	assert.Equal(t, float32(0.0), r.Accuracy())

}

func TestRoundTripHeader(t *testing.T) {
	camera := new(TestCamera)

	ts := time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)
	lts := time.Date(2019, 5, 20, 9, 8, 7, 0, time.UTC)
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes, camera)
	header := Header{
		Timestamp:    ts,
		DeviceName:   "nz42",
		DeviceID:     22,
		PreviewSecs:  8,
		MotionConfig: "keep on movin",
		Latitude:     -36.86667,
		Longitude:    174.76667,
		LocTimestamp: lts,
		Altitude:     200,
		Accuracy:     10,
		Brand:        "Dev",
		Model:        "GP",
		FPS:          camera.FPS(),
		CameraSerial: 1234567890,
		Firmware:     "1.2.3",
	}
	require.NoError(t, w.WriteHeader(header))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	assert.Equal(t, ts, r.Timestamp().UTC())
	assert.Equal(t, "nz42", r.DeviceName())
	assert.Equal(t, 22, r.DeviceID())
	assert.Equal(t, 8, r.PreviewSecs())
	assert.Equal(t, "1.2.3", r.FirmwareVersion())
	assert.Equal(t, 1234567890, r.SerialNumber())
	assert.Equal(t, "keep on movin", r.MotionConfig())
	assert.Equal(t, float32(-36.86667), r.Latitude())
	assert.Equal(t, float32(174.76667), r.Longitude())
	assert.Equal(t, lts, r.LocTimestamp().UTC())
	assert.Equal(t, float32(200), r.Altitude())
	assert.Equal(t, float32(10), r.Accuracy())
	assert.Equal(t, "Dev", r.BrandName())
	assert.Equal(t, "GP", r.ModelName())
	assert.Equal(t, camera.ResX(), r.ResX())
	assert.Equal(t, camera.ResY(), r.ResY())
	assert.Equal(t, camera.FPS(), r.FPS())

}

func TestReaderFrameCount(t *testing.T) {
	camera := new(TestCamera)
	frame := makeTestFrame(camera)
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes, camera)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	c, err := r.FrameCount()
	require.NoError(t, err)
	assert.Equal(t, 3, c)
}

func TestFrameRoundTrip(t *testing.T) {
	tempC := float64(20)
	ffcTemp := float64(25)
	camera := new(TestCamera)
	frame0 := makeTestFrame(camera)
	frame0.Status.TimeOn = 60 * time.Second
	frame0.Status.LastFFCTime = 30 * time.Second
	frame0.Status.TempC = tempC
	frame0.Status.LastFFCTempC = ffcTemp

	frame1 := makeOffsetFrame(camera, frame0)
	frame1.Status.TimeOn = 61 * time.Second
	frame1.Status.LastFFCTime = 31 * time.Second
	frame0.Status.TempC = tempC
	frame0.Status.LastFFCTempC = ffcTemp
	frame2 := makeOffsetFrame(camera, frame1)
	frame2.Status.TimeOn = 62 * time.Second
	frame2.Status.LastFFCTime = 32 * time.Second
	frame0.Status.TempC = tempC
	frame0.Status.LastFFCTempC = ffcTemp
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes, camera)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.WriteFrame(frame0))
	require.NoError(t, w.WriteFrame(frame1))
	require.NoError(t, w.WriteFrame(frame2))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)

	frameD := r.EmptyFrame()
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame0, frameD)
	assert.Equal(t, tempC, frameD.Status.TempC)
	assert.Equal(t, ffcTemp, frameD.Status.LastFFCTempC)

	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame1, frameD)
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame2, frameD)

	assert.Equal(t, io.EOF, r.ReadFrame(frameD))
}
