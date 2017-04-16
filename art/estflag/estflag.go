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
	"github.com/xStrom/patriot/art"
)

//const x0, y0 = 75, 36
const x0, y0 = 740, 900
const x1, y1 = x0 + 33 - 1, y0 + 21 - 1

// Fixes any broken pixels in the provided image
func GetWork(image *art.Image, ignorePixels map[int]bool) *art.Pixel {
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			if ignorePixels[x|(y<<16)] {
				continue
			}

			c := image.At(x, y)

			switch (y - y0) / 7 {
			case 0:
				if c != art.DarkBlue {
					return &art.Pixel{x, y, art.DarkBlue}
				}
				break
			case 1:
				if c != art.Black {
					return &art.Pixel{x, y, art.Black}
				}
				break
			case 2:
				if c != art.White {
					return &art.Pixel{x, y, art.White}
				}
				break
			}
		}
	}
	return nil
}

// TODO: Bounds check function, so that not every art needs to be looped through after every pixel update --- make a dirty region system
func CheckPixel(x, y, c int) {
	// Make sure the pixel is even in bounds
	if x >= x0 && x <= x1 && y >= y0 && y <= y1 {

	}
}
