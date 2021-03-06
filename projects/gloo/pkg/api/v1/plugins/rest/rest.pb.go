// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/gloo/api/v1/plugins/rest/rest.proto

package rest

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	transformation "github.com/solo-io/gloo/projects/gloo/pkg/api/external/envoy/extensions/transformation"
	transformation1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/transformation"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ServiceSpec struct {
	Transformations      map[string]*transformation.TransformationTemplate `protobuf:"bytes,1,rep,name=transformations,proto3" json:"transformations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	SwaggerInfo          *ServiceSpec_SwaggerInfo                          `protobuf:"bytes,2,opt,name=swagger_info,json=swaggerInfo,proto3" json:"swagger_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                          `json:"-"`
	XXX_unrecognized     []byte                                            `json:"-"`
	XXX_sizecache        int32                                             `json:"-"`
}

func (m *ServiceSpec) Reset()         { *m = ServiceSpec{} }
func (m *ServiceSpec) String() string { return proto.CompactTextString(m) }
func (*ServiceSpec) ProtoMessage()    {}
func (*ServiceSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_10f084fc89ebe515, []int{0}
}
func (m *ServiceSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSpec.Unmarshal(m, b)
}
func (m *ServiceSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSpec.Marshal(b, m, deterministic)
}
func (m *ServiceSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSpec.Merge(m, src)
}
func (m *ServiceSpec) XXX_Size() int {
	return xxx_messageInfo_ServiceSpec.Size(m)
}
func (m *ServiceSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSpec.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSpec proto.InternalMessageInfo

func (m *ServiceSpec) GetTransformations() map[string]*transformation.TransformationTemplate {
	if m != nil {
		return m.Transformations
	}
	return nil
}

func (m *ServiceSpec) GetSwaggerInfo() *ServiceSpec_SwaggerInfo {
	if m != nil {
		return m.SwaggerInfo
	}
	return nil
}

type ServiceSpec_SwaggerInfo struct {
	// Types that are valid to be assigned to SwaggerSpec:
	//	*ServiceSpec_SwaggerInfo_Url
	//	*ServiceSpec_SwaggerInfo_Inline
	SwaggerSpec          isServiceSpec_SwaggerInfo_SwaggerSpec `protobuf_oneof:"swagger_spec"`
	XXX_NoUnkeyedLiteral struct{}                              `json:"-"`
	XXX_unrecognized     []byte                                `json:"-"`
	XXX_sizecache        int32                                 `json:"-"`
}

func (m *ServiceSpec_SwaggerInfo) Reset()         { *m = ServiceSpec_SwaggerInfo{} }
func (m *ServiceSpec_SwaggerInfo) String() string { return proto.CompactTextString(m) }
func (*ServiceSpec_SwaggerInfo) ProtoMessage()    {}
func (*ServiceSpec_SwaggerInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_10f084fc89ebe515, []int{0, 1}
}
func (m *ServiceSpec_SwaggerInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServiceSpec_SwaggerInfo.Unmarshal(m, b)
}
func (m *ServiceSpec_SwaggerInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServiceSpec_SwaggerInfo.Marshal(b, m, deterministic)
}
func (m *ServiceSpec_SwaggerInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServiceSpec_SwaggerInfo.Merge(m, src)
}
func (m *ServiceSpec_SwaggerInfo) XXX_Size() int {
	return xxx_messageInfo_ServiceSpec_SwaggerInfo.Size(m)
}
func (m *ServiceSpec_SwaggerInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ServiceSpec_SwaggerInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ServiceSpec_SwaggerInfo proto.InternalMessageInfo

type isServiceSpec_SwaggerInfo_SwaggerSpec interface {
	isServiceSpec_SwaggerInfo_SwaggerSpec()
	Equal(interface{}) bool
}

type ServiceSpec_SwaggerInfo_Url struct {
	Url string `protobuf:"bytes,1,opt,name=url,proto3,oneof" json:"url,omitempty"`
}
type ServiceSpec_SwaggerInfo_Inline struct {
	Inline string `protobuf:"bytes,2,opt,name=inline,proto3,oneof" json:"inline,omitempty"`
}

func (*ServiceSpec_SwaggerInfo_Url) isServiceSpec_SwaggerInfo_SwaggerSpec()    {}
func (*ServiceSpec_SwaggerInfo_Inline) isServiceSpec_SwaggerInfo_SwaggerSpec() {}

func (m *ServiceSpec_SwaggerInfo) GetSwaggerSpec() isServiceSpec_SwaggerInfo_SwaggerSpec {
	if m != nil {
		return m.SwaggerSpec
	}
	return nil
}

func (m *ServiceSpec_SwaggerInfo) GetUrl() string {
	if x, ok := m.GetSwaggerSpec().(*ServiceSpec_SwaggerInfo_Url); ok {
		return x.Url
	}
	return ""
}

func (m *ServiceSpec_SwaggerInfo) GetInline() string {
	if x, ok := m.GetSwaggerSpec().(*ServiceSpec_SwaggerInfo_Inline); ok {
		return x.Inline
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*ServiceSpec_SwaggerInfo) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*ServiceSpec_SwaggerInfo_Url)(nil),
		(*ServiceSpec_SwaggerInfo_Inline)(nil),
	}
}

// This is only for upstream with REST service spec
type DestinationSpec struct {
	FunctionName           string                                 `protobuf:"bytes,1,opt,name=function_name,json=functionName,proto3" json:"function_name,omitempty"`
	Parameters             *transformation1.Parameters            `protobuf:"bytes,2,opt,name=parameters,proto3" json:"parameters,omitempty"`
	ResponseTransformation *transformation.TransformationTemplate `protobuf:"bytes,3,opt,name=response_transformation,json=responseTransformation,proto3" json:"response_transformation,omitempty"`
	XXX_NoUnkeyedLiteral   struct{}                               `json:"-"`
	XXX_unrecognized       []byte                                 `json:"-"`
	XXX_sizecache          int32                                  `json:"-"`
}

func (m *DestinationSpec) Reset()         { *m = DestinationSpec{} }
func (m *DestinationSpec) String() string { return proto.CompactTextString(m) }
func (*DestinationSpec) ProtoMessage()    {}
func (*DestinationSpec) Descriptor() ([]byte, []int) {
	return fileDescriptor_10f084fc89ebe515, []int{1}
}
func (m *DestinationSpec) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DestinationSpec.Unmarshal(m, b)
}
func (m *DestinationSpec) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DestinationSpec.Marshal(b, m, deterministic)
}
func (m *DestinationSpec) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DestinationSpec.Merge(m, src)
}
func (m *DestinationSpec) XXX_Size() int {
	return xxx_messageInfo_DestinationSpec.Size(m)
}
func (m *DestinationSpec) XXX_DiscardUnknown() {
	xxx_messageInfo_DestinationSpec.DiscardUnknown(m)
}

var xxx_messageInfo_DestinationSpec proto.InternalMessageInfo

func (m *DestinationSpec) GetFunctionName() string {
	if m != nil {
		return m.FunctionName
	}
	return ""
}

func (m *DestinationSpec) GetParameters() *transformation1.Parameters {
	if m != nil {
		return m.Parameters
	}
	return nil
}

func (m *DestinationSpec) GetResponseTransformation() *transformation.TransformationTemplate {
	if m != nil {
		return m.ResponseTransformation
	}
	return nil
}

func init() {
	proto.RegisterType((*ServiceSpec)(nil), "rest.plugins.gloo.solo.io.ServiceSpec")
	proto.RegisterMapType((map[string]*transformation.TransformationTemplate)(nil), "rest.plugins.gloo.solo.io.ServiceSpec.TransformationsEntry")
	proto.RegisterType((*ServiceSpec_SwaggerInfo)(nil), "rest.plugins.gloo.solo.io.ServiceSpec.SwaggerInfo")
	proto.RegisterType((*DestinationSpec)(nil), "rest.plugins.gloo.solo.io.DestinationSpec")
}

func init() {
	proto.RegisterFile("github.com/solo-io/gloo/projects/gloo/api/v1/plugins/rest/rest.proto", fileDescriptor_10f084fc89ebe515)
}

var fileDescriptor_10f084fc89ebe515 = []byte{
	// 453 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x25, 0x1b, 0xb1, 0x12, 0xce, 0xc2, 0x22, 0x6b, 0x05, 0xa1, 0x07, 0x54, 0x2d, 0x97, 0x5e,
	0xb0, 0xa1, 0x5c, 0x10, 0x88, 0xcb, 0xb2, 0x20, 0x10, 0x12, 0xa0, 0xb4, 0x5c, 0xb8, 0x54, 0xde,
	0x68, 0x92, 0x35, 0xeb, 0x78, 0x2c, 0xdb, 0x09, 0xf4, 0x27, 0xf8, 0x0e, 0xbe, 0x8b, 0x4f, 0xe0,
	0x0b, 0x50, 0xe2, 0x2c, 0x6d, 0xa3, 0x22, 0x55, 0xbd, 0x44, 0xf3, 0xec, 0xbc, 0xf7, 0x32, 0x6f,
	0x26, 0xe4, 0xbc, 0x94, 0xfe, 0xb2, 0xbe, 0x60, 0x39, 0x56, 0xdc, 0xa1, 0xc2, 0xc7, 0x12, 0x79,
	0xa9, 0x10, 0xb9, 0xb1, 0xf8, 0x0d, 0x72, 0xef, 0x02, 0x12, 0x46, 0xf2, 0xe6, 0x29, 0x37, 0xaa,
	0x2e, 0xa5, 0x76, 0xdc, 0x82, 0xf3, 0xdd, 0x83, 0x19, 0x8b, 0x1e, 0xe9, 0x83, 0x50, 0x87, 0x5b,
	0xd6, 0x32, 0x58, 0x2b, 0xc6, 0x24, 0x8e, 0x4e, 0x4a, 0x2c, 0xb1, 0x7b, 0x8b, 0xb7, 0x55, 0x20,
	0x8c, 0xca, 0xdd, 0x6d, 0xe1, 0x87, 0x07, 0xab, 0x85, 0xe2, 0xa0, 0x1b, 0x5c, 0x76, 0x50, 0x3b,
	0x89, 0xda, 0x71, 0x6f, 0x85, 0x76, 0x05, 0xda, 0x4a, 0x78, 0x89, 0x7a, 0x00, 0x7b, 0xa3, 0xf9,
	0x5e, 0xfd, 0x0d, 0x94, 0x8d, 0xb0, 0xa2, 0x02, 0x0f, 0xd6, 0x05, 0xd5, 0xd3, 0x9f, 0x31, 0x49,
	0x66, 0x60, 0x1b, 0x99, 0xc3, 0xcc, 0x40, 0x4e, 0x81, 0x1c, 0x6f, 0x52, 0x5c, 0x1a, 0x8d, 0xe3,
	0x49, 0x32, 0x7d, 0xc9, 0xfe, 0x9b, 0x0c, 0x5b, 0x13, 0x60, 0xf3, 0x4d, 0xf6, 0x1b, 0xed, 0xed,
	0x32, 0x1b, 0x6a, 0xd2, 0x2f, 0xe4, 0xc8, 0x7d, 0x17, 0x65, 0x09, 0x76, 0x21, 0x75, 0x81, 0xe9,
	0xc1, 0x38, 0x9a, 0x24, 0xd3, 0xe9, 0x8e, 0x1e, 0xb3, 0x40, 0x7d, 0xaf, 0x0b, 0xcc, 0x12, 0xb7,
	0x02, 0x23, 0x4f, 0x4e, 0xb6, 0xf9, 0xd3, 0xbb, 0x24, 0xbe, 0x82, 0x65, 0x1a, 0x8d, 0xa3, 0xc9,
	0xad, 0xac, 0x2d, 0xe9, 0x5b, 0x72, 0xb3, 0x11, 0xaa, 0x86, 0xde, 0xf9, 0x09, 0xeb, 0x66, 0xc2,
	0x84, 0x91, 0xac, 0x99, 0xb2, 0x42, 0x2a, 0x0f, 0x96, 0x5d, 0x7a, 0x6f, 0x06, 0x0d, 0xcd, 0xa1,
	0x32, 0x4a, 0x78, 0xc8, 0x02, 0xfd, 0xc5, 0xc1, 0xf3, 0x68, 0xf4, 0x81, 0x24, 0x6b, 0x5f, 0x44,
	0x29, 0x89, 0x6b, 0xab, 0x82, 0xd9, 0xbb, 0x1b, 0x59, 0x0b, 0x68, 0x4a, 0x0e, 0xa5, 0x56, 0x52,
	0x07, 0xbf, 0xf6, 0xb8, 0xc7, 0x67, 0x77, 0x56, 0x49, 0x38, 0x03, 0xf9, 0xe9, 0x9f, 0x88, 0x1c,
	0x9f, 0x83, 0xf3, 0x52, 0x77, 0x7e, 0xdd, 0x50, 0x1e, 0x91, 0xdb, 0x45, 0xad, 0xf3, 0x16, 0x2f,
	0xb4, 0xa8, 0xa0, 0x6f, 0xe4, 0xe8, 0xfa, 0xf0, 0xa3, 0xa8, 0x80, 0x7e, 0x22, 0x64, 0x35, 0xdd,
	0xbe, 0x2d, 0xce, 0x86, 0xab, 0xb4, 0x2d, 0xda, 0xcf, 0xff, 0x68, 0xd9, 0x9a, 0x04, 0x95, 0xe4,
	0xbe, 0x05, 0x67, 0x50, 0x3b, 0x58, 0x6c, 0xca, 0xa4, 0xf1, 0x9e, 0xa1, 0xdd, 0xbb, 0x16, 0xdc,
	0xbc, 0x3f, 0x7b, 0xfd, 0xeb, 0xf7, 0xc3, 0xe8, 0xeb, 0xab, 0xdd, 0x36, 0xdc, 0x5c, 0x95, 0xdb,
	0xfe, 0xe2, 0x8b, 0xc3, 0x6e, 0xa3, 0x9f, 0xfd, 0x0d, 0x00, 0x00, 0xff, 0xff, 0x59, 0xae, 0x50,
	0xbf, 0x09, 0x04, 0x00, 0x00,
}

func (this *ServiceSpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSpec)
	if !ok {
		that2, ok := that.(ServiceSpec)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.Transformations) != len(that1.Transformations) {
		return false
	}
	for i := range this.Transformations {
		if !this.Transformations[i].Equal(that1.Transformations[i]) {
			return false
		}
	}
	if !this.SwaggerInfo.Equal(that1.SwaggerInfo) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ServiceSpec_SwaggerInfo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSpec_SwaggerInfo)
	if !ok {
		that2, ok := that.(ServiceSpec_SwaggerInfo)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if that1.SwaggerSpec == nil {
		if this.SwaggerSpec != nil {
			return false
		}
	} else if this.SwaggerSpec == nil {
		return false
	} else if !this.SwaggerSpec.Equal(that1.SwaggerSpec) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ServiceSpec_SwaggerInfo_Url) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSpec_SwaggerInfo_Url)
	if !ok {
		that2, ok := that.(ServiceSpec_SwaggerInfo_Url)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Url != that1.Url {
		return false
	}
	return true
}
func (this *ServiceSpec_SwaggerInfo_Inline) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ServiceSpec_SwaggerInfo_Inline)
	if !ok {
		that2, ok := that.(ServiceSpec_SwaggerInfo_Inline)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Inline != that1.Inline {
		return false
	}
	return true
}
func (this *DestinationSpec) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*DestinationSpec)
	if !ok {
		that2, ok := that.(DestinationSpec)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.FunctionName != that1.FunctionName {
		return false
	}
	if !this.Parameters.Equal(that1.Parameters) {
		return false
	}
	if !this.ResponseTransformation.Equal(that1.ResponseTransformation) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
