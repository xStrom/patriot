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
	"fmt"
	"sync"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/art/estflag"
	"github.com/xStrom/patriot/painter"
	"github.com/xStrom/patriot/realtime"
	"github.com/xStrom/patriot/sp"
	"github.com/xStrom/patriot/work/shutdown"
)

func Work(wg *sync.WaitGroup) {
	fmt.Println("Launching painter ...")
	wg.Add(1)
	go painter.Work(wg)

	for {
		shutdown.ShutdownLock.RLock()
		if shutdown.Shutdown {
			shutdown.ShutdownLock.RUnlock()
			fmt.Printf("Shutting down work engine\n")
			wg.Done()
			break
		}
		shutdown.ShutdownLock.RUnlock()

		version := FetchImageAndCheck()
		wg.Add(1)
		realtime.Realtime(wg, version)
	}
}

func FetchImageAndCheck() int {
start:
	fmt.Printf("Fetching image ..\n")
	data, version, err := sp.FetchImage()
	if err != nil {
		fmt.Printf("Failed to fetch image: %v\n", err)
		goto start
	}
	img, err := art.ParseImage(data)
	if err != nil {
		panic("Failed to parse image")
	}
	estflag.CheckImage(img)
	return version
}
