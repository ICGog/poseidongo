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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"poseidongo/pkg/firmament"
)

const bytesToKb = 1024

var podToTD map[string]*firmament.TaskDescriptor
var jobMap map[string]*firmament.JobDescriptor

type Node struct {
	Hostname         string
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
)

type Pod struct {
	Name         string
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
}

type PodWatcher struct {
	//ID string
	clientset    kubernetes.Interface
	podWorkQueue workqueue.DelayingInterface
	controller   cache.Controller
}
