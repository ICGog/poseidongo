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
	"time"

	//	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/ICGog/poseidongo/pkg/k8sclient"
)

var (
	firmamentAddress string
	kubeConfig       string
)

func init() {
	flag.StringVar(&firmamentAddress, "firmamentAddress", "127.0.0.1:9090", "Firmament scheduler address")
	flag.StringVar(&kubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
}

func main() {
	flag.Parse()

	// firmConfig := &firmament.firmamentConfig{
	// 	address: firmamentAddress,
	// }
	// fc, err := firmClient.New(firmConfig)
	// if err != nil {
	// 	return
	// }
	// k8sConfig = &k8sclient.k8sConfig{
	// 	address: k8sAddress,
	// 	QPS:     1000,
	// 	burst:   1000,
	// }
	k8sClient, err := k8sclient.New(kubeConfig)
	if err != nil {
		panic(err.Error())
	}
	for {
		pods, err := k8sclient.Core().Pods("").List(v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		time.Sleep(10 * time.Second)
	}
	// for {
	// 	select {
	// 	case <-k8sClient.nodeCh:
	// 		fmt.Println("New Node")
	// 	case <-k8sClient.podCh:
	// 		ftm.Println("New pod")
	// 	}
	// }

}
