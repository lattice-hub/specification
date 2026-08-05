package specification_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestThinSDKV2ContractAssets(t *testing.T) {
	content, err := os.ReadFile("../../thin-sdk/target-service/v1/conformance.json")
	if err != nil {
		t.Fatalf("read TargetService vectors: %v", err)
	}
	var vectors struct {
		Contract        string `json:"contract"`
		ContractVersion string `json:"contract_version"`
		Valid           []struct {
			Input struct {
				Namespace string `json:"namespace"`
				Service   string `json:"service"`
			} `json:"input"`
			ExpectedMetadata [][2]string `json:"expected_metadata"`
		} `json:"valid"`
		Invalid []struct {
			Diagnostic string `json:"diagnostic"`
		} `json:"invalid"`
	}
	if err := json.Unmarshal(content, &vectors); err != nil {
		t.Fatalf("decode TargetService vectors: %v", err)
	}
	if vectors.Contract != "latticehub-target-service" || vectors.ContractVersion != "1.0.0" {
		t.Fatalf("unexpected contract identity: %+v", vectors)
	}
	if len(vectors.Valid) != 2 || len(vectors.Invalid) != 4 {
		t.Fatalf("unexpected TargetService vector counts")
	}
	for _, vector := range vectors.Valid {
		if len(vector.ExpectedMetadata) != 2 ||
			vector.ExpectedMetadata[0][0] != "latticehub-target-namespace" ||
			vector.ExpectedMetadata[1][0] != "latticehub-target-service" {
			t.Fatalf("invalid metadata keys: %#v", vector.ExpectedMetadata)
		}
		if strings.TrimSpace(vector.Input.Namespace) == "" || strings.TrimSpace(vector.Input.Service) == "" {
			t.Fatalf("valid vector has empty identity: %#v", vector.Input)
		}
	}
}

func TestSidecarBootstrapProtoContract(t *testing.T) {
	content, err := os.ReadFile("../../api/v1/sidecar/bootstrap.proto")
	if err != nil {
		t.Fatalf("read bootstrap proto: %v", err)
	}
	text := string(content)
	for _, required := range []string{
		"rpc OpenSession(ClientHello) returns (stream SidecarEvent)",
		"rpc OpenControlSession(stream ClientEvent) returns (stream SidecarEvent)",
		"PROTOCOL_HTTP = 1",
		"PROTOCOL_GRPC = 2",
		"PROTOCOL_DUBBO = 3",
		"PROTOCOL_THRIFT = 4",
		"ListenerSnapshot listener_snapshot = 1",
		"LocalServiceStatus local_service_status = 2",
		"LocalServiceRegistration register_local_service = 2",
		"LocalServiceUnregistration unregister_local_service = 3",
		"uint32 local_port = 5",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("bootstrap proto missing %q", required)
		}
	}
}
