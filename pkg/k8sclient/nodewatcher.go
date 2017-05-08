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
	"strconv"
	"time"

	//"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/pkg/api"
	"github.com/golang/glog"
	"github.com/shivramsrivastava/poseidongo/pkg/firmament"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

func NewNodeWatcher(client kubernetes.Interface) *NodeWatcher {

	glog.Info("Starting NodeWatcher")

	nodewatcher := &NodeWatcher{clientset: client}
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				return client.CoreV1().Nodes().List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				return client.CoreV1().Nodes().Watch(alo)
			},
		},
		&v1.Node{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				glog.Info(" Add Event on Nodewatcher called ")
				nodewatcher.enqueueNodes(obj)
			},
			UpdateFunc: func(old, new interface{}) {
				glog.Info(" Update Event on Nodewatcher called ")
				nodewatcher.enqueueNodes(new)
			},
			DeleteFunc: func(obj interface{}) {
				glog.Info(" Delete Event on Nodewatcher called ")
				nodewatcher.enqueueNodes(obj)
			},
		},
	)

	nodewatcher.controller = controller
	nodewatcher.nodeWorkQueue = workqueue.NewNamedDelayingQueue("nodeQueue")

	return nodewatcher

}

func (this *NodeWatcher) enqueueNodes(obj interface{}) {

	glog.Info("enqueueNodes function called")
	node := obj.(*v1.Node)
	//chec if node already in the map
	if node.Spec.Unschedulable {
		glog.Info("enqueueNodes: Received an Unschedulable node", node.Name)
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
	newNode := &Node{
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
	this.nodeWorkQueue.Add(newNode)
	glog.Info("enqueueNodes: Added a new node", node.Name)
}

func (this *NodeWatcher) Run(stopCh <-chan struct{}, workers int) {
	//defer runtime.HandleCrash()
	defer this.nodeWorkQueue.ShutDown()

	glog.Info("Run method of nodeWatcher called")
	defer glog.Info("Shutting down NodeWatcher")

	go this.controller.Run(stopCh) //this call will popultate the events

	if !cache.WaitForCacheSync(stopCh, this.controller.HasSynced) {

		glog.Info("Waiting in the run method for 3 sec")
		time.Sleep(time.Second * 3)
		return
	}

	glog.Info("Starting the node watcher worker methods")

	for i := 0; i < workers; i++ {

		go wait.Until(this.nodeWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (this *NodeWatcher) nodeWorker() {
	for {
		func() {
			glog.Info("nodeWorker func called")
			glog.Info("Blocking in get func of nodeWorker func")

			key, quit := this.nodeWorkQueue.Get()
			if quit {
				return
			}
			node := key.(*Node)
			glog.Info("node data from the queue", node)
			rtnd := this.createResourceTopologyForNode(node)
			glog.Info("ResourceTopologyNodeDescriptor ", rtnd)
			//firmament.NodeAdded(fc, rtnd)
			//firmament.NodeFailed(fc)
			//firmament.NodeRemoved(fc)
			//firmament.AddNodeStats(fc)
			defer this.nodeWorkQueue.Done(key)
		}()
	}
}

func (this *NodeWatcher) createResourceTopologyForNode(node *Node) *firmament.ResourceTopologyNodeDescriptor {
	resUuid := this.generateResourceID()
	rtnd := &firmament.ResourceTopologyNodeDescriptor{
		ResourceDesc: &firmament.ResourceDescriptor{
			Uuid:         resUuid,
			Type:         firmament.ResourceDescriptor_RESOURCE_MACHINE,
			State:        firmament.ResourceDescriptor_RESOURCE_IDLE,
			FriendlyName: node.Hostname,
			ResourceCapacity: &firmament.ResourceVector{
				RamCap:   uint64(node.MemCapacityKb),
				CpuCores: float32(node.CpuCapacity),
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
		puResUuid := this.generateResourceID()
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

func (this *NodeWatcher) generateResourceID() string {
	// TODO(ionel): GenerateResourceID
	return "test"
}