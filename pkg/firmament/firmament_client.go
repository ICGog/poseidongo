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
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

func Schedule(client FirmamentSchedulerClient) {
	// TODO(ionel): Implement!
}

func TaskCompleted(client FirmamentSchedulerClient, tuid *TaskUID) {
	_, err := client.TaskCompleted(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskCompleted(_) = _, %v: ", client, err)
	}
}

func TaskFailed(client FirmamentSchedulerClient, tuid *TaskUID) {
	_, err := client.TaskFailed(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskFailed(_) = _, %v: ", client, err)
	}
}

func TaskRemoved(client FirmamentSchedulerClient, tuid *TaskUID) {
	_, err := client.TaskRemoved(context.Background(), tuid)
	if err != nil {
		grpclog.Fatalf("%v.TaskRemoved(_) = _, %v: ", client, err)
	}
}

func TaskSubmitted(client FirmamentSchedulerClient, td *TaskDescription) {
	_, err := client.TaskSubmitted(context.Background(), td)
	if err != nil {
		grpclog.Fatalf("%v.TaskSubmitted(_) = _, %v: ", client, err)
	}
}

func NodeAdded(client FirmamentSchedulerClient, rtnd *ResourceTopologyNodeDescriptor) {
	_, err := client.NodeAdded(context.Background(), rtnd)
	if err != nil {
		grpclog.Fatalf("%v.NodeAdded(_) = _, %v: ", client, err)
	}
}

func NodeFailed(client FirmamentSchedulerClient, ruid *ResourceUID) {
	_, err := client.NodeFailed(context.Background(), ruid)
	if err != nil {
		grpclog.Fatalf("%v.NodeFailed(_) = _, %v: ", client, err)
	}
}

func NodeRemoved(client FirmamentSchedulerClient, ruid *ResourceUID) {
	_, err := client.NodeRemoved(context.Background(), ruid)
	if err != nil {
		grpclog.Fatalf("%v.NodeRemoved(_) = _, %v: ", client, err)
	}
}

func AddTaskStats(client FirmamentSchedulerClient, ts *TaskStats) {
	_, err := client.AddTaskStats(context.Background(), ts)
	if err != nil {
		grpclog.Fatalf("%v.AddTaskStats(_) = _, %v: ", client, err)
	}
}

func AddNodeStats(client FirmamentSchedulerClient, rs *ResourceStats) {
	_, err := client.AddNodeStats(context.Background(), rs)
	if err != nil {
		grpclog.Fatalf("%v.AddNodeStats(_) = _, %v: ", client, err)
	}
}

func New(address string) (FirmamentSchedulerClient, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	fc := NewFirmamentSchedulerClient(conn)
	return fc, nil
}
