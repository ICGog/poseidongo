// Code generated by protoc-gen-go.
// source: coco_interference_scores.proto
// DO NOT EDIT!

/*
Package firmament is a generated protocol buffer package.

It is generated from these files:
	coco_interference_scores.proto
	firmament_scheduler.proto
	job_desc.proto
	label.proto
	label_selector.proto
	machine_perf_statistics_sample.proto
	reference_desc.proto
	resource_desc.proto
	resource_topology_node_desc.proto
	resource_vector.proto
	scheduling_delta.proto
	task_desc.proto
	task_final_report.proto
	task_perf_statistics_sample.proto
	task_stats.proto
	whare_map_stats.proto

It has these top-level messages:
	CoCoInterferenceScores
	ScheduleRequest
	SchedulingDeltas
	TaskCompletedResponse
	TaskDescription
	TaskSubmittedResponse
	TaskRemovedResponse
	TaskFailedResponse
	NodeAddedResponse
	NodeRemovedResponse
	NodeFailedResponse
	NodeUpdatedResponse
	TaskStatsResponse
	ResourceStatsResponse
	TaskUID
	ResourceUID
	ResourceStats
	JobDescriptor
	Label
	LabelSelector
	MachinePerfStatisticsSample
	CpuUsage
	ReferenceDescriptor
	ResourceDescriptor
	ResourceTopologyNodeDescriptor
	ResourceVector
	SchedulingDelta
	TaskDescriptor
	TaskFinalReport
	TaskPerfStatisticsSample
	TaskStats
	WhareMapStats
*/
package firmament

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CoCoInterferenceScores struct {
	DevilPenalty  uint32 `protobuf:"varint,1,opt,name=devil_penalty,json=devilPenalty" json:"devil_penalty,omitempty"`
	RabbitPenalty uint32 `protobuf:"varint,2,opt,name=rabbit_penalty,json=rabbitPenalty" json:"rabbit_penalty,omitempty"`
	SheepPenalty  uint32 `protobuf:"varint,3,opt,name=sheep_penalty,json=sheepPenalty" json:"sheep_penalty,omitempty"`
	TurtlePenalty uint32 `protobuf:"varint,4,opt,name=turtle_penalty,json=turtlePenalty" json:"turtle_penalty,omitempty"`
}

func (m *CoCoInterferenceScores) Reset()                    { *m = CoCoInterferenceScores{} }
func (m *CoCoInterferenceScores) String() string            { return proto.CompactTextString(m) }
func (*CoCoInterferenceScores) ProtoMessage()               {}
func (*CoCoInterferenceScores) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CoCoInterferenceScores) GetDevilPenalty() uint32 {
	if m != nil {
		return m.DevilPenalty
	}
	return 0
}

func (m *CoCoInterferenceScores) GetRabbitPenalty() uint32 {
	if m != nil {
		return m.RabbitPenalty
	}
	return 0
}

func (m *CoCoInterferenceScores) GetSheepPenalty() uint32 {
	if m != nil {
		return m.SheepPenalty
	}
	return 0
}

func (m *CoCoInterferenceScores) GetTurtlePenalty() uint32 {
	if m != nil {
		return m.TurtlePenalty
	}
	return 0
}

func init() {
	proto.RegisterType((*CoCoInterferenceScores)(nil), "firmament.CoCoInterferenceScores")
}

func init() { proto.RegisterFile("coco_interference_scores.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 169 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0xce, 0x4f, 0xce,
	0x8f, 0xcf, 0xcc, 0x2b, 0x49, 0x2d, 0x4a, 0x4b, 0x2d, 0x4a, 0xcd, 0x4b, 0x4e, 0x8d, 0x2f, 0x4e,
	0xce, 0x2f, 0x4a, 0x2d, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0xcb, 0x2c, 0xca,
	0x4d, 0xcc, 0x4d, 0xcd, 0x2b, 0x51, 0xda, 0xc0, 0xc8, 0x25, 0xe6, 0x9c, 0xef, 0x9c, 0xef, 0x89,
	0xa4, 0x38, 0x18, 0xac, 0x56, 0x48, 0x99, 0x8b, 0x37, 0x25, 0xb5, 0x2c, 0x33, 0x27, 0xbe, 0x20,
	0x35, 0x2f, 0x31, 0xa7, 0xa4, 0x52, 0x82, 0x51, 0x81, 0x51, 0x83, 0x37, 0x88, 0x07, 0x2c, 0x18,
	0x00, 0x11, 0x13, 0x52, 0xe5, 0xe2, 0x2b, 0x4a, 0x4c, 0x4a, 0xca, 0x2c, 0x81, 0xab, 0x62, 0x02,
	0xab, 0xe2, 0x85, 0x88, 0xc2, 0x94, 0x29, 0x73, 0xf1, 0x16, 0x67, 0xa4, 0xa6, 0x16, 0xc0, 0x55,
	0x31, 0x43, 0xcc, 0x02, 0x0b, 0x22, 0x99, 0x55, 0x52, 0x5a, 0x54, 0x92, 0x93, 0x0a, 0x57, 0xc5,
	0x02, 0x31, 0x0b, 0x22, 0x0a, 0x55, 0x96, 0xc4, 0x06, 0xf6, 0x84, 0x31, 0x20, 0x00, 0x00, 0xff,
	0xff, 0xce, 0x84, 0x0e, 0x21, 0xe6, 0x00, 0x00, 0x00,
}
