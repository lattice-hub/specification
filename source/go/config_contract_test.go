package specification_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	config "github.com/pole-io/specification/source/go/api/v1/config"
	configmanage "github.com/pole-io/specification/source/go/api/v1/config_manage"
	model "github.com/pole-io/specification/source/go/api/v1/model"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// 固定字节由 wire 合同定义，同一向量在 Rust smoke 使用；不由被测生成器自行生成预期值。
func TestConfigBinaryAndCursorWire(t *testing.T) {
	wire := []byte{0x08, 0x01, 0x2a, 0x05, 0x0a, 0x03, 0x00, 0xff, 0x80, 0x32, 0x0b, 0x18,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01}
	var result config.ConfigResult
	if err := proto.Unmarshal(wire, &result); err != nil {
		t.Fatal(err)
	}
	if result.Presence != config.ConfigPresence_CONFIG_PRESENCE_PRESENT ||
		!bytes.Equal(result.GetSnapshot().GetContent(), []byte{0, 255, 128}) ||
		result.GetObservationCursor().GetPosition() != math.MaxUint64 {
		t.Fatalf("binary/status/uint64 wire mismatch: %v", &result)
	}
	encoded, err := proto.Marshal(&result)
	if err != nil || !bytes.Equal(encoded, wire) {
		t.Fatalf("wire changed: %x, %v", encoded, err)
	}

	data, err := protojson.Marshal(&result)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Snapshot struct {
			Content string `json:"content"`
		} `json:"snapshot"`
		Cursor struct {
			Position string `json:"position"`
		} `json:"observationCursor"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Snapshot.Content != "AP+A" || document.Cursor.Position != "18446744073709551615" {
		t.Fatalf("JSON must preserve binary and uint64: %s", data)
	}
}

func TestConfigPreviewAndSummaryCannotExposeRuntimeOrDefinition(t *testing.T) {
	if _, err := protoregistry.GlobalFiles.FindFileByPath("config_file.proto"); err == nil {
		t.Fatal("the retired configuration model must not be exported")
	}
	preview := (&config.RenderSnapshot{}).ProtoReflect().Descriptor()
	for _, name := range []protoreflect.Name{"release_id", "resolved_revision", "observation_cursor", "status", "definition"} {
		if preview.Fields().ByName(name) != nil {
			t.Fatalf("preview must not expose %s", name)
		}
	}
	summary := (&config.ReleaseSummary{}).ProtoReflect().Descriptor()
	for i := 0; i < summary.Fields().Len(); i++ {
		field := summary.Fields().Get(i)
		if field.Message() != nil && field.Message().FullName() != "pole.config.v1.Audience" {
			t.Fatalf("summary must not contain a definition: %s", field.FullName())
		}
	}
	if summary.Fields().ByName("active") == nil || (&config.ConfigRelease{}).ProtoReflect().Descriptor().Fields().ByName("active") != nil {
		t.Fatal("active belongs to the current summary, not immutable release content")
	}
	for _, message := range []proto.Message{&config.Parameter{}, &config.Scalar{}} {
		field := message.ProtoReflect().Descriptor().Fields().ByName("type")
		if field.Enum() == nil || field.Enum().FullName() != "pole.config.v1.ValueType" {
			t.Fatal("parameter and scalar must share ValueType")
		}
	}
	options := config.File_config_proto.Options().(*descriptorpb.FileOptions)
	if options.GetJavaPackage() != "io.github.latticehub.pole.specification.api.v1.config" || options.GetJavaOuterClassname() != "ConfigProto" {
		t.Fatal("configuration Java messages must use the project package convention")
	}
}

func TestConfigScalarSharesRustExactWire(t *testing.T) {
	// DECIMAL=4，value="9007199254740993.0001"，不经浮点数转换。
	wire := append([]byte{0x08, 0x04, 0x12, 0x15}, []byte("9007199254740993.0001")...)
	var scalar config.Scalar
	if err := proto.Unmarshal(wire, &scalar); err != nil {
		t.Fatal(err)
	}
	if scalar.Type != config.ValueType_VALUE_TYPE_DECIMAL || scalar.Value != "9007199254740993.0001" {
		t.Fatal("scalar enum or exact value changed")
	}
	encoded, err := proto.Marshal(&scalar)
	if err != nil || !bytes.Equal(encoded, wire) {
		t.Fatalf("scalar wire changed: %x (%v)", encoded, err)
	}
	data, err := protojson.Marshal(&scalar)
	if err != nil || !bytes.Contains(data, []byte(`"VALUE_TYPE_DECIMAL"`)) {
		t.Fatalf("scalar JSON must use ValueType name: %s (%v)", data, err)
	}
}

func TestConfigRevisionAndDefinitionRoundTrip(t *testing.T) {
	state := &config.ConfigState{
		Key: &config.ConfigKey{Environment: "prod", Group: "payments", Name: "settings"},
		Draft: &config.ConfigDraft{Revision: math.MaxUint64, Definition: &config.ConfigDefinition{
			Content: []byte("amount: {{{amount}}}\r\n"), Engine: "pole-mustache-v1", Format: "yaml",
			Parameters:    []*config.Parameter{{Name: "amount", Type: config.ValueType_VALUE_TYPE_DECIMAL, Required: true}},
			DefaultValues: map[string]*config.Scalar{"amount": {Type: config.ValueType_VALUE_TYPE_DECIMAL, Value: "9007199254740993.0001"}},
			ValueRules: []*config.ValueRule{{Id: "cn", Priority: 0,
				Audience: &config.Audience{Labels: map[string]string{"region": "cn"}},
				Values:   map[string]*config.Scalar{"amount": {Type: config.ValueType_VALUE_TYPE_DECIMAL, Value: "0.0001"}}}},
		}},
		Releases: []*config.ReleaseSummary{{Active: true,
			Id: "gray-1", Kind: config.ReleaseKind_RELEASE_KIND_GRAY, Sequence: math.MaxUint64,
		}},
		ReleaseSetRevision: math.MaxUint64, Incarnation: "epoch-1",
	}
	wire, err := proto.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	var decoded config.ConfigState
	if err := proto.Unmarshal(wire, &decoded); err != nil || !proto.Equal(state, &decoded) {
		t.Fatalf("definition/revisions must round-trip losslessly: %v", err)
	}
}

func TestConfigStatusAndEmptyContentWire(t *testing.T) {
	statuses := []config.ConfigPresence{
		config.ConfigPresence_CONFIG_PRESENCE_UNSPECIFIED, config.ConfigPresence_CONFIG_PRESENCE_PRESENT,
		config.ConfigPresence_CONFIG_PRESENCE_ABSENT,
	}
	for number, expected := range statuses {
		var result config.ConfigResult
		if err := proto.Unmarshal([]byte{0x08, byte(number)}, &result); err != nil || result.Presence != expected {
			t.Fatalf("status %d wire mismatch: %v", number, err)
		}
	}
	present := &config.ConfigResult{Presence: config.ConfigPresence_CONFIG_PRESENCE_PRESENT, Snapshot: &config.RenderSnapshot{Engine: "literal"}}
	absent := &config.ConfigResult{Presence: config.ConfigPresence_CONFIG_PRESENCE_ABSENT}
	presentWire, err := proto.Marshal(present)
	if err != nil {
		t.Fatal(err)
	}
	absentWire, err := proto.Marshal(absent)
	if err != nil || bytes.Equal(presentWire, absentWire) {
		t.Fatalf("empty PRESENT must differ from ABSENT: %x / %x (%v)", presentWire, absentWire, err)
	}
}

func TestConfigCASAndSingleDefaultSourceDescriptors(t *testing.T) {
	cases := []struct {
		message proto.Message
		fields  map[protoreflect.Name]protoreflect.FieldNumber
	}{
		{&config.SaveDraftRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_draft_revision": 3}},
		{&config.PublishRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_draft_revision": 2, "expected_release_set_revision": 3}},
		{&config.StopRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_release_set_revision": 2}},
		{&config.StopAllGrayRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_release_set_revision": 2}},
		{&config.CopyReleaseToDraftRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_draft_revision": 3, "expected_release_set_revision": 4}},
		{&config.RollbackRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_release_set_revision": 4}},
		{&config.OfflineRequest{}, map[protoreflect.Name]protoreflect.FieldNumber{"expected_release_set_revision": 2}},
	}
	for _, item := range cases {
		descriptor := item.message.ProtoReflect().Descriptor()
		for name, number := range item.fields {
			field := descriptor.Fields().ByName(name)
			if field == nil || field.Number() != number || field.Kind() != protoreflect.Uint64Kind {
				t.Fatalf("%s.%s CAS contract changed: %v", descriptor.Name(), name, field)
			}
		}
	}
	parameter := (&config.Parameter{}).ProtoReflect().Descriptor()
	if parameter.Fields().Len() != 3 || parameter.Fields().ByName("default_value") != nil {
		t.Fatal("Parameter must not introduce a second default source")
	}
}

func TestConfigDescriptorsKeepManagementOutOfRuntime(t *testing.T) {
	file := config.File_config_proto
	if file.Package() != "pole.config.v1" || file.Services().Len() != 0 {
		t.Fatal("config must only declare domain messages")
	}
	for _, name := range []protoreflect.Name{"Error", "StateResponse", "PreviewResponse", "ResolvedConfig"} {
		if file.Messages().ByName(name) != nil {
			t.Fatalf("duplicate response contract: %s", name)
		}
	}
	if file.Enums().ByName("ErrorCode") != nil {
		t.Fatal("use common model.Code")
	}
	result := (&config.ConfigResult{}).ProtoReflect().Descriptor()
	expected := []protoreflect.Name{"presence", "key", "release_id", "snapshot_revision", "snapshot", "observation_cursor"}
	if result.Fields().Len() != len(expected) {
		t.Fatal("unexpected runtime envelope fields")
	}
	for i, name := range expected {
		if result.Fields().Get(i).Name() != name {
			t.Fatalf("unexpected field %s", result.Fields().Get(i).Name())
		}
	}
	snapshot := (&config.RenderSnapshot{}).ProtoReflect().Descriptor()
	for _, name := range []protoreflect.Name{"value_rules", "default_values", "compatibility", "release_id", "content_sha256", "expected_rendered_sha256", "error", "code", "info"} {
		if snapshot.Fields().ByName(name) != nil {
			t.Fatalf("snapshot leaks other audiences or response metadata: %s", name)
		}
	}
	if snapshot.Fields().ByName("values").MapValue().Message().FullName() != "pole.config.v1.Scalar" {
		t.Fatal("snapshot requires typed effective values")
	}
	if file.Enums().ByName("ConfigPresence").Values().Len() != 3 {
		t.Fatal("presence must not duplicate request errors")
	}
}

func TestConfigPayloadsUseCommonResponse(t *testing.T) {
	for _, payload := range []proto.Message{
		&config.ConfigResult{Presence: config.ConfigPresence_CONFIG_PRESENCE_PRESENT, Snapshot: &config.RenderSnapshot{Content: []byte("{{{name}}}"), Engine: "pole-mustache-v1", Values: map[string]*config.Scalar{"name": {Type: config.ValueType_VALUE_TYPE_STRING, Value: "SDK"}}}},
		&config.ConfigResult{Presence: config.ConfigPresence_CONFIG_PRESENCE_ABSENT},
		&config.RenderSnapshot{Engine: "literal", Content: []byte{0, 255}},
		&config.ConfigState{NextPageToken: "next", ActiveReleaseIds: []string{"gray"}},
	} {
		data, err := anypb.New(payload)
		if err != nil {
			t.Fatal(err)
		}
		response := &model.Response{Code: uint32(model.Code_ExecuteSuccess), Info: "success", Data: data}
		document, err := protojson.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		var decoded model.Response
		if err := protojson.Unmarshal(document, &decoded); err != nil {
			t.Fatal(err)
		}
		restored, err := decoded.Data.UnmarshalNew()
		if err != nil || !proto.Equal(restored, payload) || decoded.Code != uint32(model.Code_ExecuteSuccess) {
			t.Fatalf("common response lost payload: %v", err)
		}
	}
	for _, code := range []model.Code{model.Code_InvalidParameter, model.Code_DataConflict, model.Code_AuthTokenForbidden, model.Code_ConfigCursorReset, model.Code_ConfigRenderFailed, model.Code_ConfigUnavailable} {
		response := &model.Response{Code: uint32(code), Info: "failure"}
		wire, err := proto.Marshal(response)
		if err != nil {
			t.Fatal(err)
		}
		var decoded model.Response
		if err := proto.Unmarshal(wire, &decoded); err != nil || decoded.Data != nil || decoded.Code != uint32(code) {
			t.Fatal("common failure must not carry a successful result")
		}
	}
}

func TestConfigCompatibilityRequiresExplicitProtocolOptIn(t *testing.T) {
	var definition config.ConfigDefinition
	if definition.GetCompatibility().GetApollo() || definition.GetCompatibility().GetNacos() || definition.GetCompatibility().GetConsul() {
		t.Fatal("compatibility must default off")
	}
	definition.Compatibility = &config.ProtocolCompatibility{Nacos: true}
	wire, err := proto.Marshal(&definition)
	if err != nil {
		t.Fatal(err)
	}
	var decoded config.ConfigDefinition
	if err := proto.Unmarshal(wire, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.GetCompatibility().GetNacos() || decoded.GetCompatibility().GetApollo() || decoded.GetCompatibility().GetConsul() {
		t.Fatal("enabling one protocol must not enable the others")
	}
}

func TestConfigGRPCOnlyExposesClientMethods(t *testing.T) {
	var _ interface {
		Send(*config.DiscoverRequest) error
		Recv() (*model.Response, error)
	} = (configmanage.ConfigGRPC_DiscoverClient)(nil)
	var _ interface {
		Recv() (*config.DiscoverRequest, error)
		Send(*model.Response) error
	} = (configmanage.ConfigGRPC_DiscoverServer)(nil)
	file := configmanage.File_grpc_config_api_proto
	service := file.Services().ByName("ConfigGRPC")
	if file.Services().Len() != 1 || service == nil || service.FullName() != "v1.ConfigGRPC" || service.Methods().Len() != 1 {
		t.Fatal("configuration gRPC must have exactly one Discover method")
	}
	want := map[protoreflect.Name]struct {
		input, output protoreflect.FullName
		stream        bool
	}{
		"Discover": {"pole.config.v1.DiscoverRequest", "v1.Response", true},
	}
	for name, expected := range want {
		method := service.Methods().ByName(name)
		if method == nil || method.Input().FullName() != expected.input || method.Output().FullName() != expected.output ||
			method.IsStreamingClient() != expected.stream || method.IsStreamingServer() != expected.stream {
			t.Fatalf("RPC %s descriptor mismatch: %v", name, method)
		}
	}
	for _, name := range []protoreflect.Name{"CreateConfigFile", "UpdateConfigFile", "PublishConfigFile", "PreviewConfigTemplate", "SaveDraft", "GetState", "Preview", "Publish", "Stop", "StopAllGray", "CopyReleaseToDraft", "Rollback", "Offline", "Resolve", "Watch"} {
		if service.Methods().ByName(name) != nil {
			t.Fatalf("additional RPC %s must not be exposed", name)
		}
	}
	for _, name := range []protoreflect.Name{"ResolveRequest", "WatchRequest"} {
		if config.File_config_proto.Messages().ByName(name) != nil {
			t.Fatalf("additional client message %s must not be exposed", name)
		}
	}
}

func TestConfigPagedStateFieldsKeepWireTags(t *testing.T) {
	requestWire := []byte{0x10, 0x14, 0x1a, 0x01, 'p', 0x22, 0x01, 'r'}
	var request config.GetStateRequest
	if err := proto.Unmarshal(requestWire, &request); err != nil {
		t.Fatal(err)
	}
	if request.PageSize != 20 || request.PageToken != "p" || request.ReleaseId != "r" {
		t.Fatalf("paged request field tags changed: %v", &request)
	}
	responseWire := []byte{0x3a, 0x01, 'p', 0x42, 0x03, 0x0a, 0x01, 'r', 0x4a, 0x01, 'a'}
	var response config.ConfigState
	if err := proto.Unmarshal(responseWire, &response); err != nil {
		t.Fatal(err)
	}
	if response.NextPageToken != "p" || response.GetSelectedRelease().GetId() != "r" ||
		len(response.ActiveReleaseIds) != 1 || response.ActiveReleaseIds[0] != "a" {
		t.Fatalf("paged response field tags changed: %v", &response)
	}
	for _, item := range []struct {
		message proto.Message
		wire    []byte
	}{{&request, requestWire}, {&response, responseWire}} {
		wire, err := proto.Marshal(item.message)
		if err != nil || !bytes.Equal(wire, item.wire) {
			t.Fatalf("fixed pagination vector changed: %x (%v)", wire, err)
		}
	}
	if got := request.ProtoReflect().Descriptor().Fields().ByName("page_size").Kind(); got != protoreflect.Uint32Kind {
		t.Fatalf("page_size type = %v", got)
	}
	active := response.ProtoReflect().Descriptor().Fields().ByName("active_release_ids")
	if active.Cardinality() != protoreflect.Repeated || active.Kind() != protoreflect.StringKind {
		t.Fatal("active IDs must remain repeated strings")
	}
}

func TestConfigMetadataPageAndSelectedDefinitionRemainSeparate(t *testing.T) {
	response := &config.ConfigState{
		Draft:         &config.ConfigDraft{Definition: &config.ConfigDefinition{Content: []byte{0, 255}, Engine: "literal"}},
		Releases:      []*config.ReleaseSummary{{Id: "history-2", Sequence: 2}},
		NextPageToken: "opaque-page",
		SelectedRelease: &config.ConfigRelease{Id: "history-1", Sequence: 1,
			Definition: &config.ConfigDefinition{Content: []byte{128, 0, 255}, Engine: "literal"}},
		ActiveReleaseIds: []string{"active-normal", "active-gray"},
	}
	wire, err := proto.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var decoded config.ConfigState
	if err := proto.Unmarshal(wire, &decoded); err != nil || !proto.Equal(response, &decoded) {
		t.Fatalf("bounded management projection wire failed: %v", err)
	}
	if decoded.Releases[0].ProtoReflect().Descriptor().Fields().ByName("definition") != nil || decoded.GetSelectedRelease().GetDefinition() == nil ||
		len(decoded.ActiveReleaseIds) != 2 {
		t.Fatal("metadata page, selected definition and full active IDs must remain independent")
	}
}

func TestConfigAudienceConditionsWire(t *testing.T) {
	// labels 保留 tag 1；conditions 使用 tag 2，条件字段使用 1/2/3。
	wire := []byte{0x12, 0x0b, 0x0a, 0x01, 'k', 0x10, 0x06, 0x1a, 0x04, '1', '~', '1', '0'}
	var audience config.Audience
	if err := proto.Unmarshal(wire, &audience); err != nil {
		t.Fatal(err)
	}
	if len(audience.Conditions) != 1 || audience.Conditions[0].Key != "k" || audience.Conditions[0].Type != config.MatchType_MATCH_TYPE_RANGE || audience.Conditions[0].Value != "1~10" {
		t.Fatalf("condition wire mismatch: %v", &audience)
	}
	encoded, err := proto.Marshal(&audience)
	if err != nil || !bytes.Equal(encoded, wire) {
		t.Fatalf("condition wire changed: %x (%v)", encoded, err)
	}
	legacy := []byte{0x0a, 0x06, 0x0a, 0x01, 'k', 0x12, 0x01, 'v'}
	if err := proto.Unmarshal(legacy, &audience); err != nil || audience.Labels["k"] != "v" || len(audience.Conditions) != 0 {
		t.Fatalf("legacy labels changed: %v (%v)", &audience, err)
	}
	for number, name := range []string{"MATCH_TYPE_UNSPECIFIED", "MATCH_TYPE_EXACT", "MATCH_TYPE_REGEX", "MATCH_TYPE_NOT_EQUALS", "MATCH_TYPE_IN", "MATCH_TYPE_NOT_IN", "MATCH_TYPE_RANGE"} {
		if config.MatchType(number).String() != name {
			t.Fatalf("match type %d changed", number)
		}
	}
}
