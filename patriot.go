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

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const UserAgent = "Patriot/1.0 (https://github.com/xStrom/patriot)"

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
	Lime
	Green
	Aqua
	LightBlue
	Blue
	DarkPink
	Purple
)

func main() {
	fmt.Println("Launching queue handler ...")
	go executeQueue()

	fetchAndCheck()

	fmt.Println("Waiting for queue to be empty ...")
	for {
		queueLock.Lock()
		if len(queue) > 0 {
			fmt.Printf("Still have %v items in queue ..\n", len(queue))
		} else {
			queueLock.Unlock()
			break
		}
		queueLock.Unlock()
		time.Sleep(10 * time.Second)
	}
	fmt.Println("All done!")
}

func drawPixel(x, y, c int) error {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://josephg.com/sp/edit?x=%v&y=%v&c=%v", x, y, c), nil)
	if err != nil {
		return errors.Wrap(err, "Failed creating request")
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed performing request")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Got non-OK status: %v\n", resp.StatusCode)
	}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return errors.Wrap(err, "Failed reading response")
	} else if len(b) > 0 {
		fmt.Printf("Got response:\n%v\n", string(b))
	}
	fmt.Printf("Drew: %v - %v - %v\n", x, y, c)
	return nil
}

type Work struct {
	X int
	Y int
	C int
}

var queue []*Work
var queueLock sync.Mutex

func executeQueue() {
	var w *Work
	for {
		queueLock.Lock()
		if len(queue) > 0 {
			w, queue = queue[0], queue[1:]
		}
		queueLock.Unlock()
		if w != nil {
			if err := drawPixel(w.X, w.Y, w.C); err != nil {
				fmt.Printf("Failed drawing %v:%v to %v, because: %v", w.X, w.Y, w.C, err)
				queueLock.Lock()
				queue = append(queue, w)
				queueLock.Unlock()
			}
			w = nil
		}
		time.Sleep(1 * time.Second)
	}
}

func addToQueue(w *Work) {
	queueLock.Lock()
	queue = append(queue, w)
	queueLock.Unlock()
}

func fetchAndCheck() {
start:
	fmt.Printf("Fetching image ..\n")
	data, err := fetchImage()
	if err != nil {
		fmt.Printf("Failed to fetch image: %v\n", err)
		goto start
	}
	img, err := parseImage(data)
	if err != nil {
		panic("Failed to parse image")
	}
	checkFlag(img)
}

func getTestImage() ([]byte, error) {
	return ioutil.ReadFile("current.png")
}

func fetchImage() ([]byte, error) {
	req, err := http.NewRequest("GET", "https://josephg.com/sp/current", nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed creating request")
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "Failed performing request")
	}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, errors.Wrap(err, "Failed reading response")
	} else {
		return b, nil
	}
}

func parseImage(data []byte) (image.Image, error) {
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

func checkFlag(image image.Image) {
	x0, y0 := 75, 36
	x1, y1 := 107, 56

	blue := color.RGBA64{0, 0, 60138, 65535}
	black := color.RGBA64{8738, 8738, 8738, 65535}
	white := color.RGBA64{65535, 65535, 65535, 65535}

	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			c := image.At(x, y)

			switch (y - y0) / 7 {
			case 0:
				if !sameColor(c, blue) {
					addToQueue(&Work{x, y, Blue})
				}
			case 1:
				if !sameColor(c, black) {
					addToQueue(&Work{x, y, Black})
				}
			case 2:
				if !sameColor(c, white) {
					addToQueue(&Work{x, y, White})
				}
			}
		}
	}
}

func sameColor(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
