package specification_test

import (
	"bytes"
	"testing"

	faulttolerance "github.com/pole-io/specification/source/go/api/v1/fault_tolerance"
	security "github.com/pole-io/specification/source/go/api/v1/security"
	trafficmanage "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestGovernanceRuleNamespaceContract(t *testing.T) {
	tests := []struct {
		name        string
		message     proto.Message
		fieldNumber protoreflect.FieldNumber
	}{
		{name: "RouteRule", message: &trafficmanage.RouteRule{}, fieldNumber: 15},
		{name: "RateLimit", message: &trafficmanage.RateLimit{}, fieldNumber: 15},
		{name: "CircuitBreakerRule", message: &faulttolerance.CircuitBreakerRule{}, fieldNumber: 16},
		{name: "FaultDetectRule", message: &faulttolerance.FaultDetectRule{}, fieldNumber: 13},
		{name: "LaneGroup", message: &trafficmanage.LaneGroup{}, fieldNumber: 13},
		{name: "LosslessRule", message: &trafficmanage.LosslessRule{}, fieldNumber: 10},
		{name: "TrafficSecurityRule", message: &security.TrafficSecurityRule{}, fieldNumber: 15},
		{name: "TrafficMirror", message: &trafficmanage.TrafficMirror{}, fieldNumber: 15},
		{name: "TrafficMock", message: &trafficmanage.TrafficMock{}, fieldNumber: 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			descriptor := tt.message.ProtoReflect().Descriptor()
			field := descriptor.Fields().ByName("namespace")
			if field == nil {
				t.Fatalf("%s must expose a top-level namespace field", descriptor.FullName())
			}
			if got := field.Number(); got != tt.fieldNumber {
				t.Fatalf("namespace field number = %d, want %d", got, tt.fieldNumber)
			}
			if field.Kind() != protoreflect.StringKind {
				t.Fatalf("namespace field kind = %s, want string", field.Kind())
			}

			const namespace = "production"
			tt.message.ProtoReflect().Set(field, protoreflect.ValueOfString(namespace))
			encoded, err := proto.Marshal(tt.message)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if number, wireType, n := protowire.ConsumeTag(encoded); n < 0 || number != protowire.Number(tt.fieldNumber) || wireType != protowire.BytesType {
				t.Fatalf("namespace wire tag = (%d, %v), want (%d, bytes)", number, wireType, tt.fieldNumber)
			}

			decoded := tt.message.ProtoReflect().Type().New().Interface()
			if err := proto.Unmarshal(encoded, decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := decoded.ProtoReflect().Get(field).String(); got != namespace {
				t.Fatalf("namespace after round trip = %q, want %q", got, namespace)
			}

			jsonEncoded, err := protojson.Marshal(tt.message)
			if err != nil {
				t.Fatalf("marshal JSON: %v", err)
			}
			if !bytes.Contains(jsonEncoded, []byte(`"namespace":"production"`)) {
				t.Fatalf("JSON payload %s does not contain namespace", jsonEncoded)
			}
			jsonDecoded := tt.message.ProtoReflect().Type().New().Interface()
			if err := protojson.Unmarshal(jsonEncoded, jsonDecoded); err != nil {
				t.Fatalf("unmarshal JSON: %v", err)
			}
			if got := jsonDecoded.ProtoReflect().Get(field).String(); got != namespace {
				t.Fatalf("JSON namespace after round trip = %q, want %q", got, namespace)
			}
		})
	}
}

func TestGovernanceRuleLegacyPayloadRemainsReadable(t *testing.T) {
	tests := []struct {
		name    string
		message proto.Message
	}{
		{name: "RouteRule", message: &trafficmanage.RouteRule{}},
		{name: "RateLimit", message: &trafficmanage.RateLimit{}},
		{name: "CircuitBreakerRule", message: &faulttolerance.CircuitBreakerRule{}},
		{name: "FaultDetectRule", message: &faulttolerance.FaultDetectRule{}},
		{name: "LaneGroup", message: &trafficmanage.LaneGroup{}},
		{name: "LosslessRule", message: &trafficmanage.LosslessRule{}},
		{name: "TrafficSecurityRule", message: &security.TrafficSecurityRule{}},
		{name: "TrafficMirror", message: &trafficmanage.TrafficMirror{}},
		{name: "TrafficMock", message: &trafficmanage.TrafficMock{}},
	}

	// Every governance root keeps field 1 as its historical ID. A payload written
	// before namespace existed therefore remains valid and leaves namespace empty.
	legacyPayload := protowire.AppendString(protowire.AppendTag(nil, 1, protowire.BytesType), "legacy-id")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := proto.Unmarshal(legacyPayload, tt.message); err != nil {
				t.Fatalf("unmarshal legacy payload: %v", err)
			}
			fields := tt.message.ProtoReflect().Descriptor().Fields()
			id := fields.ByNumber(1)
			if got := tt.message.ProtoReflect().Get(id).String(); got != "legacy-id" {
				t.Fatalf("legacy id = %q, want legacy-id", got)
			}
			namespace := fields.ByName("namespace")
			if namespace == nil {
				t.Fatalf("namespace descriptor missing")
			}
			if got := tt.message.ProtoReflect().Get(namespace).String(); got != "" {
				t.Fatalf("legacy namespace = %q, want empty", got)
			}

			legacyJSON := []byte(`{"id":"legacy-id"}`)
			jsonDecoded := tt.message.ProtoReflect().Type().New().Interface()
			if err := protojson.Unmarshal(legacyJSON, jsonDecoded); err != nil {
				t.Fatalf("unmarshal legacy JSON: %v", err)
			}
			if got := jsonDecoded.ProtoReflect().Get(namespace).String(); got != "" {
				t.Fatalf("legacy JSON namespace = %q, want empty", got)
			}
		})
	}
}
