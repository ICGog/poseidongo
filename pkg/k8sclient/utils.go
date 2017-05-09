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
	"math/rand"
	"time"

	"sync"

	"bytes"
	"encoding/gob"
	"hash/fnv"

	"github.com/golang/glog"
	"github.com/google/uuid"
)

var (
	seedOnce  sync.Once
	uuidMutex sync.Mutex
)

func inti() {

	glog.Info("Init Called")

}

//generateResourceID

// GenerateUUID is seeded by the default random number generator
func GenerateUUID() string {
	var stringUUID string
	// initialize the seed only once
	seedOnce.Do(func() {
		newRandSource := rand.New(rand.NewSource(time.Now().UnixNano()))
		uuid.SetRand(newRandSource) // random number generator to be set once for UUID.
		glog.Info("GenerateUUID: Seeding Once")
	})
	// Need to lock with a mutex
	// Since the Source generated from rand.NewSource is not thread safe.
	uuidMutex.Lock()
	stringUUID = uuid.New().String()
	uuidMutex.Unlock()

	return stringUUID

}

// getBytes returns byte slice for the given value
func getBytes(value interface{}) []byte {
	var byteBuffer bytes.Buffer
	gobEncoder := gob.NewEncoder(&byteBuffer)
	if err := gobEncoder.Encode(value); err != nil {
		glog.Fatalln("utils:getBytes failed to encode value")
		return nil
	}
	return byteBuffer.Bytes()

}

// HashCombine return a hash for any type of values
func HashCombine(valueOne, valueTwo interface{}) uint64 {
	newHash := fnv.New64()

	valueOneByte := getBytes(valueOne)
	valueTwoByte := getBytes(valueTwo)

	if valueOneByte != nil && valueTwoByte != nil {

		newHash.Write(append(valueOneByte, valueTwoByte...))
		return newHash.Sum64()
	}

	return 0
}
