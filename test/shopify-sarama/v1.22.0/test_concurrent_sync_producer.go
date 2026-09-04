// Copyright (c) 2026 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import "sync"

func main() {
	const producerCount = 16

	var wg sync.WaitGroup
	wg.Add(producerCount)
	for range producerCount {
		go func() {
			defer wg.Done()
			producer, err := createSyncProducer()
			if err != nil {
				panic(err)
			}
			if err := producer.Close(); err != nil {
				panic(err)
			}
		}()
	}
	wg.Wait()
}
