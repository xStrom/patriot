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

package art

import (
	"bytes"
	"image/color"
	"image/png"
	"sync"

	"github.com/pkg/errors"

	"github.com/xStrom/patriot/log"
)

const (
	White = iota
	LightGray
	Gray
	Black
	Pink
	Red
	Orange
	Brown
	Yellow
	LightGreen
	Green
	Cyan
	MediumBlue
	DarkBlue
	LightPurple
	DarkPurple
)

var colorToRGBA = map[int]color.RGBA64{
	White:       color.RGBA64{65535, 65535, 65535, 65535},
	LightGray:   color.RGBA64{58596, 58596, 58596, 65535},
	Gray:        color.RGBA64{34952, 34952, 34952, 65535},
	Black:       color.RGBA64{8738, 8738, 8738, 65535},
	Pink:        color.RGBA64{65535, 42919, 53713, 65535},
	Red:         color.RGBA64{58853, 0, 2313, 65535},
	Orange:      color.RGBA64{58853, 38293, 0, 65535},
	Brown:       color.RGBA64{41120, 27242, 16962, 65535},
	Yellow:      color.RGBA64{58853, 55769, 0, 65535},
	LightGreen:  color.RGBA64{38036, 57568, 17476, 65535},
	Green:       color.RGBA64{514, 48830, 257, 65535},
	Cyan:        color.RGBA64{0, 54227, 56797, 65535},
	MediumBlue:  color.RGBA64{0, 33667, 51143, 65535},
	DarkBlue:    color.RGBA64{0, 0, 60138, 65535},
	LightPurple: color.RGBA64{53199, 28270, 58596, 65535},
	DarkPurple:  color.RGBA64{33410, 0, 32896, 65535},
}

type Pixel struct {
	X int
	Y int
	C int
}

type Image struct {
	lock    sync.RWMutex
	version int
	colors  map[int]int
}

func (i *Image) Version() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.version
}

func (i *Image) At(x, y int) int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	if c, ok := i.colors[x|(y<<16)]; ok {
		return c
	}
	return -1
}

func (i *Image) ParseKeyframe(version int, data []byte) error {
	buf := bytes.NewBuffer(data)
	img, err := png.Decode(buf)
	if err != nil {
		return errors.Wrap(err, "Failed to decode image")
	}
	minX := img.Bounds().Min.X
	maxX := img.Bounds().Max.X
	minY := img.Bounds().Min.Y
	maxY := img.Bounds().Max.Y
	if minX != 0 || minY != 0 || maxX != 1000 || maxY != 1000 {
		return errors.New("Unexpected image bounds")
	}
	// Convert colors
	colors := make(map[int]int, (maxX-minX)*(maxY-minY))
	for x := minX; x < maxX; x++ {
		for y := minY; y < maxY; y++ {
			coords := x | (y << 16)
			c := img.At(x, y)
			color := -1
			for cc, rgba := range colorToRGBA {
				if sameColor(c, rgba) {
					color = cc
					break
				}
			}
			colors[coords] = color
			if color == -1 {
				r, g, b, a := c.RGBA()
				log.Infof("Unknown color in keyframe: %v:%v - %v:%v:%v:%v", x, y, r, g, b, a)
			}
		}
	}
	i.lock.Lock()
	if i.version > version {
		log.Infof("New image version is old! %v >= %v", i.version, version)
	}
	i.version = version
	i.colors = colors
	i.lock.Unlock()
	return nil
}

func (i *Image) UpdatePixel(x, y, color, version int) {
	i.lock.Lock()
	if i.version > version {
		log.Infof("New pixel version is old! %v >= %v", i.version, version)
	}
	i.version = version
	i.colors[x|(y<<16)] = color
	i.lock.Unlock()
}

func sameColor(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
