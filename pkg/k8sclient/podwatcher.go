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

	"github.com/ICGog/poseidongo/pkg/firmament"
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
)

func NewPodWatcher(client kubernetes.Interface, firmamentAddress string) *PodWatcher {
	glog.Info("Starting PodWatcher...")
	fc, err := firmament.New(firmamentAddress)
	if err != nil {
		panic(err)
	}
	podWatcher := &PodWatcher{
		clientset: client,
		fc:        fc,
	}
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
				podWatcher.enqueuePodAddition(obj)
			},
			UpdateFunc: func(old, new interface{}) {
				podWatcher.enqueuePodUpdate(old, new)
			},
			DeleteFunc: func(obj interface{}) {
				podWatcher.enqueuePodDeletion(obj)
			},
		},
	)
	podWatcher.controller = controller
	podWatcher.podWorkQueue = workqueue.NewNamedDelayingQueue("PodQueue")
	return podWatcher
}

func (this *PodWatcher) enqueuePodAddition(obj interface{}) {
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
	addedPod := &Pod{
		Name:         pod.Namespace + "/" + pod.Name,
		State:        podPhase,
		CpuRequest:   cpuReq,
		MemRequestKb: memReq / bytesToKb,
		Labels:       pod.Labels,
		Annotations:  pod.Annotations,
		NodeSelector: pod.Spec.NodeSelector,
	}
	this.podWorkQueue.Add(addedPod)
	glog.Info("enqueuePodAddition: Added pod ", addedPod.Name)
}

func (this *PodWatcher) enqueuePodDeletion(obj interface{}) {
	pod := obj.(*v1.Pod)
	deletedPod := &Pod{
		Name:  pod.Namespace + "/" + pod.Name,
		State: PodDeleted,
	}
	this.podWorkQueue.Add(deletedPod)
	glog.Info("enqueuePodDeletion: Added pod ", deletedPod.Name)
}

func (this *PodWatcher) enqueuePodUpdate(oldObj, newObj interface{}) {
	// TODO(ionel): Implement!
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
			case PodPending:
				// TODO(ionel): We generate a job per pod. Add a field to the Pod struct that uniquely identifies jobs/daemon sets and use that one instead to group pods into Firmament jobs.
				jobId := this.generateJobID(pod.Name)
				jd, ok := jobIDToJD[jobId]
				if !ok {
					jd = this.createNewJob(pod.Name)
					jobIDToJD[jobId] = jd
					jobNumIncompleteTasks[jobId] = 0
				}
				td := this.addTaskToJob(pod, jd)
				jobNumIncompleteTasks[jobId]++
				podToTD[pod.Name] = td
				TaskIDToPod[td.GetUid()] = pod.Name
				taskDescription := &firmament.TaskDescription{
					TaskDescriptor: td,
					JobDescriptor:  jd,
				}
				firmament.TaskSubmitted(this.fc, taskDescription)
			case PodSucceeded:
				td, ok := podToTD[pod.Name]
				if !ok {
					glog.Fatalf("Pod %v does not exist", pod.Name)
				}
				firmament.TaskCompleted(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				delete(podToTD, pod.Name)
				delete(TaskIDToPod, td.GetUid())
				jobId := this.generateJobID(pod.Name)
				jobNumIncompleteTasks[jobId]--
				if jobNumIncompleteTasks[jobId] == 0 {
					// All tasks completed => Job is completed.
					delete(jobNumIncompleteTasks, jobId)
					delete(jobIDToJD, jobId)
				} else {
					// XXX(ionel): We currently leave the deleted task in the job's
					// root/spawned task list. We do not delete it in case we need
					// to consistently regenerate taskUids out of job name and tasks'
					// position within the Spawned list.
				}
			case PodDeleted:
				td, ok := podToTD[pod.Name]
				if !ok {
					glog.Fatalf("Pod %v does not exist", pod.Name)
				}
				firmament.TaskRemoved(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				delete(podToTD, pod.Name)
				delete(TaskIDToPod, td.GetUid())
				jobId := this.generateJobID(pod.Name)
				jobNumIncompleteTasks[jobId]--
				if jobNumIncompleteTasks[jobId] == 0 {
					// Clean state because the job doesn't have any tasks left.
					delete(jobNumIncompleteTasks, jobId)
					delete(jobIDToJD, jobId)
				} else {
					// XXX(ionel): We currently leave the deleted task in the job's
					// root/spawned task list. We do not delete it in case we need
					// to consistently regenerate taskUids out of job name and tasks'
					// position within the Spawned list.
				}
			case PodFailed:
				td, ok := podToTD[pod.Name]
				if !ok {
					glog.Fatalf("Pod %v does not exist", pod.Name)
				}
				firmament.TaskFailed(this.fc, &firmament.TaskUID{TaskUid: td.Uid})
				// TODO(ionel): We do not delete the task from podToTD and taskIDToPod in case the task may be rescheduled. Check how K8s restart policies work and decide what to do here.
				// TODO(ionel): Should we delete the task from JD's spawned field?
			case PodRunning:
				// We don't have to do anything.
			case PodUnknown:
				glog.Error("Pod %v in uknown state", pod.Name)
				// TODO(ionel): Handle Unknown case.
			default:
				glog.Fatalf("Pod %v in unexpected state %v", pod.Name, pod.State)
			}
			glog.Info("Pod data received from the queue", pod)
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

func (this *PodWatcher) addTaskToJob(pod *Pod, jd *firmament.JobDescriptor) *firmament.TaskDescriptor {
	task := &firmament.TaskDescriptor{
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
