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
	jobMap           map[string]*firmament.JobDescriptor
	podToTD          map[string]*firmament.TaskDescriptor
)

func init() {
	flag.StringVar(&firmamentAddress, "firmamentAddress", "127.0.0.1:9090", "Firmament scheduler address")
	flag.StringVar(&kubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
	flag.Parse()
}

func generateResourceID() string {
	// TODO(ionel): GenerateResourceID
	return "test"
}

func createResourceTopologyForNode(node *k8sclient.Node) *firmament.ResourceTopologyNodeDescriptor {
	resUuid := generateResourceID()
	rtnd := &firmament.ResourceTopologyNodeDescriptor{
		ResourceDesc: &firmament.ResourceDescriptor{
			Uuid:         resUuid,
			Type:         firmament.ResourceDescriptor_RESOURCE_MACHINE,
			State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
			FriendlyName: node.Hostname,
			ResourceCapacity: &firmament.ResourceVector{
				RamCap:   node.MemCapacityKb,
				CpuCores: node.CpuCapacity,
			},
		},
	}
	// TODO(ionel) Add annotations.
	// Add labels.
	for label, value := range node.Labels {
		rtnd.ResourceDesc.Labels = append(rtnd.ResourceDesc.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}

	// TODO(ionel): In the future, we want to get real node topology rather
	// than manually connecting PU RDs to the machine RD.
	for num_pu := int64(0); num_pu < node.CpuCapacity; num_pu++ {
		puResUuid := generateResourceID()
		puRtnd := &firmament.ResourceTopologyNodeDescriptor{
			ResourceDesc: &firmament.ResourceDescriptor{
				Uuid:         puResUuid,
				Type:         firmament.ResourceDescriptor_RESOURCE_PU,
				State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
				FriendlyName: node.Hostname + "_pu" + strconv.FormatInt(num_pu, 10),
				Labels:       rtnd.ResourceDesc.Labels,
			},
			ParentId: resUuid,
		}
		rtnd.Children = append(rtnd.Children, puRtnd)
	}
	return rtnd
}

func generateJobID() string {
	// TODO(ionel): Implement!
	return "test"
}

func createNewJob(controllerId string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  generateJobID(),
		Name:  controllerId,
		State: firmament.JobDescriptor_CREATED,
	}
	return jobDesc
}

func generateRootTaskID() uint64 {
	// TODO(ionel): Implement!
	return 0
}

func addTaskToJob(pod *k8sclient.Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Uid:   generateRootTaskID(),
		Name:  pod.Name,
		State: firmament.TaskDescriptor_CREATED,
		JobId: jd.Uuid,
		ResourceRequest: &firmament.ResourceVector{
			CpuCores: pod.CpuRequest,
			RamCap:   pod.MemRequestKb,
		},
		// TODO(ionel): Populate LabelSelector.
	}
	// Add labels.
	for label, value := range pod.Labels {
		task.Labels = append(task.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}
	if jd.RootTask == nil {
		jd.RootTask = task
	} else {
		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
	}
	return task
}

func main() {
	// fc, err := firmament.New(firmamentAddress)
	// if err != nil {
	// 	return
	// }

	nodeCh, podCh := k8sclient.New(kubeConfig)
	for {
		select {
		case node := <-nodeCh:
			rtnd := createResourceTopologyForNode(node)
			//firmament.NodeAdded(fc, rtnd)
			fmt.Println("New node")

			//firmament.NodeFailed(fc)
			//firmament.NodeRemoved(fc)
			//firmament.AddNodeStats(fc)
		case pod := <-podCh:
			switch pod.State {
			case "Pending":
				jd, ok := jobMap[generateJobID()]
				if !ok {
					jd := createNewJob()
				}
				td := addTaskToJob(pod, jd)
				//firmament.TaskSubmitted(fc, td)
			case "Succeeded":
				td, ok := podToTD[pod.Name]
				//firmament.TaskCompleted(fc, td)
			case "Failed":
				td, ok := podToTD[pod.Name]
				//firmament.TaskFailed(fc, td)
			case "Running":
				//firmament.AddTaskStats(fc)
			case "Unknown":
				//firmament.TaskRemoved(fc)
			}

		}
		//firmament.Schedule(fc)
	}
}
