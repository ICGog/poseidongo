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

type Node struct {
	ID string
}

type Pod struct {
	ID string
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
				nodeCh <- &Node{
					ID: node.Name,
				}
			},
			UpdateFunc: func(oldNodeObj, newNodeObj interface{}) {},
			DeleteFunc: func(nodeObj interface{}) {},
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
				podCh <- &Pod{
					ID: pod.Name,
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

func New(kubeConfig string) (int, int) {
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
	for {
		select {
		case <-nodeCh:
			fmt.Println("New node")
		case <-podCh:
			fmt.Println("New pod")
		}
	}

	return 1, 1
}
