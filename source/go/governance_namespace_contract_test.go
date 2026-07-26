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

func TestAlpha37WireLayoutRemainsStable(t *testing.T) {
	tests := []struct {
		name        string
		message     proto.Message
		field       protoreflect.Name
		number      protoreflect.FieldNumber
		kind        protoreflect.Kind
		cardinality protoreflect.Cardinality
	}{
		{name: "TrafficMirror.caller", message: &trafficmanage.TrafficMirror{}, field: "caller", number: 4, kind: protoreflect.MessageKind},
		{name: "TrafficMirror.callee", message: &trafficmanage.TrafficMirror{}, field: "callee", number: 5, kind: protoreflect.MessageKind},
		{name: "TrafficMirror.rules", message: &trafficmanage.TrafficMirror{}, field: "rules", number: 6, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "TrafficMock.caller", message: &trafficmanage.TrafficMock{}, field: "caller", number: 4, kind: protoreflect.MessageKind},
		{name: "TrafficMock.callee", message: &trafficmanage.TrafficMock{}, field: "callee", number: 5, kind: protoreflect.MessageKind},
		{name: "TrafficMock.rules", message: &trafficmanage.TrafficMock{}, field: "rules", number: 6, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "MirrorRule.apis", message: &trafficmanage.MirrorRule{}, field: "apis", number: 1, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "MirrorRule.duration", message: &trafficmanage.MirrorRule{}, field: "duration", number: 5, kind: protoreflect.MessageKind},
		{name: "MirrorRule.disable", message: &trafficmanage.MirrorRule{}, field: "disable", number: 6, kind: protoreflect.BoolKind},
		{name: "MockRule.apis", message: &trafficmanage.MockRule{}, field: "apis", number: 1, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "MockRule.delay", message: &trafficmanage.MockRule{}, field: "delay", number: 5, kind: protoreflect.MessageKind},
		{name: "MockRule.disable", message: &trafficmanage.MockRule{}, field: "disable", number: 6, kind: protoreflect.BoolKind},
		{name: "MockResponse.code", message: &trafficmanage.MockResponse{}, field: "code", number: 1, kind: protoreflect.StringKind},
		{name: "LimitTrigger.apis", message: &trafficmanage.LimitTrigger{}, field: "apis", number: 2, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "BlockConfig.apis", message: &faulttolerance.BlockConfig{}, field: "apis", number: 2, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "BlockConfig.regex_separate", message: &faulttolerance.BlockConfig{}, field: "regex_separate", number: 5, kind: protoreflect.BoolKind},
		{name: "TrafficSecurityPolicy.apis", message: &security.TrafficSecurityPolicy{}, field: "apis", number: 1, kind: protoreflect.MessageKind, cardinality: protoreflect.Repeated},
		{name: "TrafficSecurityPolicy.managed_caller", message: &security.TrafficSecurityPolicy{}, field: "managed_caller", number: 5, kind: protoreflect.MessageKind},
		{name: "TrafficSecurityRejectEffect.code", message: &security.TrafficSecurityRejectEffect{}, field: "code", number: 1, kind: protoreflect.StringKind},
		{name: "FallbackResponse.code", message: &faulttolerance.FallbackResponse{}, field: "code", number: 1, kind: protoreflect.StringKind},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := tt.message.ProtoReflect().Descriptor().Fields().ByName(tt.field)
			if field == nil {
				t.Fatalf("field %q is missing", tt.field)
			}
			if field.Number() != tt.number {
				t.Fatalf("field number = %d, want %d", field.Number(), tt.number)
			}
			if field.Kind() != tt.kind {
				t.Fatalf("field kind = %s, want %s", field.Kind(), tt.kind)
			}
			wantCardinality := tt.cardinality
			if wantCardinality == 0 {
				wantCardinality = protoreflect.Optional
			}
			if field.Cardinality() != wantCardinality {
				t.Fatalf("field cardinality = %s, want %s", field.Cardinality(), wantCardinality)
			}

			message := tt.message.ProtoReflect()
			switch {
			case field.IsList():
				list := message.Mutable(field).List()
				list.Append(list.NewElement())
			case field.Kind() == protoreflect.MessageKind:
				message.Mutable(field)
			case field.Kind() == protoreflect.BoolKind:
				message.Set(field, protoreflect.ValueOfBool(true))
			case field.Kind() == protoreflect.StringKind:
				message.Set(field, protoreflect.ValueOfString("alpha-37"))
			default:
				t.Fatalf("unsupported test field kind %s", field.Kind())
			}

			encoded, err := proto.Marshal(tt.message)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			number, _, n := protowire.ConsumeTag(encoded)
			if n < 0 || number != protowire.Number(tt.number) {
				t.Fatalf("encoded field number = %d, want %d", number, tt.number)
			}

			decoded := tt.message.ProtoReflect().Type().New().Interface()
			if err := proto.Unmarshal(encoded, decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !decoded.ProtoReflect().Has(field) {
				t.Fatalf("field %q was lost after wire round trip", tt.field)
			}
		})
	}
}
