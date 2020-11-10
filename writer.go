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
	"io"
	"time"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

// NewWriter creates and returns a new Writer component
func NewWriter(w io.Writer, c cptvframe.CameraSpec) *Writer {
	return &Writer{
		bldr: NewBuilder(w),
		comp: NewCompressor(c),
	}
}

// Writer uses a Builder and Compressor to create CPTV files.
type Writer struct {
	bldr *Builder
	comp *Compressor
}

// Header defines the information stored in the header of a CPTV
// file. All the fields are optional.
type Header struct {
	Timestamp       time.Time
	DeviceName      string
	DeviceID        int
	CameraSerial    int
	Firmware        string
	PreviewSecs     int
	MotionConfig    string
	Latitude        float32
	Longitude       float32
	LocTimestamp    time.Time
	Altitude        float32
	Accuracy        float32
	FPS             int
	Brand           string
	Model           string
	BackgroundFrame *cptvframe.Frame
}

// WriteHeader writes a CPTV file header
func (w *Writer) WriteHeader(header Header) error {
	t := header.Timestamp
	if t.IsZero() {
		t = time.Now()
	}
	fields := NewFieldWriter()
	fields.Timestamp(Timestamp, t)
	fields.Uint32(XResolution, uint32(w.comp.cols))
	fields.Uint32(YResolution, uint32(w.comp.rows))
	fields.Uint8(Compression, 1)
	fields.Uint32(CameraSerial, uint32(header.CameraSerial))

	if len(header.DeviceName) > 0 {
		err := fields.String(DeviceName, header.DeviceName)
		if err != nil {
			return err
		}
	}

	if len(header.Firmware) > 0 {
		err := fields.String(Firmware, header.Firmware)
		if err != nil {
			return err
		}
	}

	if len(header.Model) > 0 {
		err := fields.String(Model, header.Model)
		if err != nil {
			return err
		}
	}
	if len(header.Brand) > 0 {
		err := fields.String(Brand, header.Brand)
		if err != nil {
			return err
		}
	}

	if header.FPS > 0 {
		fields.Uint8(FPS, uint8(header.FPS))
	}

	if header.DeviceID > 0 {
		fields.Uint32(DeviceID, uint32(header.DeviceID))
	}

	fields.Uint8(PreviewSecs, uint8(header.PreviewSecs))

	if len(header.MotionConfig) > 0 {
		err := fields.String(MotionConfig, header.MotionConfig)
		if err != nil {
			return err
		}
	}

	// The location fields are optional. They are only written out if they have non-zero values.
	if header.Latitude != 0.0 {
		fields.Float32(Latitude, header.Latitude)
	}
	if header.Longitude != 0.0 {
		fields.Float32(Longitude, header.Longitude)
	}
	if !header.LocTimestamp.IsZero() {
		fields.Timestamp(LocTimestamp, header.LocTimestamp)
	}
	if header.Altitude >= 0.0 {
		fields.Float32(Altitude, header.Altitude)
	}
	if header.Accuracy != 0.0 {
		fields.Float32(Accuracy, header.Accuracy)
	}
	if header.BackgroundFrame != nil {
		fields.Uint8(BackgroundFrame, 1)
	}
	err := w.bldr.WriteHeader(fields)
	if err != nil {
		return err
	}

	if header.BackgroundFrame != nil {
		header.BackgroundFrame.Status.BackgroundFrame = true
		return w.WriteFrame(header.BackgroundFrame)
	}
	return err
}

// WriteFrame writes a CPTV frame
func (w *Writer) WriteFrame(frame *cptvframe.Frame) error {
	bitWidth, compFrame := w.comp.Next(frame)
	fields := NewFieldWriter()
	if frame.Status.BackgroundFrame {
		fields.Uint8(BackgroundFrame, uint8(1))
	} else {
		fields.Uint32(TimeOn, durationToMillis(frame.Status.TimeOn))
		fields.Uint32(LastFFCTime, durationToMillis(frame.Status.LastFFCTime))
		fields.Float32(TempC, float32(frame.Status.TempC))
		fields.Float32(LastFFCTempC, float32(frame.Status.LastFFCTempC))
	}
	fields.Uint8(BitWidth, uint8(bitWidth))
	fields.Uint32(FrameSize, uint32(len(compFrame)))
	return w.bldr.WriteFrame(fields, compFrame)
}

// Close closes the CPTV file
func (w *Writer) Close() error {
	return w.bldr.Close()
}

func durationToMillis(d time.Duration) uint32 {
	return uint32(d / time.Millisecond)
}
