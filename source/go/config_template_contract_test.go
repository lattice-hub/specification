package specification_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	config "github.com/pole-io/specification/source/go/api/v1/config"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// 向量校验只证明输入合同与预期摘要自洽，不代替 SDK/控制台或可选兼容适配器的渲染验收。
func TestConfigTemplateVectorsUseUnifiedPreviewContract(t *testing.T) {
	content, err := os.ReadFile("../../CONFIG_TEMPLATE_TEST_VECTORS.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Profile string `json:"profile"`
		Vectors []struct {
			Name            string          `json:"name"`
			Request         json.RawMessage `json:"request"`
			Valid           bool            `json:"valid"`
			ExpectedContent *string         `json:"expectedContent"`
			ExpectedSHA256  string          `json:"expectedSha256"`
			Reason          string          `json:"reason"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	if document.Profile != "pole-mustache-v1" || len(document.Vectors) == 0 {
		t.Fatal("missing rendering vectors")
	}
	names := map[string]bool{}
	for _, vector := range document.Vectors {
		if vector.Name == "" || names[vector.Name] {
			t.Fatalf("duplicate/empty vector name: %q", vector.Name)
		}
		names[vector.Name] = true
		t.Run(vector.Name, func(t *testing.T) {
			var request config.PreviewRequest
			if err := protojson.Unmarshal(vector.Request, &request); err != nil {
				t.Fatal(err)
			}
			if request.Key == nil || request.Definition == nil {
				t.Fatal("missing preview input")
			}
			wire, err := proto.Marshal(&request)
			if err != nil {
				t.Fatal(err)
			}
			var decoded config.PreviewRequest
			if err := proto.Unmarshal(wire, &decoded); err != nil || !proto.Equal(&request, &decoded) {
				t.Fatalf("input round trip: %v", err)
			}
			if !vector.Valid {
				if vector.Reason == "" || vector.ExpectedContent != nil {
					t.Fatal("invalid vector must state failure without successful output")
				}
				return
			}
			if vector.ExpectedContent == nil || vector.Reason != "" {
				t.Fatal("valid vector must carry exact output, including empty output")
			}
			output, err := base64.StdEncoding.DecodeString(*vector.ExpectedContent)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(output)
			if hex.EncodeToString(digest[:]) != vector.ExpectedSHA256 {
				t.Fatal("expected bytes/hash mismatch")
			}
		})
	}
}
