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

package estcows

import (
	"io/ioutil"

	"github.com/xStrom/patriot/art"
)

const w, h = 35, 23
const x0, y0 = 74, 35
const x1, y1 = x0 + w - 1, y0 + h - 1

var resource = &art.Image{}

func init() {
	data, err := ioutil.ReadFile("data/estcows.png")
	if err != nil {
		panic("Failed to read estcows.png")
	}
	err = resource.ParseKeyframe(1, data, true)
	if err != nil {
		panic("Failed to parse estcows.png")
	}
}

// Fixes any broken pixels in the provided image
func GetWork(image *art.Image, ignorePixels map[int]bool) *art.Pixel {
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			if ignorePixels[x|(y<<16)] {
				continue
			}
			c1 := resource.At(x-x0, y-y0)
			if c1 == art.Transparent {
				continue
			}
			c2 := image.At(x, y)
			if c1 != c2 {
				return &art.Pixel{x, y, c1}
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
