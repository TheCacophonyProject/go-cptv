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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

type TestCamera struct {
}

func (cam *TestCamera) ResX() int {
	return 200
}

func (cam *TestCamera) ResY() int {
	return 20
}

func (cam *TestCamera) FPS() int {
	return 9
}

const cptvFileName = "v2.cptv"

// Create a frame for playing with.
func makeTestFrame(c cptvframe.CameraSpec) *cptvframe.Frame {
	// Generate a frame with values between 1024 and 8196
	out := cptvframe.NewFrame(c)
	const minVal = 1024
	const maxVal = 8196
	for y, row := range out.Pix {
		for x, _ := range row {
			out.Pix[y][x] = uint16(((y * x) % (maxVal - minVal)) + minVal)
		}
	}

	return out
}

// Create a cptv file for testing purposes
func createCPTVFile(cptvFileName string) {

	camera := new(TestCamera)
	w, _ := cptv.NewWriter(cptvFileName, camera)

	ts := time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)
	lts := time.Date(2019, 5, 20, 9, 8, 7, 0, time.UTC)
	header := cptv.Header{
		Timestamp:    ts,
		DeviceName:   "nz42",
		DeviceID:     90,
		PreviewSecs:  8,
		MotionConfig: "keep on movin",
		Latitude:     -36.86667,
		Longitude:    174.76667,
		LocTimestamp: lts,
		Altitude:     200,
		Accuracy:     10,
	}
	w.WriteHeader(header)

	frame := makeTestFrame(camera)
	w.WriteFrame(frame)
	w.WriteFrame(frame)
	w.WriteFrame(frame)

	w.Close()

}

// Open a cptv file and show the header, number of frames etc.
func openAndDisplayCPTVFileContents(cptvFileName string) {

	file, err := os.Open(cptvFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	r, err := cptv.NewReader(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("CPTV File details:")
	fmt.Println("\tVersion =", r.Version())
	fmt.Println("\tTimeStamp =", r.Timestamp().UTC())
	fmt.Println("\tDeviceName =", r.DeviceName())
	fmt.Println("\tDeviceID =", r.DeviceID())
	fmt.Println("\tPreviewSecs =", r.PreviewSecs())
	fmt.Println("\tMotionConfig =", r.MotionConfig())
	fmt.Println("\tLatitude =", r.Latitude())
	fmt.Println("\tLongitude =", r.Longitude())
	fmt.Println("\tlocTimeStamp =", r.LocTimestamp().UTC())
	fmt.Println("\tAltitude =", r.Altitude())
	fmt.Println("\tAccuracy =", r.Accuracy())
	fmt.Println("\tYResolution =", r.ResY())
	fmt.Println("\tXResolution =", r.ResX())

	frameCount, err := r.FrameCount()
	fmt.Println("\tNum Frames =", frameCount)

}

func main() {
	createCPTVFile(cptvFileName)

	openAndDisplayCPTVFileContents(cptvFileName)
}
