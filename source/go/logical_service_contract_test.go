package specification_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	servicemanage "github.com/pole-io/specification/source/go/api/v1/service_manage"
)

func TestLogicalServiceManagementMessagesRoundTripWithoutChangingRuntimeService(t *testing.T) {
	logical := &servicemanage.LogicalService{
		Id: "logical-1", Name: "checkout", EnvironmentCount: 2,
		HealthyInstanceCount: 3, TotalInstanceCount: 4,
	}
	encoded, err := protojson.Marshal(logical)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"environment_count":2`) {
		t.Fatalf("logical service JSON does not use the management field contract: %s", encoded)
	}
	roundTrip := &servicemanage.LogicalService{}
	if err := protojson.Unmarshal(encoded, roundTrip); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(logical, roundTrip) {
		t.Fatalf("logical service round trip mismatch: got %v want %v", roundTrip, logical)
	}
	if _, err := anypb.New(logical); err != nil {
		t.Fatalf("logical service must support standard management Any responses: %v", err)
	}

	runtimeService := &servicemanage.Service{Name: "checkout-prod", Namespace: "production"}
	runtimeJSON, err := protojson.Marshal(runtimeService)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(runtimeJSON), "logical_service") {
		t.Fatalf("runtime Service unexpectedly exposes logical service identity: %s", runtimeJSON)
	}
}
