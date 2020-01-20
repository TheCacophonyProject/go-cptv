// Copyright 2020 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package cptvframe

// Interface that all thermal camera implementations should implement
type CameraSpec interface {
	ResX() int
	ResY() int
	FPS() int
}
