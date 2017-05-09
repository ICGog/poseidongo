// Poseidon
// Copyright (c) The Poseidon Authors.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// THIS CODE IS PROVIDED ON AN *AS IS* BASIS, WITHOUT WARRANTIES OR
// CONDITIONS OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING WITHOUT
// LIMITATION ANY IMPLIED WARRANTIES OR CONDITIONS OF TITLE, FITNESS FOR
// A PARTICULAR PURPOSE, MERCHANTABLITY OR NON-INFRINGEMENT.
//
// See the Apache Version 2.0 License for specific language governing
// permissions and limitations under the License.

package k8sclient

import (
	"bytes"
	"encoding/gob"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

var (
	seedOnce  sync.Once
	uuidMutex sync.Mutex
)

func GenerateUUID() string {
	var stringUUID string
	// Initialize the seed only once.
	seedOnce.Do(func() {
		uuid.SetRand(rand.New(rand.NewSource(time.Now().UnixNano())))
	})
	// Lock with mutex because the rand source is not thread safe.
	uuidMutex.Lock()
	stringUUID = uuid.New().String()
	uuidMutex.Unlock()
	return stringUUID
}

// getBytes returns byte slice for the given value.
func getBytes(value interface{}) []byte {
	var byteBuffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&byteBuffer)
	if err := gobEncoder.Encode(value); err != nil {
		glog.Fatalln("Failed to encode value")
		return nil
	}
	return byteBuffer.Bytes()
}

func HashCombine(valueOne, valueTwo interface{}) uint64 {
	newHash := fnv.New64()
	valueOneBytes := getBytes(valueOne)
	valueTwoBytes := getBytes(valueTwo)
	newHash.Write(append(valueOneBytes, valueTwoBytes...))
	return newHash.Sum64()
}
