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

package painter

import (
	"sync"
	"time"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/art/estcows"
	"github.com/xStrom/patriot/art/estville"
	"github.com/xStrom/patriot/log"
	"github.com/xStrom/patriot/sp"
	"github.com/xStrom/patriot/work/shutdown"
)

func Work(wg *sync.WaitGroup, image *art.Image) {
	inFlight := map[int]bool{}
	inFlightLock := sync.Mutex{}
	for {
		// Make sure we have some image data to work with
		if image.Version() == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		shutdown.ShutdownLock.RLock()
		if shutdown.Shutdown {
			shutdown.ShutdownLock.RUnlock()
			log.Infof("Shutting down painter")
			wg.Done()
			break
		}
		shutdown.ShutdownLock.RUnlock()

		inFlightLock.Lock()

		// #1 priority Estville [Bottom right project]
		p := estville.GetWork(image, inFlight)

		// #2 priority Estonian flag [Classic above the fold flag]
		/*
			if p == nil {
				p = estflag.GetWork(image, inFlight)
			}
		*/

		// #2 priority Estonian flag with lovely cows
		if p == nil {
			p = estcows.GetWork(image, inFlight)
		}

		if p != nil {
			inFlight[p.X|(p.Y<<16)] = true
			dc := allocateDrawCall(image.At(p.X, p.Y))
			go func(p *art.Pixel, dc *drawCall) {
				if err := sp.DrawPixel(p.X, p.Y, p.C); err != nil {
					dc.Cancel()
					log.Infof("Failed drawing %v:%v to %v, because: %v", p.X, p.Y, p.C, err)
				} else {
					dc.Finalize()
				}
				time.Sleep(5 * time.Second) // Allow another additional 5 seconds for realtime to update after the request is done
				inFlightLock.Lock()
				delete(inFlight, p.X|(p.Y<<16))
				inFlightLock.Unlock()
			}(p, dc)
		}

		inFlightLock.Unlock()

		// Sleep until we can perform the next move
		if sleep := getTimeUntilNextMove(); sleep > 0 {
			log.Infof("Sleeping %v", sleep)
			time.Sleep(sleep)
		}
	}
}

type drawCall struct {
	time time.Time
	cost int
}

func (dc *drawCall) Cancel() {
	drawcallsLock.Lock()
	defer drawcallsLock.Unlock()
	for i := range drawcalls {
		if drawcalls[i] == dc {
			drawcalls[i], drawcalls = drawcalls[len(drawcalls)-1], drawcalls[:len(drawcalls)-1]
			return
		}
	}
}

func (dc *drawCall) Finalize() {
	drawcallsLock.Lock()
	dc.time = time.Now()
	drawcallsLock.Unlock()
}

var drawcalls []*drawCall
var drawcallsLock sync.Mutex

const (
	scorePerWindow     = 50
	scoreWindow        = 10 * time.Second
	paintOverWhiteCost = 2
	paintOverOtherCost = 5
)

func drawCallCost(oldColor int) int {
	if oldColor == art.White {
		return paintOverWhiteCost
	}
	return paintOverOtherCost
}

func allocateDrawCall(oldColor int) *drawCall {
	dc := &drawCall{cost: drawCallCost(oldColor)}
	drawcallsLock.Lock()
	drawcalls = append(drawcalls, dc)
	drawcallsLock.Unlock()
	return dc
}

func getTimeUntilNextMove() time.Duration {
	drawcallsLock.Lock()

	// Clean up entries older than the configured time window
	now := time.Now()
	for i := 0; i < len(drawcalls); i++ {
		if !drawcalls[i].time.IsZero() && now.Sub(drawcalls[i].time) >= scoreWindow {
			drawcalls[i], drawcalls = drawcalls[len(drawcalls)-1], drawcalls[:len(drawcalls)-1]
			i--
		}
	}

	// Add up the score
	score := 0
	for i := range drawcalls {
		score += drawcalls[i].cost
	}

	drawcallsLock.Unlock()

	// Can we do the next move immediately?
	if scorePerWindow-score >= paintOverOtherCost {
		return 0
	}

	// Determine time until enough free moves for any color surface
	for {
		var until time.Duration
		freeMoves := scorePerWindow - score
		drawcallsLock.Lock()
		now := time.Now()
		for i := range drawcalls {
			if !drawcalls[i].time.IsZero() {
				freeMoves += drawcalls[i].cost
				remaining := drawcalls[i].time.Sub(now)
				if remaining > until {
					until = remaining
				}
				if freeMoves >= paintOverOtherCost {
					drawcallsLock.Unlock()
					return until
				} else {
					log.Infof("Not enough free moves: %v", freeMoves)
				}
			}
		}
		drawcallsLock.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}
