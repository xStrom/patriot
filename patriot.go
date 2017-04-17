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
	"os"
	"os/signal"
	"sync"

	"github.com/xStrom/patriot/log"
	"github.com/xStrom/patriot/realtime"
	"github.com/xStrom/patriot/work"
	"github.com/xStrom/patriot/work/shutdown"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	wg := &sync.WaitGroup{}

	log.Infof("Launching work engine ...")
	wg.Add(1)
	go work.Work(wg)

mainLoop:
	for {
		select {
		case <-interrupt:
			log.Infof("interrupt -- starting shutdown sequence ..")
			shutdown.ShutdownLock.Lock()
			shutdown.Shutdown = true
			shutdown.ShutdownLock.Unlock()
			realtime.Shutdown()
			break mainLoop
		}
	}

	log.Infof("Waiting for clean shutdown ..")
	wg.Wait()
	log.Infof("Clean shutdown done :>")
}
