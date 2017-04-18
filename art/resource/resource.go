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

package resource

import (
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/xStrom/patriot/art"
)

type Resource struct {
	img *art.Image
	x0  int
	x1  int
	y0  int
	y1  int
}

func New(x, y int, filepath string) (*Resource, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read file")
	}
	img := &art.Image{}
	err = img.ParseKeyframe(1, data, true)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse image")
	}
	w, h := img.Dimensions()
	r := &Resource{
		img: img,
		x0:  x,
		x1:  x + w - 1,
		y0:  y,
		y1:  y + h - 1,
	}
	return r, nil
}

// Fixes any broken pixels in the provided image
func (r *Resource) GetWork(image *art.Image, ignorePixels map[int]bool) *art.Pixel {
	for x := r.x0; x <= r.x1; x++ {
		for y := r.y0; y <= r.y1; y++ {
			if ignorePixels[x|(y<<16)] {
				continue
			}
			c1 := r.img.At(x-r.x0, y-r.y0)
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
func (r *Resource) CheckPixel(x, y, c int) {
	// Make sure the pixel is even in bounds
	if x >= r.x0 && x <= r.x1 && y >= r.y0 && y <= r.y1 {

	}
}
