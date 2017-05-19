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
	"reflect"
	"sync"
	"time"

	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"
)

func NewPodWatcher(kubeVerMajor, kubeVerMinor int, schedulerName string, client kubernetes.Interface, fc firmament.FirmamentSchedulerClient) *PodWatcher {
	glog.Info("Starting PodWatcher...")
	PodsCond = sync.NewCond(&sync.Mutex{})
	PodToTD = make(map[PodIdentifier]*firmament.TaskDescriptor)
	TaskIDToPod = make(map[uint64]PodIdentifier)
	jobIDToJD = make(map[string]*firmament.JobDescriptor)
	jobNumTasksToRemove = make(map[string]int)
	podWatcher := &PodWatcher{
		clientset: client,
		fc:        fc,
	}
	schedulerSelector := fields.Everything()
	podSelector := labels.Everything()
	if kubeVerMajor >= 1 && kubeVerMinor >= 6 {
		// schedulerName is only available in Kubernetes >= 1.6.
		schedulerSelector = fields.ParseSelectorOrDie("spec.schedulerName==" + schedulerName)
	} else {
		var err error
		podSelector, err = labels.Parse("scheduler in (" + schedulerName + ")")
		if err != nil {
			glog.Fatal("Failed to parse scheduler label selector")
		}
	}
	_, controller := cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(alo metav1.ListOptions) (runtime.Object, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").List(alo)
			},
			WatchFunc: func(alo metav1.ListOptions) (watch.Interface, error) {
				alo.FieldSelector = schedulerSelector.String()
				alo.LabelSelector = podSelector.String()
				return client.CoreV1().Pods("").Watch(alo)
			},
		},
		&v1.Pod{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Errorf("AddFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodAddition(key, obj)
			},
			UpdateFunc: func(old, new interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(new)
				if err != nil {
					glog.Errorf("UpdateFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodUpdate(key, old, new)
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err != nil {
					glog.Errorf("DeleteFunc: error getting key %v", err)
				}
				podWatcher.enqueuePodDeletion(key, obj)
			},
		},
	)
	podWatcher.controller = controller
	podWatcher.podWorkQueue = NewKeyedQueue()
	return podWatcher
}

func (this *PodWatcher) getCpuMemRequest(pod *v1.Pod) (int64, int64) {
	cpuReq := int64(0)
	memReq := int64(0)
	for _, container := range pod.Spec.Containers {
		request := container.Resources.Requests
		cpuReqQuantity := request["cpu"]
		cpuReq += cpuReqQuantity.MilliValue()
		memReqQuantity := request["memory"]
		memReqCont, _ := memReqQuantity.AsInt64()
		memReq += memReqCont
	}
	return cpuReq, memReq
}

func (this *PodWatcher) parsePod(pod *v1.Pod) *Pod {
	cpuReq, memReq := this.getCpuMemRequest(pod)
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
	return &Pod{
		Identifier: PodIdentifier{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
		State:        podPhase,
		CpuRequest:   cpuReq,
		MemRequestKb: memReq / bytesToKb,
		Labels:       pod.Labels,
		Annotations:  pod.Annotations,
		NodeSelector: pod.Spec.NodeSelector,
	}
}

func (this *PodWatcher) enqueuePodAddition(key interface{}, obj interface{}) {
	pod := obj.(*v1.Pod)
	addedPod := this.parsePod(pod)
	this.podWorkQueue.Add(key, addedPod)
	glog.Info("enqueuePodAddition: Added pod ", addedPod.Identifier)
}

func (this *PodWatcher) enqueuePodDeletion(key interface{}, obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.DeletionTimestamp != nil {
		// Only delete pods if they have a DeletionTimestamp.
		deletedPod := &Pod{
			Identifier: PodIdentifier{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			State: PodDeleted,
		}
		this.podWorkQueue.Add(key, deletedPod)
		glog.Info("enqueuePodDeletion: Added pod ", deletedPod.Identifier)
	}
}

func (this *PodWatcher) enqueuePodUpdate(key, oldObj, newObj interface{}) {
	oldPod := oldObj.(*v1.Pod)
	newPod := newObj.(*v1.Pod)
	if oldPod.Status.Phase != newPod.Status.Phase {
		// TODO(ionel): This code assumes that if other fields changed as well then Firmament will automatically update them upon state transition. This is currently not true.
		updatedPod := this.parsePod(newPod)
		this.podWorkQueue.Add(key, updatedPod)
		glog.Infof("enqueuePodUpdate: Updated pod state change %v %s", updatedPod.Identifier, updatedPod.State)
		return
	}
	oldCpuReq, oldMemReq := this.getCpuMemRequest(oldPod)
	newCpuReq, newMemReq := this.getCpuMemRequest(newPod)
	if oldCpuReq != newCpuReq || oldMemReq != newMemReq ||
		!reflect.DeepEqual(oldPod.Labels, newPod.Labels) ||
		!reflect.DeepEqual(oldPod.Annotations, newPod.Annotations) ||
		!reflect.DeepEqual(oldPod.Spec.NodeSelector, newPod.Spec.NodeSelector) {
		updatedPod := this.parsePod(newPod)
		this.podWorkQueue.Add(key, updatedPod)
		glog.Info("enqueuePodUpdate: Updated pod ", updatedPod.Identifier)
		return
	}
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
			key, items, quit := this.podWorkQueue.Get()
			if quit {
				return
			}
			for _, item := range items {
				pod := item.(*Pod)
				switch pod.State {
				case PodPending:
					glog.V(2).Info("PodPending ", pod.Identifier)
					PodsCond.L.Lock()
					// TODO(ionel): We generate a job per pod. Add a field to the Pod struct that uniquely identifies jobs/daemon sets and use that one instead to group pods into Firmament jobs.
					jobId := this.generateJobID(pod.Identifier.Name)
					jd, ok := jobIDToJD[jobId]
					if !ok {
						jd = this.createNewJob(pod.Identifier.Name)
						jobIDToJD[jobId] = jd
						jobNumTasksToRemove[jobId] = 0
					}
					td := this.addTaskToJob(pod, jd)
					jobNumTasksToRemove[jobId]++
					PodToTD[pod.Identifier] = td
					TaskIDToPod[td.GetUid()] = pod.Identifier
					taskDescription := &firmament.TaskDescription{
						TaskDescriptor: td,
						JobDescriptor:  jd,
					}
					PodsCond.L.Unlock()
					firmament.TaskSubmitted(this.fc, taskDescription)
				case PodSucceeded:
					glog.V(2).Info("PodSucceeded ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %v does not exist", pod.Identifier)
					}
					firmament.TaskCompleted(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				case PodDeleted:
					glog.V(2).Info("PodDeleted ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %s does not exist", pod.Identifier)
					}
					firmament.TaskRemoved(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
					PodsCond.L.Lock()
					delete(PodToTD, pod.Identifier)
					delete(TaskIDToPod, td.GetUid())
					// TODO(ionel): Should we delete the task from JD's spawned field?
					jobId := this.generateJobID(pod.Identifier.Name)
					jobNumTasksToRemove[jobId]--
					if jobNumTasksToRemove[jobId] == 0 {
						// Clean state because the job doesn't have any tasks left.
						delete(jobNumTasksToRemove, jobId)
						delete(jobIDToJD, jobId)
					}
					PodsCond.L.Unlock()
				case PodFailed:
					glog.V(2).Info("PodFailed ", pod.Identifier)
					PodsCond.L.Lock()
					td, ok := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !ok {
						glog.Fatalf("Pod %s does not exist", pod.Identifier)
					}
					firmament.TaskFailed(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				case PodRunning:
					glog.V(2).Info("PodRunning ", pod.Identifier)
					// We don't have to do anything.
				case PodUnknown:
					glog.Errorf("Pod %s in unknown state", pod.Identifier)
					// TODO(ionel): Handle Unknown case.
				case PodUpdated:
					glog.V(2).Info("PodUpdated ", pod.Identifier)
					PodsCond.L.Lock()
					jobId := this.generateJobID(pod.Identifier.Name)
					jd, okJob := jobIDToJD[jobId]
					td, okPod := PodToTD[pod.Identifier]
					PodsCond.L.Unlock()
					if !okJob {
						glog.Fatalf("Pod's %v job does not exist", pod.Identifier)
					}
					if !okPod {
						glog.Fatalf("Pod %v does not exist", pod.Identifier)
					}
					this.updateTask(pod, td)
					taskDescription := &firmament.TaskDescription{
						TaskDescriptor: td,
						JobDescriptor:  jd,
					}
					firmament.TaskUpdated(this.fc, taskDescription)
				default:
					glog.Fatalf("Pod %v in unexpected state %v", pod.Identifier, pod.State)
				}
			}
			defer this.podWorkQueue.Done(key)
		}()
	}
}

func (this *PodWatcher) createNewJob(jobName string) *firmament.JobDescriptor {
	jobDesc := &firmament.JobDescriptor{
		Uuid:  this.generateJobID(jobName),
		Name:  jobName,
		State: firmament.JobDescriptor_CREATED,
	}
	return jobDesc
}

func (this *PodWatcher) updateTask(pod *Pod, td *firmament.TaskDescriptor) {
	// TODO(ionel): Update LabelSelector!
	td.ResourceRequest.CpuCores = float32(pod.CpuRequest)
	td.ResourceRequest.RamCap = uint64(pod.MemRequestKb)
	// Update labels.
	td.Labels = nil
	for label, value := range pod.Labels {
		td.Labels = append(td.Labels,
			&firmament.Label{
				Key:   label,
				Value: value,
			})
	}
}

func (this *PodWatcher) addTaskToJob(pod *Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
		Name:  pod.Identifier.UniqueName(),
		State: firmament.TaskDescriptor_CREATED,
		JobId: jd.Uuid,
		ResourceRequest: &firmament.ResourceVector{
			// TODO(ionel): Update types so no cast is required.
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
		task.Uid = this.generateTaskID(jd.Name, 0)
		jd.RootTask = task
	} else {
		task.Uid = this.generateTaskID(jd.Name, len(jd.RootTask.Spawned)+1)
		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
	}
	return task
}

func (this *PodWatcher) generateJobID(seed string) string {
	return GenerateUUID(seed)
}

func (this *PodWatcher) generateTaskID(jdUid string, taskNum int) uint64 {
	return HashCombine(jdUid, taskNum)
}
