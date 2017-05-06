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
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const nodeBufferSize = 1000
const podBufferSize = 1000
const bytesToKb = 1024

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

func StartNodeWatcher(clientset *kubernetes.Clientset) chan *Node {
	nodeCh := make(chan *Node, nodeBufferSize)
	nodeListWatcher := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())
	_, nodeInformer := cache.NewInformer(
		nodeListWatcher,
		&v1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(nodeObj interface{}) {
				node := nodeObj.(*v1.Node)
				if node.Spec.Unschedulable {
					return
				}
				isReady := false
				isOutOfDisk := false
				for _, cond := range node.Status.Conditions {
					switch cond.Type {
					case "OutOfDisk":
						isOutOfDisk = cond.Status == "True"
					case "Ready":
						isReady = cond.Status == "True"
					}
				}
				cpuCapQuantity := node.Status.Capacity["cpu"]
				cpuCap, _ := cpuCapQuantity.AsInt64()
				cpuAllocQuantity := node.Status.Allocatable["cpu"]
				cpuAlloc, _ := cpuAllocQuantity.AsInt64()
				memCapQuantity := node.Status.Capacity["memory"]
				memCap, _ := memCapQuantity.AsInt64()
				memAllocQuantity := node.Status.Allocatable["memory"]
				memAlloc, _ := memAllocQuantity.AsInt64()
				nodeCh <- &Node{
					Hostname:         node.Name,
					IsReady:          isReady,
					IsOutOfDisk:      isOutOfDisk,
					CpuCapacity:      cpuCap,
					CpuAllocatable:   cpuAlloc,
					MemCapacityKb:    memCap / bytesToKb,
					MemAllocatableKb: memAlloc / bytesToKb,
					Labels:           node.Labels,
					Annotations:      node.Annotations,
				}

			},
			UpdateFunc: func(oldNodeObj, newNodeObj interface{}) {
				oldNode := oldNodeObj.(*v1.Node)
				fmt.Printf("%+v\n", oldNode)
				newNode := newNodeObj.(*v1.Node)
				fmt.Printf("%+v\n", newNode)
			},
			DeleteFunc: func(nodeObj interface{}) {
				node := nodeObj.(*v1.Node)
				fmt.Printf("%+v\n", node)
			},
		},
	)
	stopCh := make(chan struct{})
	go nodeInformer.Run(stopCh)
	return nodeCh
}

func StartPodWatcher(clientset *kubernetes.Clientset) chan *Pod {
	podCh := make(chan *Pod, podBufferSize)
	podListWatcher := cache.NewListWatchFromClient(clientset.Core().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
	_, podInformer := cache.NewInformer(
		podListWatcher,
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(podObj interface{}) {
				pod := podObj.(*v1.Pod)
				cpuReq := int64(0)
				memReq := int64(0)
				for _, container := range pod.Spec.Containers {
					request := container.Resources.Requests
					cpuReqQuantity := request["cpu"]
					cpuReqCont, _ := cpuReqQuantity.AsInt64()
					cpuReq += cpuReqCont
					memReqQuantity := request["memory"]
					memReqCont, _ := memReqQuantity.AsInt64()
					memReq += memReqCont
				}
				podPhase := PodPhase("Unknown")
				switch pod.Status.Phase {
				case "Pending":
					podPhase = "Pending"
				case "Running":
					podPhase = "Running"
				case "Succeeded":
					podPhase = "Succeeded"
				case "Failed":
					podPhase = "Failed"
				}
				podCh <- &Pod{
					Name:         pod.Name,
					State:        podPhase,
					CpuRequest:   cpuReq,
					MemRequestKb: memReq / bytesToKb,
					Labels:       pod.Labels,
					Annotations:  pod.Annotations,
					NodeSelector: pod.Spec.NodeSelector,
				}

			},
			UpdateFunc: func(oldPodObj, newPodObj interface{}) {},
			DeleteFunc: func(podObj interface{}) {},
		},
	)
	stopCh := make(chan struct{})
	go podInformer.Run(stopCh)
	return podCh
}

func BindPodToNode() {
	// TODO(ionel): Implement!
}

func DeletePod() {
	// TODO(ionel): Implement!
}

func New(kubeConfig string) (chan *Node, chan *Pod) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	nodeCh := StartNodeWatcher(clientset)
	podCh := StartPodWatcher(clientset)
	return nodeCh, podCh
}
