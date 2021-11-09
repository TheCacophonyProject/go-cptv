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
	// "compress/gzip"
	"io"
)

// NewBuilder returns a new Builder instance, ready to emit a gzip
// compressed CPTV file to the provided Writer.
func NewBuilder(w io.Writer) *Builder {
	return &Builder{
		w: w,
		// w: gzip.NewWriter(w),
	}
}

// Builder handles the low-level construction of CPTV sections and
// fields. See Writer for a higher-level interface.
type Builder struct {
	w           io.Writer
	fieldOffset int64
}

// WriteHeader writes a CPTV header to the current Writer
func (b *Builder) WriteHeader(f *FieldWriter) error {
	fieldData, numFields := f.Bytes()
	pre := append(
		[]byte(magic),
		version,
		HeaderSection,
		byte(numFields),
	)
	_, err := b.w.Write(pre)
	if err != nil {
		return err
	}
	b.fieldOffset = int64(len(pre))
	_, err = b.w.Write(fieldData)
	return err
}

// WriteFrame writes a CPTV frame to the current Writer
func (b *Builder) WriteFrame(f *FieldWriter, frameData []byte) error {
	// Frame header
	fieldData, numFields := f.Bytes()
	_, err := b.w.Write([]byte{FrameSection, byte(numFields)})
	if err != nil {
		return err
	}

	// Frame fields
	_, err = b.w.Write(fieldData)
	if err != nil {
		return err
	}

	// Frame thermal data
	_, err = b.w.Write(frameData)
	return err
}

// Close closes the current Writer
func (b *Builder) Close() error {
	// TO DFO GP
	return nil
	// if err := b.w.Flush(); err != nil {
	// 	return err
	// }
	// return nil
}
