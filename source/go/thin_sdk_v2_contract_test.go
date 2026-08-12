package specification_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func TestTrafficContextV1ContractAssets(t *testing.T) {
	const assetDirectory = "../../thin-sdk/traffic-context/v1"
	content, err := os.ReadFile(filepath.Join(assetDirectory, "conformance.json"))
	if err != nil {
		t.Fatalf("read TrafficContext vectors: %v", err)
	}
	var vectors struct {
		Contract        string `json:"contract"`
		ContractVersion string `json:"contract_version"`
		WireVersion     int    `json:"wire_version"`
		Valid           []struct {
			Input struct {
				Labels struct {
					Campaign *string `json:"campaign"`
					Lane     *string `json:"lane"`
					Bucket   *int    `json:"bucket"`
				} `json:"labels"`
			} `json:"input"`
			ExpectedBaggage string `json:"expected_baggage"`
		} `json:"valid"`
		SidecarReceive struct {
			Valid   []json.RawMessage `json:"valid"`
			Invalid []struct {
				Diagnostic string `json:"diagnostic"`
			} `json:"invalid"`
		} `json:"sidecar_receive"`
	}
	if err := json.Unmarshal(content, &vectors); err != nil {
		t.Fatalf("decode TrafficContext vectors: %v", err)
	}
	if vectors.Contract != "latticehub-traffic-context" ||
		vectors.ContractVersion != "1.0.0" || vectors.WireVersion != 1 {
		t.Fatalf("unexpected TrafficContext identity: %+v", vectors)
	}
	if len(vectors.Valid) != 4 || len(vectors.SidecarReceive.Valid) != 1 ||
		len(vectors.SidecarReceive.Invalid) != 6 {
		t.Fatalf("unexpected TrafficContext vector counts")
	}
	for _, vector := range vectors.Valid {
		labelPresence := map[string]bool{
			"latticehub.traffic.campaign=": vector.Input.Labels.Campaign != nil,
			"latticehub.traffic.lane=":     vector.Input.Labels.Lane != nil,
			"latticehub.traffic.bucket=":   vector.Input.Labels.Bucket != nil,
		}
		hasLabels := false
		for key, present := range labelPresence {
			hasLabels = hasLabels || present
			if strings.Contains(vector.ExpectedBaggage, key) != present {
				t.Fatalf("TrafficContext baggage presence mismatch for %q: %s", key, vector.ExpectedBaggage)
			}
		}
		if strings.Contains(vector.ExpectedBaggage, "latticehub.traffic.version=1") != hasLabels {
			t.Fatalf("TrafficContext baggage version presence mismatch: %s", vector.ExpectedBaggage)
		}
	}
	foundByteLimitVector := false
	for _, vector := range vectors.SidecarReceive.Invalid {
		foundByteLimitVector = foundByteLimitVector || vector.Diagnostic == "LABEL_TOO_LARGE"
	}
	if !foundByteLimitVector {
		t.Fatal("TrafficContext vectors missing UTF-8 byte-limit rejection")
	}

	sums, err := os.ReadFile(filepath.Join(assetDirectory, "SHA256SUMS"))
	if err != nil {
		t.Fatalf("read TrafficContext SHA256SUMS: %v", err)
	}
	expected := map[string]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(sums)), "\n") {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			t.Fatalf("invalid TrafficContext SHA256SUMS line %q", line)
		}
		expected[parts[1]] = parts[0]
	}
	for _, name := range []string{"schema.json", "conformance.json"} {
		asset, err := os.ReadFile(filepath.Join(assetDirectory, name))
		if err != nil {
			t.Fatalf("read TrafficContext asset %s: %v", name, err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(asset))
		if expected[name] != actual {
			t.Fatalf("TrafficContext checksum mismatch for %s", name)
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
