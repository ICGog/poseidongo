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

package stats

import (
	"github.com/ICGog/poseidongo/pkg/firmament"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"io"
	"net"
)

type poseidonStatsServer struct {
	firmamentClient firmament.FirmamentSchedulerClient
}

func convertPodStatToTaskStat(podStats *PodStats) *firmament.TaskStats {
	return &firmament.TaskStats{
		CpuLimit:            podStats.GetCpuLimit(),
		CpuRequest:          podStats.GetCpuRequest(),
		CpuUsage:            podStats.GetCpuUsage(),
		MemLimit:            podStats.GetMemLimit(),
		MemRequest:          podStats.GetMemRequest(),
		MemUsage:            podStats.GetMemUsage(),
		MemWorkingSet:       podStats.GetMemWorkingSet(),
		MemPageFaults:       podStats.GetMemPageFaults(),
		MemPageFaultsRate:   podStats.GetMemPageFaultsRate(),
		MajorPageFaults:     podStats.GetMajorPageFaults(),
		MajorPageFaultsRate: podStats.GetMajorPageFaultsRate(),
		NetRx:               podStats.GetNetRx(),
		NetRxErrors:         podStats.GetNetRxErrors(),
		NetRxErrorsRate:     podStats.GetNetRxErrorsRate(),
		NetRxRate:           podStats.GetNetRxRate(),
		NetTx:               podStats.GetNetTx(),
		NetTxErrors:         podStats.GetNetTxErrors(),
		NetTxErrorsRate:     podStats.GetNetTxErrorsRate(),
		NetTxRate:           podStats.GetNetTxRate(),
	}
}

func convertNodeStatToResourceStat(nodeStats *NodeStats) *firmament.ResourceStats {
	return &firmament.ResourceStats{
		CpuAllocatable: nodeStats.GetCpuAllocatable(),
		CpuCapacity:    nodeStats.GetCpuCapacity(),
		CpuReservation: nodeStats.GetCpuReservation(),
		CpuUtilization: nodeStats.GetCpuUtilization(),
		MemAllocatable: nodeStats.GetMemAllocatable(),
		MemCapacity:    nodeStats.GetMemCapacity(),
		MemReservation: nodeStats.GetMemReservation(),
		MemUtilization: nodeStats.GetMemUtilization(),
	}

}

func (s *poseidonStatsServer) ReceiveNodeStats(stream PoseidonStats_ReceiveNodeStatsServer) error {
	for {
		nodeStat, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		resourceStat := convertNodeStatToResourceStat(nodeStat)

		//TODO:(shiv)check the NodetoResource Map and get the resourceid and set the ResourceUid below

		//resourceStat.ResourceUid =

		firmament.AddNodeStats(s.firmamentClient, resourceStat)

	}
}

func (s *poseidonStatsServer) ReceivePodStats(stream PoseidonStats_ReceivePodStatsServer) error {
	for {
		podStat, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		taskStat := convertPodStatToTaskStat(podStat)
		//TODO:(shiv) Check the PodtoTask mao and get the taskUid and set the TaskUis below
		//taskStat.TaskUid

		firmament.AddTaskStats(s.firmamentClient, taskStat)

	}
}

func NewposeidonStatsServer(firmamentAddress string) *poseidonStatsServer {
	newfirmamentClient, _, err := firmament.New(firmamentAddress)
	if err != nil {
		glog.Fatalln("Unable to initialze firmamentClient in stats", err)

	}
	return &poseidonStatsServer{firmamentClient: newfirmamentClient}
}

// StartgRPCStatsServer will run in a separate goroutine
//
func StartgRPCStatsServer(ip, port, firmamentAddress string) {
	lis, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	RegisterPoseidonStatsServer(grpcServer, NewposeidonStatsServer(firmamentAddress))
	grpcServer.Serve(lis)
}
