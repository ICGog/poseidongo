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
	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func BindPodToNode() {
	// TODO(ionel): Implement!
}

func DeletePod() {
	// TODO(ionel): Implement!
}

func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func New(kubeConfig string, schedulerName string) (chan *Node, chan *Pod) {
	config, err := GetClientConfig(kubeConfig)
	if err != nil {
		glog.Fatalf("Failed to load client config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create connection: %v", err)
	}
	glog.Info("k8s newclient called")
	stopCh := make(chan struct{})
	go NewPodWatcher(clientset, schedulerName).Run(stopCh, 10)
	go NewNodeWatcher(clientset).Run(stopCh, 10)

	//we block here
	<-stopCh
	return nil, nil
}
