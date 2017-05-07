package k8sclient

import (
	"github.com/shivramsrivastava/poseidongo/pkg/firmament"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const nodeBufferSize = 1000
const podBufferSize = 1000
const bytesToKb = 1024

var podToTD map[string]*firmament.TaskDescriptor
var jobMap map[string]*firmament.JobDescriptor

type Node struct {
	Hostname         string
	IsReady          bool
	IsOutOfDisk      bool
	CpuCapacity      int64
	CpuAllocatable   int64
	MemCapacityKb    int64
	MemAllocatableKb int64
	Labels           map[string]string
	Annotations      map[string]string
}

type PodPhase string

const (
	PodPending   PodPhase = "Pending"
	PodRunning   PodPhase = "Running"
	PodSucceeded PodPhase = "Succeeded"
	PodFailed    PodPhase = "Failed"
	PodUnknown   PodPhase = "Unknown"
)

type Pod struct {
	Name         string
	State        PodPhase
	CpuRequest   int64
	MemRequestKb int64
	Labels       map[string]string
	Annotations  map[string]string
	NodeSelector map[string]string
}

type NodeWatcher struct {
	//ID string
	clientset     kubernetes.Interface
	nodeWorkQueue workqueue.DelayingInterface
	controller    cache.Controller
}

type PodWatcher struct {
	//ID string
	clientset    kubernetes.Interface
	podWorkQueue workqueue.DelayingInterface
	controller   cache.Controller
}
