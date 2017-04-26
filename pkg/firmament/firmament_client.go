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

package firmament

import (
	"google.golang.org/grpc"
)

type firmamentConfig struct {
	address string
}

func schedule(client FirmamentSchedulerClient) {
}

func taskCompleted(client FirmamentSchedulerClient) {
}

func taskFailed(client FirmamentSchedulerClient) {
}

func taskRemoved(client FirmamentSchedulerClient) {
}

func taskSubmitted(client FirmamentSchedulerClient) {
}

func nodeAdded(client FirmamentSchedulerClient) {
}

func nodeFailed(client FirmamentSchedulerClient) {
}

func nodeRemoved(client FirmamentSchedulerClient) {
}

func addTaskStats(client FirmamentSchedulerClient) {
}

func addNodeStats(client FirmamentSchedulerClient) {
}

func New(firmamentConfig config) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err = grpc.Dial(*config.address, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	firmamentClient := NewFirmamentSchedulerClient(conn)
}
