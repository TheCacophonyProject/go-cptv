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
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"time"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

type DualWriter interface {
	SeekTemp(offset int64, whence int) (int64, error)
	FlushTemp() error
	CompressedWriter() *bufio.Writer
	TempWriter() io.Writer
	TempReader() io.Reader

	CloseTemp() error
	CloseCompressed() error
	DeleteTemp() error
}

type DualFileWriter struct {
	DualWriter
	tempF       *os.File
	compressedF *os.File
	tempWriter  *bufio.Writer
}

func NewDualFileWriter(filename string) (*DualFileWriter, error) {
	tempF, err := os.Create(filename + ".tmp")
	if err != nil {
		return nil, err
	}
	compressedF, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &DualFileWriter{
		tempF:       tempF,
		compressedF: compressedF,
		tempWriter:  bufio.NewWriter(tempF),
	}, nil
}

func (rw *DualFileWriter) SeekTemp(offset int64, whence int) (int64, error) {
	return rw.tempF.Seek(offset, whence)
}

func (rw *DualFileWriter) FlushTemp() error {
	return rw.tempWriter.Flush()
}
func (rw *DualFileWriter) TempReader() io.Reader {
	return bufio.NewReader(rw.tempF)
}
func (rw *DualFileWriter) CompressedWriter() *bufio.Writer {
	return bufio.NewWriter(rw.compressedF)
}
func (rw *DualFileWriter) TempWriter() io.Writer {
	return rw.tempWriter
}
func (w *DualFileWriter) DeleteTemp() error {
	return os.Remove(w.tempF.Name())
}

func (w *DualFileWriter) CloseTemp() error {
	return w.tempF.Close()
}
func (w *DualFileWriter) CloseCompressed() error {
	return w.compressedF.Close()
}

// NewWriter creates and returns a new Writer component
func NewWriter(filename string, c cptvframe.CameraSpec) (*Writer, error) {
	fileWriter, err := NewDualFileWriter(filename)
	if err != nil {
		return nil, err
	}
	tempWriter := fileWriter.TempWriter()
	return &Writer{
		fileWriter: fileWriter,
		rw:         tempWriter,
		bldr:       NewBuilder(tempWriter),
		comp:       NewCompressor(c),
	}, nil
}

// Writer uses a Builder and Compressor to create CPTV files.
type Writer struct {
	fileWriter DualWriter
	rw         io.Writer
	bldr       *Builder
	comp       *Compressor
	frames     uint16
	maxP       uint16
	minP       uint16
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
	// Placeholders these get written on close
	fields.Uint16(NumFrames, 0)
	fields.Uint16(MaxTemp, 0)
	fields.Uint16(MinTemp, 0)
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
	w.frames += 1
	bitWidth, maxP, minP, compFrame := w.comp.Next(frame)
	if w.minP == 0 || minP < w.minP {
		w.minP = minP
	}
	if w.maxP == 0 || maxP > w.maxP {
		w.maxP = maxP
	}
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

// Compress and Close closes the CPTV file
func (w *Writer) Close() error {
	err := w.Compress()
	if err != nil {
		return err
	}
	w.fileWriter.CloseTemp()
	return w.fileWriter.DeleteTemp()
}

func (w *Writer) Compress() error {
	w.fileWriter.FlushTemp()

	fields := NewFieldWriter()
	fields.Uint16(NumFrames, w.frames)
	fields.Uint16(MaxTemp, w.maxP)
	fields.Uint16(MinTemp, w.minP)
	b, _ := fields.Bytes()
	w.fileWriter.SeekTemp(w.bldr.fieldOffset, 0)
	w.rw.Write(b)
	w.fileWriter.FlushTemp()

	cw := w.fileWriter.CompressedWriter()
	compressor := gzip.NewWriter(cw)
	w.fileWriter.SeekTemp(0, 0)
	_, err := io.Copy(compressor, bufio.NewReader(w.fileWriter.TempReader()))
	if err != nil {
		return err
	}
	compressor.Flush()
	compressor.Close()
	cw.Flush()
	return w.fileWriter.CloseCompressed()
}

func durationToMillis(d time.Duration) uint32 {
	return uint32(d / time.Millisecond)
}
