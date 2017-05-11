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
	"github.com/ICGog/poseidongo/pkg/firmament"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const bytesToKb = 1024

var PodToTD map[PodIdentifier]*firmament.TaskDescriptor
var TaskIDToPod map[uint64]PodIdentifier
var jobIDToJD map[string]*firmament.JobDescriptor
var jobNumIncompleteTasks map[string]int
var NodeToRTND map[string]*firmament.ResourceTopologyNodeDescriptor
var ResIDToNode map[string]string

type NodePhase string

const (
	NodeAdded   NodePhase = "Added"
	NodeDeleted NodePhase = "Deleted"
	NodeFailed  NodePhase = "Failed"
	NodeUpdated NodePhase = "Updated"
)

type Node struct {
	Hostname         string
	Phase            NodePhase
	IsReady          bool
	IsOutOfDisk      bool
	CpuCapacity      int64
	CpuAllocatable   int64
	MemCapacityKb    int64
	MemAllocatableKb int64
	Labels           map[string]string
	Annotations      map[string]string
}

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
	PodUnknown   PodPhase = "Unknown"
	// Internal phase used for removed pods.
	PodDeleted PodPhase = "Deleted"
)

type PodIdentifier struct {
	Name      string
	Namespace string
}

func (this *PodIdentifier) UniqueName() string {
	return this.Namespace + "/" + this.Name
}

type Pod struct {
	Identifier   PodIdentifier
	State        PodPhase
	CpuRequest   int64
	MemRequestKb int64
	Labels       map[string]string
	Annotations  map[string]string
	NodeSelector map[string]string
}

type NodeWatcher struct {
	//ID string
	clientset     kubernetes.Interface
	nodeWorkQueue workqueue.DelayingInterface
	controller    cache.Controller
	fc            firmament.FirmamentSchedulerClient
}

type PodWatcher struct {
	//ID string
	clientset    kubernetes.Interface
	podWorkQueue workqueue.DelayingInterface
	controller   cache.Controller
	fc           firmament.FirmamentSchedulerClient
}
