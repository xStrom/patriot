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
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/art/resource"
	"github.com/xStrom/patriot/log"
	"github.com/xStrom/patriot/sp"
	"github.com/xStrom/patriot/work/shutdown"
)

type ResourceInfo struct {
	x        int
	y        int
	filepath string
}

var resourceInfos = []*ResourceInfo{
//&ResourceInfo{74, 35, "data/estflag.png"},     // Estonian flag [Classic above the fold flag]
//&ResourceInfo{150, 284, "data/dota.png"},      // Dota 2 logo
//&ResourceInfo{0, 0, "data/acdc.png"},          // AC/DC logo [Top left corner]
//&ResourceInfo{735, 875, "data/estville2.png"}, // Estville [Bottom right project]
//&ResourceInfo{74, 35, "data/estcows.png"},     // Estonian flag with 3rd party cows [Classic above the fold flag position]
}

func Work(wg *sync.WaitGroup, image *art.Image) {
	// Load resources
	log.Infof("Loading resources ..")
	resources := make([]*resource.Resource, 0, len(resourceInfos))
	for _, ri := range resourceInfos {
		r, err := resource.New(ri.x, ri.y, ri.filepath)
		if err != nil {
			panic(fmt.Sprintf("Failed to load resource: %v", err))
		}
		resources = append(resources, r)
	}

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

		// Sleep until we can perform the next move
		sleepUntilNextMove()

		inFlightLock.Lock()

		var p *art.Pixel
		for _, r := range resources {
			if p = r.GetWork(image, inFlight); p != nil {
				break
			}
		}

		if p != nil {
			inFlight[p.X|(p.Y<<16)] = true
			cost := drawCallCost(image.At(p.X, p.Y))
			cs := addCycleCost(cost)
			go func(p *art.Pixel, cs int64, cost int) {
				//log.Infof("Requesting draw of %v:%v - %v", p.X, p.Y, p.C)
				if err, statusCode := sp.DrawPixel(p.X, p.Y, p.C); err != nil {
					// Don't remove the cycle cost in case of 403, because that means we hit the server rate limiting
					if statusCode != http.StatusForbidden {
						removeCycleCost(cs, cost)
					}
					log.Infof("Failed drawing %v:%v to %v, because: %v", p.X, p.Y, p.C, err)
				}
				time.Sleep(5 * time.Second) // Allow another additional 5 seconds for realtime to update after the request is done
				inFlightLock.Lock()
				delete(inFlight, p.X|(p.Y<<16))
				inFlightLock.Unlock()
			}(p, cs, cost)
		}

		inFlightLock.Unlock()

		// Prevent hot spin if there's nothing to do
		if p == nil {
			time.Sleep(1 * time.Second)
		}
	}
}

const (
	scorePerWindow     = 30
	scoreWindowSecs    = 10
	paintOverWhiteCost = 2
	paintOverOtherCost = 5
)

func drawCallCost(oldColor int) int {
	if oldColor == art.White {
		return paintOverWhiteCost
	}
	return paintOverOtherCost
}

var cycleCost int
var cycleStart int64
var cycleLock sync.Mutex

func addCycleCost(cost int) int64 {
	cycleLock.Lock()
	defer cycleLock.Unlock()
	cycleCost += cost
	return cycleStart
}

func removeCycleCost(start int64, cost int) {
	cycleLock.Lock()
	defer cycleLock.Unlock()
	if cycleStart == start {
		cycleCost -= cost
	}
}

func sleepUntilNextMove() {
	for {
		cycleLock.Lock()

		now := time.Now().Unix()
		if cycleStart+scoreWindowSecs <= now {
			cycleStart = now
			cycleCost = 0
			//log.Infof("New cycle started at %v", cycleStart)
			cycleLock.Unlock()
			return
		}

		// Can we do the next move?
		if scorePerWindow-cycleCost >= paintOverOtherCost {
			//log.Infof("Can still do another move (%v/%v)", cycleCost, scorePerWindow)
			cycleLock.Unlock()
			return
		}

		cycleLock.Unlock()

		time.Sleep(1 * time.Second)
	}
}
