package specification_test

import (
	"testing"

	trafficmanage "github.com/pole-io/specification/source/go/api/v1/traffic_manage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestRateLimitCalleeContract(t *testing.T) {
	descriptor := (&trafficmanage.RateLimit{}).ProtoReflect().Descriptor()
	field := descriptor.Fields().ByName("callee")
	if field == nil {
		t.Fatal("RateLimit must expose callee")
	}
	if field.Number() != 16 || field.Kind() != protoreflect.MessageKind {
		t.Fatalf("callee descriptor = (%d, %s), want (16, message)", field.Number(), field.Kind())
	}

	rule := &trafficmanage.RateLimit{
		Namespace: "governance-production",
		Callee: &trafficmanage.DestinationService{
			Namespace: "runtime-production",
			Service:   "llm-gateway",
		},
	}
	encoded, err := protojson.Marshal(rule)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	decoded := &trafficmanage.RateLimit{}
	if err := protojson.Unmarshal(encoded, decoded); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if decoded.GetNamespace() != "governance-production" {
		t.Fatalf("owner namespace = %q", decoded.GetNamespace())
	}
	if decoded.GetCallee().GetNamespace() != "runtime-production" || decoded.GetCallee().GetService() != "llm-gateway" {
		t.Fatalf("callee after round trip = %+v", decoded.GetCallee())
	}
}
