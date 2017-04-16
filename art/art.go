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
	"image"
	"image/color"
	"image/png"

	"github.com/pkg/errors"
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

type Pixel struct {
	X int
	Y int
	C int
}

func ParseImage(data []byte) (image.Image, error) {
	buf := bytes.NewBuffer(data)
	img, err := png.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to decode image")
	}
	if img.Bounds().Min.X != 0 || img.Bounds().Min.Y != 0 || img.Bounds().Max.X != 1000 || img.Bounds().Max.Y != 1000 {
		return nil, errors.New("Unexpected image bounds")
	}
	return img, nil
}

func SameColor(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
