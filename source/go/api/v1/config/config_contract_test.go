package config

import (
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestBindingPresenceAndManagementRoundTrip(t *testing.T) {
	d := &ConfigDefinition{ParameterBindings: []*ParameterBindings{{Name: "empty", Fallback: &Scalar{Type: ValueType_VALUE_TYPE_STRING}}, {Name: "missing"}}}
	b, err := proto.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ConfigDefinition
	if err := proto.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.ParameterBindings[0].Fallback == nil || decoded.ParameterBindings[0].Fallback.Value != "" || decoded.ParameterBindings[1].Fallback != nil {
		t.Fatal("fallback presence changed")
	}
	original := &ConfigFileList{Items: []*ConfigFileSummary{{Key: &ConfigFileKey{Group: "g", Name: "n"}, Environment: "prod", HasSharedTemplate: true, DraftRevision: 4}}, NextPageToken: "next"}
	b, err = proto.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var list ConfigFileList
	if err = proto.Unmarshal(b, &list); err != nil || !proto.Equal(original, &list) {
		t.Fatalf("management round trip: %v", err)
	}
}

func TestPromoteGrayRequestWireRoundTrip(t *testing.T) {
	original := &PromoteGrayRequest{Key: &ConfigKey{Environment: "prod", Group: "g", Name: "n"}, ReleaseId: "r", ExpectedReleaseSetRevision: ^uint64(0)}
	b, err := proto.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded PromoteGrayRequest
	if err = proto.Unmarshal(b, &decoded); err != nil || !proto.Equal(original, &decoded) {
		t.Fatalf("promote request: %v", err)
	}
}
