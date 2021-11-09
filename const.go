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

const (
	magic        = "CPTV"
	version byte = 0x02

	HeaderSection = 'H'
	FrameSection  = 'F'

	// Header field keys
	Timestamp    byte = 'T'
	XResolution  byte = 'X'
	YResolution  byte = 'Y'
	Compression  byte = 'C'
	DeviceName   byte = 'D'
	DeviceID     byte = 'I'
	MotionConfig byte = 'M'
	PreviewSecs  byte = 'P'
	Latitude     byte = 'L'
	Longitude    byte = 'O'
	LocTimestamp byte = 'S'
	Altitude     byte = 'A'
	Accuracy     byte = 'U'
	FPS          byte = 'Z'
	Model        byte = 'E'
	Brand        byte = 'B'
	Firmware     byte = 'V'
	CameraSerial byte = 'N'
	// Frame field keys
	TimeOn          byte = 't'
	BitWidth        byte = 'w'
	FrameSize       byte = 'f'
	LastFFCTime     byte = 'c'
	TempC           byte = 'a'
	LastFFCTempC    byte = 'b'
	BackgroundFrame byte = 'g'
	NumFrames       byte = 'd'
	MaxTemp         byte = 'e'
	MinTemp         byte = 'h'
)
