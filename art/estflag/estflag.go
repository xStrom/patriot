// Copyright 2017 Kaur Kuut
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package estflag

import (
	"image"
	"image/color"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/painter"
)

var x0, y0 = 75, 36
var x1, y1 = 107, 56

var blue = color.RGBA64{0, 0, 60138, 65535}
var black = color.RGBA64{8738, 8738, 8738, 65535}
var white = color.RGBA64{65535, 65535, 65535, 65535}

// Fixes any broken pixels in the provided image
func CheckImage(image image.Image) {
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			c := image.At(x, y)

			switch (y - y0) / 7 {
			case 0:
				if !art.SameColor(c, blue) {
					painter.SetPixel(&art.Pixel{x, y, art.DarkBlue})
				}
			case 1:
				if !art.SameColor(c, black) {
					painter.SetPixel(&art.Pixel{x, y, art.Black})
				}
			case 2:
				if !art.SameColor(c, white) {
					painter.SetPixel(&art.Pixel{x, y, art.White})
				}
			}
		}
	}
}

// Fixes the provided pixel if needed
func CheckPixel(x, y, c int) {
	// Make sure the pixel is even in bounds
	if x >= x0 && x <= x1 && y >= y0 && y <= y1 {
		switch (y - y0) / 7 {
		case 0:
			if c != art.DarkBlue {
				painter.SetPixel(&art.Pixel{x, y, art.DarkBlue})
			}
		case 1:
			if c != art.Black {
				painter.SetPixel(&art.Pixel{x, y, art.Black})
			}
		case 2:
			if c != art.White {
				painter.SetPixel(&art.Pixel{x, y, art.White})
			}
		}
	}
}
