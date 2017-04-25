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
	"k8s.io/kubernetes/pkg/api"
	k8sClient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/watch"
)

type k8sConfig struct {
	address string
	QPS     int
	burst   int
}

func New(k8sConfig config) (*Client, error) {
	restClientConfig := &restClient.Config{
		Host:  fmt.Sprintf("http://%s", config.address),
		QPS:   Config.QPS,
		Burst: Config.burst,
	}
	c, err := k8sClient.New(restClientConfig)
	if err != nil {
		return nil, err
	}
}
