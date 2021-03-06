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

package sp

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/xStrom/patriot/log"
)

const saveImageAsFile = false

const UserAgent = "Patriot/1.1 (https://github.com/xStrom/patriot)"

var client = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   60 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	},
}

func FetchImageFromFile() ([]byte, error) {
	return ioutil.ReadFile("snapshots/current1.png")
}

func FetchImage() ([]byte, int, error) {
	t := time.Now()
	req, err := http.NewRequest("GET", "https://josephg.com/sp/current", nil)
	if err != nil {
		return nil, -1, errors.Wrap(err, "Failed creating request")
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, -1, errors.Wrap(err, "Failed performing request")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, -1, errors.Errorf("Got non-OK status: %v", resp.StatusCode)
	}
	// Extract version
	version := -1
	xcv := resp.Header["X-Content-Version"]
	if len(xcv) > 0 {
		if ver, err := strconv.ParseInt(xcv[0], 10, 64); err != nil {
			return nil, -1, errors.Errorf("Failed to parse X-Content-Version: %v", xcv[0])
		} else {
			version = int(ver)
		}
	}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return nil, -1, errors.Wrap(err, "Failed reading response")
	} else {
		if saveImageAsFile {
			if err := ioutil.WriteFile("out.png", b, os.ModeAppend); err != nil {
				return nil, -1, errors.Wrap(err, "Failed to save fetched image")
			}
		}
		log.Infof("Got image v%v @ %vKB [%v]", version, len(b)/1000, time.Since(t))
		return b, version, nil
	}
}

func DrawPixel(x, y, c int) (error, int) {
	t := time.Now()
	req, err := http.NewRequest("POST", fmt.Sprintf("https://josephg.com/sp/edit?x=%v&y=%v&c=%v", x, y, c), nil)
	if err != nil {
		return errors.Wrap(err, "Failed creating request"), -1
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed performing request"), -1
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("Got non-OK status: %v", resp.StatusCode), resp.StatusCode
	}
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		return errors.Wrap(err, "Failed reading response"), resp.StatusCode
	} else if len(b) > 0 {
		log.Infof("Got response:\n%v", string(b))
	}
	log.Infof("Drew: %v - %v - %v [%dms]", x, y, c, time.Since(t)/time.Millisecond)
	return nil, resp.StatusCode
}
