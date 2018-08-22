package main

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

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/TheCacophonyProject/go-cptv"
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
	bfile := bufio.NewReader(file)
	r, err := cptv.NewParser(bfile)
	if err != nil {
		return err
	}
	fields, err := r.Header()
	if err != nil {
		return err
	}

	ts, err := fields.Timestamp(cptv.Timestamp)
	fmt.Println("Timestamp:   ", ts, err)
	xres, err := fields.Uint32(cptv.XResolution)
	fmt.Println("X Resolution:", xres, err)
	yres, err := fields.Uint32(cptv.YResolution)
	fmt.Println("YResolution: ", yres, err)
	cmpr, err := fields.Uint8(cptv.Compression)
	fmt.Println("Compression: ", cmpr, err)
	devc, err := fields.String(cptv.DeviceName)
	fmt.Println("DeviceName: ", devc, err)

	frames := 0
	for {
		fields, frameReader, err := r.Frame()

		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// skip past the frame
		frameLen, err := fields.Uint32(cptv.FrameSize)
		buf := make([]byte, frameLen)
		bytesLeft := frameLen
		for frameLen > 0 {
			n, err := frameReader.Read(buf)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			bytesLeft = bytesLeft - uint32(n)
		}
		frames++
	}
	fmt.Println("Frames:      ", frames)
	return nil
}
