package specification_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	apimodel "github.com/pole-io/specification/source/go/api/v1/model"
)

func TestNamespaceKindContractRoundTripsAcrossWireAndJSON(t *testing.T) {
	namespace := &apimodel.Namespace{
		Name: "pole-system",
		Kind: apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM,
	}

	wire, err := proto.Marshal(namespace)
	if err != nil {
		t.Fatal(err)
	}
	var kindField []byte
	for cursor := wire; len(cursor) > 0; {
		number, wireType, fieldLength := protowire.ConsumeField(cursor)
		if fieldLength < 0 {
			t.Fatalf("invalid namespace wire payload: %v", cursor)
		}
		if number == 10 {
			if wireType != protowire.VarintType {
				t.Fatalf("namespace kind field must use enum wire type")
			}
			kindField = cursor[:fieldLength]
			break
		}
		cursor = cursor[fieldLength:]
	}
	if len(kindField) == 0 {
		t.Fatalf("namespace kind must use enum wire field 10")
	}
	_, _, tagLength := protowire.ConsumeTag(kindField)
	value, _ := protowire.ConsumeVarint(kindField[tagLength:])
	if value != 1 {
		t.Fatalf("namespace SYSTEM wire value must be 1: %d", value)
	}
	roundTrip := &apimodel.Namespace{}
	if err := proto.Unmarshal(wire, roundTrip); err != nil {
		t.Fatal(err)
	}
	if roundTrip.GetKind() != apimodel.NamespaceKind_NAMESPACE_KIND_SYSTEM {
		t.Fatalf("namespace kind wire round trip mismatch: %v", roundTrip.GetKind())
	}

	encoded, err := protojson.Marshal(namespace)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"kind":"NAMESPACE_KIND_SYSTEM"`) {
		t.Fatalf("namespace kind JSON contract mismatch: %s", encoded)
	}
}

func TestNamespaceKindTreatsLegacyPayloadAsBusiness(t *testing.T) {
	legacyWire := protowire.AppendString(
		protowire.AppendTag(nil, 1, protowire.BytesType),
		"legacy-id",
	)
	namespace := &apimodel.Namespace{}
	if err := proto.Unmarshal(legacyWire, namespace); err != nil {
		t.Fatal(err)
	}
	if namespace.GetId() != "legacy-id" {
		t.Fatalf("legacy wire field was not preserved: %q", namespace.GetId())
	}
	if namespace.GetKind() != apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS {
		t.Fatalf("legacy payload must default to BUSINESS: %v", namespace.GetKind())
	}

	namespace.Reset()
	if err := protojson.Unmarshal([]byte(`{"id":"legacy-json"}`), namespace); err != nil {
		t.Fatal(err)
	}
	if namespace.GetId() != "legacy-json" ||
		namespace.GetKind() != apimodel.NamespaceKind_NAMESPACE_KIND_BUSINESS {
		t.Fatalf("legacy JSON must preserve fields and default to BUSINESS: %v", namespace)
	}

	field := namespace.ProtoReflect().Descriptor().Fields().ByName("kind")
	if field == nil || field.Number() != 10 || field.Enum() == nil {
		t.Fatalf("namespace kind descriptor must remain enum field 10: %v", field)
	}
}
