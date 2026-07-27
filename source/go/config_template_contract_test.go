package specification_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	configmanage "github.com/pole-io/specification/source/go/api/v1/config_manage"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestConfigTemplateSharedVectorsAreLanguageNeutral(t *testing.T) {
	content, err := os.ReadFile("../../CONFIG_TEMPLATE_TEST_VECTORS.json")
	if err != nil {
		t.Fatalf("read shared template vectors: %v", err)
	}
	var document struct {
		Profile string `json:"profile"`
		Vectors []struct {
			Name           string `json:"name"`
			Valid          bool   `json:"valid"`
			Expected       string `json:"expected"`
			ExpectedSHA256 string `json:"expected_sha256"`
			DiagnosticCode string `json:"diagnostic_code"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatalf("decode shared template vectors: %v", err)
	}
	if document.Profile != "pole-mustache-v1" || len(document.Vectors) < 5 {
		t.Fatalf("unexpected shared vector document: profile=%q vectors=%d",
			document.Profile, len(document.Vectors))
	}
	for _, vector := range document.Vectors {
		if vector.Name == "" {
			t.Fatal("shared vector name must not be empty")
		}
		if vector.Valid && (vector.Expected == "" || len(vector.ExpectedSHA256) != 64) {
			t.Fatalf("valid vector %q must carry exact output and SHA-256", vector.Name)
		}
		if !vector.Valid && vector.DiagnosticCode == "" {
			t.Fatalf("invalid vector %q must carry a diagnostic code", vector.Name)
		}
	}
}

func TestConfigFileTemplateModeKeepsLegacyWireLayout(t *testing.T) {
	messages := []proto.Message{
		&configmanage.ConfigFile{},
		&configmanage.ConfigFileRelease{},
		&configmanage.ConfigFileReleaseHistory{},
		&configmanage.ConfigFilePublishInfo{},
	}

	for _, message := range messages {
		descriptor := message.ProtoReflect().Descriptor()
		t.Run(string(descriptor.Name()), func(t *testing.T) {
			configType := descriptor.Fields().ByName("config_type")
			if configType == nil || configType.Number() != 200 || configType.Kind() != protoreflect.EnumKind {
				t.Fatalf("config_type must remain enum field 200, got %v", configType)
			}
			if got := configType.Enum().Values().ByName("CONFIG_FILE").Number(); got != 0 {
				t.Fatalf("CONFIG_FILE number = %d, want 0", got)
			}
			if got := configType.Enum().Values().ByName("CONFIG_TEMPLATE").Number(); got != 1 {
				t.Fatalf("CONFIG_TEMPLATE number = %d, want 1", got)
			}

			placeholderValues := descriptor.Fields().ByName("placeholder_value_map")
			if placeholderValues == nil || placeholderValues.Number() != 201 {
				t.Fatalf("placeholder_value_map must remain field 201")
			}
			options, ok := placeholderValues.Options().(*descriptorpb.FieldOptions)
			if !ok || !options.GetDeprecated() {
				t.Fatalf("placeholder_value_map must be deprecated")
			}

			binding := descriptor.Fields().ByName("template_binding")
			if binding == nil || binding.Number() != 202 || binding.Kind() != protoreflect.MessageKind {
				t.Fatalf("template_binding must be message field 202, got %v", binding)
			}
		})
	}

	legacyPayload := protowire.AppendVarint(
		protowire.AppendTag(nil, 200, protowire.VarintType),
		uint64(configmanage.ConfigFile_CONFIG_TEMPLATE),
	)
	var decoded configmanage.ConfigFile
	if err := proto.Unmarshal(legacyPayload, &decoded); err != nil {
		t.Fatalf("unmarshal legacy config file: %v", err)
	}
	if decoded.GetConfigType() != configmanage.ConfigFile_CONFIG_TEMPLATE {
		t.Fatalf("legacy config type = %s, want CONFIG_TEMPLATE", decoded.GetConfigType())
	}
	if decoded.GetTemplateBinding() != nil {
		t.Fatalf("legacy payload must leave template binding unset")
	}
}

func TestRenderSnapshotCarriesTypedValueAndDiscoverResponse(t *testing.T) {
	snapshot := &configmanage.RenderSnapshot{
		TemplateBinding: &configmanage.ConfigTemplateBinding{
			TemplateId:        42,
			TemplateReleaseId: "template-release-7",
			BindingReleaseId:  "binding-release-3",
		},
		TemplateRelease: &configmanage.ConfigTemplateRelease{
			Id:         "template-release-7",
			TemplateId: 42,
			Content:    "port: {{{server.port}}}\n",
			Format:     "yaml",
			Engine: &configmanage.ConfigTemplateEngine{
				Name:    "pole-mustache",
				Version: "v1",
			},
			ParameterSchema: []*configmanage.ConfigTemplateParameterSchema{{
				Name:     "server.port",
				Type:     configmanage.ConfigTemplateParameterType_TEMPLATE_PARAMETER_INTEGER,
				Required: true,
			}},
		},
		ValueRelease: &configmanage.NamespaceTemplateValueRelease{
			Id:                "value-release-9",
			Namespace:         "production",
			TemplateId:        42,
			TemplateReleaseId: "template-release-7",
			ReleaseType:       configmanage.NamespaceTemplateValueReleaseType_TEMPLATE_VALUE_RELEASE_GRAY,
			Priority:          100,
			Active:            true,
			Version:           9,
			Values: map[string]*configmanage.ConfigTemplateValue{
				"server.port": {
					Value: &configmanage.ConfigTemplateValue_IntegerValue{IntegerValue: 8080},
				},
			},
		},
		Revision:               "combined-revision",
		ExpectedRenderedSha256: "reference-sha256",
	}
	response := &configmanage.ConfigDiscoverResponse{
		Type:           configmanage.ConfigDiscoverResponse_CONFIG_FILE,
		Revision:       snapshot.Revision,
		RenderSnapshot: snapshot,
	}

	encoded, err := proto.Marshal(response)
	if err != nil {
		t.Fatalf("marshal discover response: %v", err)
	}
	var decoded configmanage.ConfigDiscoverResponse
	if err := proto.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal discover response: %v", err)
	}
	gotValue := decoded.GetRenderSnapshot().GetValueRelease().GetValues()["server.port"]
	if got := gotValue.GetIntegerValue(); got != 8080 {
		t.Fatalf("integer template value = %d, want 8080", got)
	}
	if got := decoded.GetRenderSnapshot().GetExpectedRenderedSha256(); got != "reference-sha256" {
		t.Fatalf("expected rendered sha256 = %q", got)
	}

	jsonEncoded, err := protojson.Marshal(response)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	for _, expected := range [][]byte{
		[]byte(`"render_snapshot"`),
		[]byte(`"template_release_id":"template-release-7"`),
		[]byte(`"integer_value":"8080"`),
	} {
		if !bytes.Contains(jsonEncoded, expected) {
			t.Fatalf("JSON payload %s does not contain %s", jsonEncoded, expected)
		}
	}

	field := response.ProtoReflect().Descriptor().Fields().ByName("render_snapshot")
	if field == nil || field.Number() != 8 {
		t.Fatalf("render_snapshot must use new field 8")
	}
}

func TestTemplateEngineCapabilityAndPreviewContract(t *testing.T) {
	filter := &configmanage.ConfigDiscoverFilter{
		SupportedTemplateEngines: []*configmanage.ConfigTemplateEngine{{
			Name:    "pole-mustache",
			Version: "v1",
		}},
	}
	encoded, err := proto.Marshal(filter)
	if err != nil {
		t.Fatalf("marshal capability: %v", err)
	}
	number, wireType, n := protowire.ConsumeTag(encoded)
	if n < 0 || number != 3 || wireType != protowire.BytesType {
		t.Fatalf("capability wire tag = (%d, %v), want (3, bytes)", number, wireType)
	}

	preview := &configmanage.RenderPreview{
		RenderedContent:   "port: 8080\n",
		Format:            "yaml",
		TemplateReleaseId: "template-release-7",
		ValueReleaseId:    "value-release-9",
		Engine: &configmanage.ConfigTemplateEngine{
			Name:    "pole-mustache",
			Version: "v1",
		},
		RenderedSha256: "reference-sha256",
		Diagnostics: []*configmanage.RenderDiagnostic{{
			Severity: configmanage.RenderDiagnostic_DIAGNOSTIC_INFO,
			Code:     "VALID",
			Message:  "rendered output is valid",
		}},
		Valid: true,
		Code:  uint32(200000),
		Info:  "ExecuteSuccess",
	}
	previewEncoded, err := proto.Marshal(preview)
	if err != nil {
		t.Fatalf("marshal preview: %v", err)
	}
	var previewDecoded configmanage.RenderPreview
	if err := proto.Unmarshal(previewEncoded, &previewDecoded); err != nil {
		t.Fatalf("unmarshal preview: %v", err)
	}
	if !previewDecoded.GetValid() || previewDecoded.GetRenderedContent() != "port: 8080\n" {
		t.Fatalf("preview round trip lost reference result: %+v", previewDecoded)
	}
	if previewDecoded.GetCode() != 200000 {
		t.Fatalf("preview result code = %d, want 200000", previewDecoded.GetCode())
	}
	previewDescriptor := preview.ProtoReflect().Descriptor()
	if previewDescriptor.Fields().ByName("code").Number() != 9 ||
		previewDescriptor.Fields().ByName("info").Number() != 10 {
		t.Fatal("RenderPreview code/info must use fields 9/10")
	}

	request := &configmanage.RenderPreviewRequest{
		Input: &configmanage.ConfigTemplateRenderInput{
			Content: "port: {{{server.port}}}\n",
			Format:  "yaml",
			Engine: &configmanage.ConfigTemplateEngine{
				Name:    "pole-mustache",
				Version: "v1",
			},
			Values: map[string]*configmanage.ConfigTemplateValue{
				"server.port": {
					Value: &configmanage.ConfigTemplateValue_IntegerValue{IntegerValue: 8080},
				},
			},
		},
	}
	requestEncoded, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("marshal draft preview request: %v", err)
	}
	var requestDecoded configmanage.RenderPreviewRequest
	if err := proto.Unmarshal(requestEncoded, &requestDecoded); err != nil {
		t.Fatalf("unmarshal draft preview request: %v", err)
	}
	if got := requestDecoded.GetInput().GetContent(); got != "port: {{{server.port}}}\n" {
		t.Fatalf("draft preview content = %q", got)
	}

	service := configmanage.File_grpc_config_api_proto.Services().ByName("ConfigGRPC")
	if service == nil || service.Methods().ByName("PreviewConfigTemplate") == nil {
		t.Fatalf("ConfigGRPC must expose PreviewConfigTemplate")
	}
}
