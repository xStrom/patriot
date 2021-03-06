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

package realtime

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/xStrom/patriot/art"
	"github.com/xStrom/patriot/log"
)

var c *websocket.Conn
var done chan struct{}

func Realtime(wg *sync.WaitGroup, image *art.Image) {
	done = make(chan struct{})
	u := url.URL{Scheme: "wss", Host: "josephg.com", Path: "/sp/ws", RawQuery: fmt.Sprintf("from=%v", image.Version())}

connect:
	log.Infof("connecting to %s", u.String())
	var err error
	c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Infof("dial err: %v", err)
		goto connect
	}

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Infof("read error: %v", err)
			break
		}
		if bytes.Compare(message, []byte("reload")) == 0 {
			log.Infof("Got reload command")
			break
		}
		if bytes.Compare(message, []byte("refresh")) == 0 {
			log.Infof("Got refresh command")
			break
		}
		if len(message) >= 7 {
			version := int(binary.LittleEndian.Uint32(message[0:4]))
			for i := 4; i < len(message); i += 3 {
				if len(message) >= i+3 {
					x, y, color := decodeEdit(message[i : i+3])
					image.UpdatePixel(x, y, color, version)
				} else {
					log.Infof("recv unknown suffix on edits: %v", message[i:])
				}
			}
		} else {
			log.Infof("recv unknown: %v", message)
		}
	}

	log.Infof("Close in Realtime")
	c.Close()
	close(done)
	wg.Done()
}

// returns x, y, color.
func decodeEdit(data []byte) (int, int, int) {
	xx := uint(data[0])
	yx := uint(data[1])
	cy := uint(data[2])

	x := xx | ((yx & 0x3) << 8)
	y := (yx >> 2) | ((cy & 0xf) << 6)
	color := cy >> 4

	return int(x), int(y), int(color)
}

// TODO: Currently c isn't set to nil (need sync for that anyway), so Shutdown could be called for old closed connection
func Shutdown() {
	if c == nil {
		return
	}
	// To cleanly close a connection, a client should send a close
	// frame and wait for the server to close the connection.
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Infof("write close error: %v", err)
		return
	}
	select {
	case <-done:
	case <-time.After(time.Second):
	}
	log.Infof("Close in Shutdown")
	err = c.Close()
	if err != nil {
		log.Infof("close error: %v", err)
	}
}
