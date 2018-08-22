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
// limitations under the License.package main

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

func main() {
	err := runMain()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runMain() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <filename>", os.Args[0])
	}
	file, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer file.Close()

	r, err := cptv.NewReader(file)
	if err != nil {
		return err
	}

	fmt.Println("Timestamp:   ", r.Timestamp())
	fmt.Println("Device Name: ", r.DeviceName())

	// Read the frames and get a frame count. This is an illustration of
	// frame reading - the r.FrameCount method will do the same thing (and
	// will similarly leave the file pointer at EOF)
	frames := 0
	var frame lepton3.Frame
	for {
		err := r.ReadFrame(&frame)
		if err != nil {
			if err == io.EOF {
				fmt.Print("<EOF>")
				frames++ // the last valid read returns EOF
				break
			}
			return err
		}
		frames++
		fmt.Print(".")
		if frames%80 == 0 {
			fmt.Print("\n")
		}
	}
	fmt.Print("\n")
	fmt.Println("Frame Count: ", frames)

	return nil
}
