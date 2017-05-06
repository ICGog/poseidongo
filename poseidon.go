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

package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/ICGog/poseidongo/pkg/k8sclient"
)

var (
	firmamentAddress string
	kubeConfig       string
)

func init() {
	flag.StringVar(&firmamentAddress, "firmamentAddress", "127.0.0.1:9090", "Firmament scheduler address")
	flag.StringVar(&kubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
	flag.Parse()
}

func GenerateResourceID() string {
	// TODO(ionel): GenerateResourceID
	return "test"
}

func createResourceTopologyForNode(hostname string, cpuCapacity int64, ramCapacity int64) *firmament.ResourceTopologyNodeDescriptor {
	resUuid := GenerateResourceID()
	rtnd := &firmament.ResourceTopologyNodeDescriptor{
		ResourceDesc: &firmament.ResourceDescriptor{
			Uuid:         resUuid,
			Type:         firmament.ResourceDescriptor_RESOURCE_MACHINE,
			State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
			FriendlyName: hostname,
			ResourceCapacity: &firmament.ResourceVector{
				RamCap:   ramCapacity,
				CpuCores: cpuCapacity,
			},
		},
	}
	// TODO(ionel): In the future, we want to get real node topology rather
	// than manually connecting PU RDs to the machine RD.
	for num_pu := int64(0); num_pu < cpuCapacity; num_pu++ {
		puResUuid := GenerateResourceID()
		puRtnd := &firmament.ResourceTopologyNodeDescriptor{
			ResourceDesc: &firmament.ResourceDescriptor{
				Uuid:         puResUuid,
				Type:         firmament.ResourceDescriptor_RESOURCE_PU,
				State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
				FriendlyName: hostname + "_pu" + strconv.FormatInt(num_pu, 10),
			},
			ParentId: resUuid,
		}
		rtnd.Children = append(rtnd.Children, puRtnd)
	}
	return rtnd
}

// func createNewJob(controllerId string) {
// 	// jobDesc := firmament.JobDescriptor{
// 	// 	Uuid:
// 	// 	Name: controllerId,
// 	// 	State: firmament.JobDescriptor_Created,
// 	// }
// }

// func addTaskToJob(jd *firmament.JobDescriptor) {
// 	task := &firmament.TaskDescriptor{
// 		//	Uid: ,
// 		Name:  name,
// 		State: firmament.TaskDescriptor_Created,
// 		JobID: jobUuid,
// 	}
// 	if jd.RootTask == nil {
// 		jd.RootTask = task
// 	} else {
// 		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
// 	}
// }

func main() {
	// fc, err := firmament.New(firmamentAddress)
	// if err != nil {
	// 	return
	// }

	// firmament.AddNodeStats(fc, &firmament.ResourceStats{})

	nodeCh, podCh := k8sclient.New(kubeConfig)
	for {
		select {
		case node := <-nodeCh:
			rtnd := createResourceTopologyForNode(node.Hostname, node.CpuCapacity, node.MemCapacityKb)
			//firmament.NodeAdded(fc, rtnd)
			fmt.Println("New node")
		case <-podCh:
			fmt.Println("New pod")
			//firmament.TaskSubmitted(fc, td)
		}
	}
}
