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

	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/ICGog/poseidongo/pkg/k8sclient"
)

var (
	firmamentAddress string
	kubeConfig       string
)

func init() {

	flag.StringVar(&firmamentAddress, "firmamentAddress", "127.0.0.1:9090", "Firmament scheduler address")
	flag.StringVar(&kubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
	flag.Parse()
}

// func createNewJob(controllerId string) {
// 	// jobDesc := firmament.JobDescriptor{
// 	// 	Uuid:
// 	// 	Name: controllerId,
// 	// 	State: firmament.JobDescriptor_Created,
// 	// }
// }

// func addTaskToJob(jd *firmament.JobDescriptor) {
// 	task := &firmament.TaskDescriptor{
// 		//	Uid: ,
// 		Name:  name,
// 		State: firmament.TaskDescriptor_Created,
// 		JobID: jobUuid,
// 	}
// 	if jd.RootTask == nil {
// 		jd.RootTask = task
// 	} else {
// 		jd.RootTask.Spawned = append(jd.RootTask.Spawned, task)
// 	}
// }

func main() {
	fc, err := firmament.New(firmamentAddress)
	if err != nil {
		return
	}

	firmament.AddNodeStats(fc, &firmament.ResourceStats{})

	k8sClient, erra := k8sclient.New(kubeConfig)
	fmt.Printf("%d %d", k8sClient, erra)
}
