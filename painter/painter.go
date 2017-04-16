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
	"github.com/xStrom/patriot/log"
	"github.com/xStrom/patriot/sp"
	"github.com/xStrom/patriot/work/shutdown"
)

// TODO: Instead of a dumb queue use a coordinate based map so we always only draw the latest requested value

var queue []*art.Pixel
var queueLock sync.Mutex

func Work(wg *sync.WaitGroup) {
	var p *art.Pixel
	for {
		shutdown.ShutdownLock.RLock()
		if shutdown.Shutdown {
			shutdown.ShutdownLock.RUnlock()
			log.Infof("Shutting down painter")
			wg.Done()
			break
		}
		shutdown.ShutdownLock.RUnlock()

		queueLock.Lock()
		if len(queue) > 0 {
			p, queue = queue[0], queue[1:]
		}
		queueLock.Unlock()
		if p != nil {
			if err := sp.DrawPixel(p.X, p.Y, p.C); err != nil {
				log.Infof("Failed drawing %v:%v to %v, because: %v", p.X, p.Y, p.C, err)
				queueLock.Lock()
				queue = append(queue, p)
				queueLock.Unlock()
			}
			p = nil
		}
		time.Sleep(1 * time.Second) // Non-white pixels limited to 1/sec by server (white to 2.5/sec)
	}
}

func SetPixel(p *art.Pixel) {
	queueLock.Lock()
	queue = append(queue, p)
	queueLock.Unlock()
}
