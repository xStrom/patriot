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

package work

import (
	"sync"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/log"
	"github.com/xStrom/patriot/painter"
	"github.com/xStrom/patriot/realtime"
	"github.com/xStrom/patriot/sp"
	"github.com/xStrom/patriot/work/shutdown"
)

func Work(wg *sync.WaitGroup) {
	img := &art.Image{}

	log.Infof("Launching painter ...")
	wg.Add(1)
	go painter.Work(wg, img)

	for {
		shutdown.ShutdownLock.RLock()
		if shutdown.Shutdown {
			shutdown.ShutdownLock.RUnlock()
			log.Infof("Shutting down work engine")
			wg.Done()
			break
		}
		shutdown.ShutdownLock.RUnlock()

		UpdateImage(img)
		wg.Add(1)
		realtime.Realtime(wg, img)
	}
}

func UpdateImage(img *art.Image) {
start:
	log.Infof("Fetching image ..")
	data, version, err := sp.FetchImage()
	if err != nil {
		log.Infof("Failed to fetch image: %v", err)
		goto start
	}
	if err := img.ParseKeyframe(version, data, false); err != nil {
		panic("Failed to parse image")
	}
}
