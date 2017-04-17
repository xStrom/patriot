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

		t := time.Now()
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
			go func(p *art.Pixel) {
				if err := sp.DrawPixel(p.X, p.Y, p.C); err != nil {
					log.Infof("Failed drawing %v:%v to %v, because: %v", p.X, p.Y, p.C, err)
				}
				time.Sleep(5 * time.Second) // Allow another additional 5 seconds for realtime to update after the request is done
				inFlightLock.Lock()
				delete(inFlight, p.X|(p.Y<<16))
				inFlightLock.Unlock()
			}(p)
		}

		inFlightLock.Unlock()

		// TODO: Non-white pixels limited to 1/sec by server (white to 2.5/sec)
		if sleep := 1*time.Second - time.Since(t); sleep > 0 {
			time.Sleep(sleep)
		}
	}
}
