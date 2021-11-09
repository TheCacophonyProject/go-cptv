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
	// "bufio"
	// "bytes"
	"bufio"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
	"time"

	"github.com/spf13/afero"
)

type DualTestFileWriter struct {
	DualWriter
	tempF       afero.File
	compressedF afero.File
	tempWriter  *bufio.Writer
	afs         *afero.Afero
}

func NewDualTestFileWriter(afs *afero.Afero, filename string) (*DualTestFileWriter, error) {

	tempF, err := afs.Create(filename + ".tmp")
	if err != nil {
		return nil, err
	}
	compressedF, err := afs.Create(filename)
	if err != nil {
		return nil, err
	}
	return &DualTestFileWriter{
		tempF:       tempF,
		compressedF: compressedF,
		tempWriter:  bufio.NewWriter(tempF),
		afs:         afs,
	}, nil
}

func (rw *DualTestFileWriter) SeekTemp(offset int64, whence int) (int64, error) {
	return rw.tempF.Seek(offset, whence)
}

func (rw *DualTestFileWriter) FlushTemp() error {
	return rw.tempWriter.Flush()
}
func (rw *DualTestFileWriter) TempReader() io.Reader {
	return bufio.NewReader(rw.tempF)
}
func (rw *DualTestFileWriter) CompressedWriter() *bufio.Writer {
	return bufio.NewWriter(rw.compressedF)
}
func (rw *DualTestFileWriter) TempWriter() io.Writer {
	return rw.tempWriter
}
func (w *DualTestFileWriter) DeleteTemp() error {
	return w.afs.Remove(w.tempF.Name())
}

func (w *DualTestFileWriter) CloseTemp() error {
	return w.tempF.Close()
}
func (w *DualTestFileWriter) CloseCompressed() error {
	return w.compressedF.Close()
}

func NewTestWriter(afs *afero.Afero, filename string, c cptvframe.CameraSpec) (*Writer, error) {
	fileWriter, err := NewDualTestFileWriter(afs, filename)
	if err != nil {
		return nil, err
	}
	tempWriter := fileWriter.TempWriter()

	return &Writer{
		fileWriter: fileWriter,
		name:       filename,
		rw:         tempWriter,
		bldr:       NewBuilder(tempWriter),
		comp:       NewCompressor(c),
	}, nil
}

func TestRoundTripHeaderDefaults(t *testing.T) {
	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	camera := new(TestCamera)
	w, err := NewTestWriter(afs, "test.cptv", camera)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.Close())
	f, err := afs.Open("test.cptv")
	r, err := NewReader(f)
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
	ts := time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)
	lts := time.Date(2019, 5, 20, 9, 8, 7, 0, time.UTC)
	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	camera := new(TestCamera)
	w, err := NewTestWriter(afs, "test.cptv", camera)
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

	f, err := afs.Open("test.cptv")
	r, err := NewReader(f)
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
	assert.False(t, r.HasBackgroundFrame())
}

func TestReaderFrameCount(t *testing.T) {

	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	camera := new(TestCamera)
	w, err := NewTestWriter(afs, "test.cptv", camera)

	require.NoError(t, w.WriteHeader(Header{}))
	frame := makeTestFrame(camera)
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.Close())
	f, err := afs.Open("test.cptv")
	r, err := NewReader(f)
	assert.Equal(t, 2, r.Version())
	assert.Equal(t, uint16(3), r.NumFrames())

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

	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	w, err := NewTestWriter(afs, "test.cptv", camera)

	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.WriteFrame(frame0))
	require.NoError(t, w.WriteFrame(frame1))
	require.NoError(t, w.WriteFrame(frame2))
	require.NoError(t, w.Close())

	f, err := afs.Open("test.cptv")
	r, err := NewReader(f)
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

func TestBackgroundFrame(t *testing.T) {
	tempC := float64(20)
	ffcTemp := float64(25)
	camera := new(TestCamera)
	background := makeTestFrame(camera)
	background.Status.BackgroundFrame = true
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

	fs := afero.NewMemMapFs()
	afs := &afero.Afero{Fs: fs}
	w, err := NewTestWriter(afs, "test.cptv", camera)

	require.NoError(t, w.WriteHeader(Header{BackgroundFrame: background}))
	require.NoError(t, w.WriteFrame(frame0))
	require.NoError(t, w.WriteFrame(frame1))
	require.NoError(t, w.WriteFrame(frame2))
	require.NoError(t, w.Close())

	f, err := afs.Open("test.cptv")
	r, err := NewReader(f)
	require.NoError(t, err)

	frameD := r.EmptyFrame()
	assert.True(t, r.HasBackgroundFrame())
	require.NoError(t, r.ReadFrame(frameD))
	assert.True(t, frameD.Status.BackgroundFrame)
	assert.Equal(t, frameD, background)

	require.NoError(t, r.ReadFrame(frameD))
	assert.False(t, frameD.Status.BackgroundFrame)
	assert.Equal(t, frame0, frameD)
	assert.Equal(t, tempC, frameD.Status.TempC)
	assert.Equal(t, ffcTemp, frameD.Status.LastFFCTempC)

	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame1, frameD)
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame2, frameD)

	assert.Equal(t, io.EOF, r.ReadFrame(frameD))
}
