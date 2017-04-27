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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type Pod struct {
	ID string
}

type Node struct {
	ID string
}

// type Client struct {
// 	k8sApi *k8sClient.Client
// 	nodeCh chan *Node
// 	podCh  chan *Pod
// }

// func StartNodeWatcher(clientset Clientset) (chan *Node, chan struct{}) {
// 	nodeCh := make(chan *Node, NodeBufferSize)

// _, nodeInformer := framework.NewInformer(
// 	cache.NewListWatchFromClient(c, "nodes", api.NamespaceAll,
// 		fields.ParseSelectorOrDie("")),
// 	&api.Node{},
// 	0,
// 	framework.ResourceEventHandlerFuncs{
// 		AddFunc: func(nodeObj interface{}) {
// 			node := nodeObj.(*api.Node)
// 			if node.Spec.Unschedulable {
// 				return
// 			}
// 			nodeCh <- &Node{
// 				ID: node.Name,
// 			}
// 		},
// 		UpdateFunc: func(oldNodeObj, newNodeObj interface{}) {},
// 		DeleteFunc: func(nodeObj interface{}) {},
// 	},
// )
//	stopCh := make(chan struct{})
//	go nodeInformer.Run(stopCh)
//	return nodeCh, stopCh
//}

// func StartPodWatcher() (chan *Pod, chan struct{}) {
// 	podCh := make(chan *Pod, PodBufferSize)

// 	podSelector := fields.ParseSelectorOrDie("spec.nodeName==")
// 	podInformer := framework.NewSharedInformer(
// 		&cache.ListWatch{
// 			ListFunc: func(options api.ListOptions) (runtime.Object, error) {
// 				options.FieldSelector = podSelector
// 				return c.Pods(api.NamespaceAll).List(options)
// 			},
// 			WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
// 				options.FieldSelector = podSelector
// 				return c.Pods(api.NamespaceAll).Watch(options)
// 			},
// 		},
// 		&api.Pod{},
// 		0,
// 	)
// 	podInformer.AddEventHandler(framework.ResourceEventHandlerFuncs{
// 		AddFunc: func(podObj interface{}) {
// 			pod := podObj.(*api.Pod)
// 			podeCh <- &Pod{
// 				ID: path.Join(pod.Namespace, pod.Name),
// 			}
// 		},
// 		UpdateFunc: func(oldPodObj, newPodObj interface{}) {},
// 		DeleteFunc: func(podObj interface{}) {},
// 	})
// 	stopCh := make(chan struct{})
// 	go podInformer.Run(stopCh)
// 	return podCh, stopCh
// }

func New(kubeConfig string) (int, int) {
	// nodeCh, stopNodeCh = StartNodeWatcher()
	// podCh, stopPodCh = StartPodWatcher()
	// return &Client{
	// 	k8sApi: c,
	// 	nodeCh: nodeCh,
	// 	podCh:  podCh,
	// }, nil
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := clientset.Core().Pods("").List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		time.Sleep(10 * time.Second)
	}

	return 1, 1
}
