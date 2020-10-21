// Copyright 2020 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package cptvframe

import "time"

type Telemetry struct {
	TimeOn          time.Duration
	FFCState        string
	FrameCount      int
	FrameMean       uint16
	TempC           float64
	LastFFCTempC    float64
	LastFFCTime     time.Duration
	BackgroundFrame bool
}
