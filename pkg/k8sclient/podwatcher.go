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
	"time"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"poseidongo/pkg/firmament"
)

func NewPodWatcher(client kubernetes.Interface) *PodWatcher {
	glog.Info("Starting PodWatcher...")
	podWatcher := &PodWatcher{clientset: client}
	podStatuSelector := fields.ParseSelectorOrDie("spec.nodeName==")
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				alo.FieldSelector = podStatuSelector.String()
				return client.CoreV1().Pods("").List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				alo.FieldSelector = podStatuSelector.String()
				return client.CoreV1().Pods("").Watch(alo)
			},
		},
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				glog.Info(" Add Event on Podwatcher")
				podWatcher.enqueuePods(obj)
			},
			UpdateFunc: func(old, new interface{}) {
				glog.Info(" Update Event on Podwatcher")
				podWatcher.enqueuePods(new)
			},
			DeleteFunc: func(obj interface{}) {
				glog.Info(" Delete Event on Podwatcher")
				podWatcher.enqueuePods(obj)
			},
		},
	)
	podWatcher.controller = controller
	podWatcher.podWorkQueue = workqueue.NewNamedDelayingQueue("PodQueue")
	return podWatcher
}

func (this *PodWatcher) enqueuePods(obj interface{}) {
	glog.Info("enqueuePods function called")
	pod := obj.(*v1.Pod)
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
	newpod := &Pod{
		Name:         pod.Name,
		State:        podPhase,
		CpuRequest:   cpuReq,
		MemRequestKb: memReq / bytesToKb,
		Labels:       pod.Labels,
		Annotations:  pod.Annotations,
		NodeSelector: pod.Spec.NodeSelector,
	}
	this.podWorkQueue.Add(newpod)
	glog.Info("enqueuePods: Added a new Pod", pod.Name)
}

func (this *PodWatcher) Run(stopCh <-chan struct{}, nWorkers int) {
	defer utilruntime.HandleCrash()

	// The workers can stop when we are done.
	defer this.podWorkQueue.ShutDown()
	defer glog.Info("Shutting down PodWatcher")
	glog.Info("Getting pod updates...")

	go this.controller.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, this.controller.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	glog.Info("Starting pod watching workers")
	for i := 0; i < nWorkers; i++ {
		go wait.Until(this.podWorker, time.Second, stopCh)
	}

	<-stopCh
	glog.Info("Stopping pod watcher")
}

func (this *PodWatcher) podWorker() {
	for {
		func() {
			key, quit := this.podWorkQueue.Get()
			if quit {
				return
			}
			pod := key.(*Pod)
			switch pod.State {
			case "Pending":
				var jd *firmament.JobDescriptor
				var ok bool
				jd, ok = jobMap[this.generateJobID()]
				if !ok {
					jd = this.createNewJob("")
				}
				glog.Info("Firmament Jobdescriptor ", jd)
				//td := this.addTaskToJob(pod, jd)
				//firmament.TaskSubmitted(fc, td)
			case "Succeeded":
				//td, ok := podToTD[pod.Name]
				//firmament.TaskCompleted(fc, td)
			case "Failed":
				//td, ok := podToTD[pod.Name]
				//firmament.TaskFailed(fc, td)
			case "Running":
				//firmament.AddTaskStats(fc)
			case "Unknown":
				//firmament.TaskRemoved(fc)
			}
			glog.Info("Pod data received from the queue", pod)
			defer this.podWorkQueue.Done(key)
		}()
	}
}

func (this *PodWatcher) createNewJob(controllerId string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  this.generateJobID(),
		Name:  controllerId,
		State: firmament.JobDescriptor_CREATED,
	}
	return jobDesc
}

func (this *PodWatcher) addTaskToJob(pod *Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Uid:   this.generateRootTaskID(jd),
		Name:  pod.Name,
		State: firmament.TaskDescriptor_CREATED,
		JobId: jd.Uuid,
		ResourceRequest: &firmament.ResourceVector{
			CpuCores: float32(pod.CpuRequest),
			RamCap:   uint64(pod.MemRequestKb),
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
func (this *PodWatcher) generateJobID() string {
	// TODO(ionel): Implement!
	return GenerateUUID()
}

func (this *PodWatcher) generateRootTaskID(jd *firmament.JobDescriptor) uint64 {
	// TODO(shiv): No error handling

	return HashCombine(jd.GetName(), jd.GetRootTask().GetBinary())

}
