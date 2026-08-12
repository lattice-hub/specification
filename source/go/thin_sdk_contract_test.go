package specification_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

type contractHeader [2]string

type contractEnvelope struct {
	Namespace        string `json:"namespace"`
	Service          string `json:"service"`
	Protocol         string `json:"protocol"`
	Group            string `json:"group"`
	ServiceVersion   string `json:"service_version"`
	Method           string `json:"method"`
	OriginalEndpoint string `json:"original_endpoint"`
}

type contractVector struct {
	Name            string            `json:"name"`
	Input           contractEnvelope  `json:"input"`
	BaseHeaders     []contractHeader  `json:"base_headers"`
	Normalized      map[string]string `json:"normalized"`
	ExpectedHeaders []contractHeader  `json:"expected_headers"`
	Diagnostic      string            `json:"diagnostic"`
}

type receiveVector struct {
	Name             string            `json:"name"`
	Headers          []contractHeader  `json:"headers"`
	ExpectedEnvelope map[string]string `json:"expected_envelope"`
	Diagnostic       string            `json:"diagnostic"`
}

func TestThinSDKTargetEnvelopeConformance(t *testing.T) {
	content, err := os.ReadFile("../../thin-sdk/target-envelope/v1/conformance.json")
	if err != nil {
		t.Fatalf("read TargetEnvelope vectors: %v", err)
	}
	var vectors struct {
		Contract                string           `json:"contract"`
		ContractVersion         string           `json:"contract_version"`
		EnvelopeVersion         string           `json:"envelope_version"`
		Valid                   []contractVector `json:"valid"`
		Invalid                 []contractVector `json:"invalid"`
		LanguageSpecificInvalid []struct {
			Name           string   `json:"name"`
			Field          string   `json:"field"`
			UTF16CodeUnits []string `json:"utf16_code_units"`
			Diagnostic     string   `json:"diagnostic"`
		} `json:"language_specific_invalid"`
		SidecarReceive struct {
			Valid   []receiveVector `json:"valid"`
			Invalid []receiveVector `json:"invalid"`
		} `json:"sidecar_receive"`
	}
	if err := json.Unmarshal(content, &vectors); err != nil {
		t.Fatalf("decode TargetEnvelope vectors: %v", err)
	}
	if vectors.Contract != "pole-target-envelope" ||
		vectors.ContractVersion != "1.0.0" ||
		vectors.EnvelopeVersion != "1" {
		t.Fatalf("unexpected TargetEnvelope identity: %+v", vectors)
	}
	if len(vectors.Valid) != 9 || len(vectors.Invalid) != 20 ||
		len(vectors.LanguageSpecificInvalid) != 2 ||
		len(vectors.SidecarReceive.Valid) != 2 ||
		len(vectors.SidecarReceive.Invalid) != 8 {
		t.Fatalf("unexpected vector counts")
	}

	names := map[string]struct{}{}
	for _, vector := range vectors.Valid {
		assertUniqueVectorName(t, names, vector.Name)
		normalized, diagnostic := normalizeContractEnvelope(vector.Input)
		if diagnostic != "" {
			t.Fatalf("valid vector %q returned %s", vector.Name, diagnostic)
		}
		if !reflect.DeepEqual(normalizedEnvelopeMap(normalized), vector.Normalized) {
			t.Fatalf("valid vector %q normalized = %#v, want %#v",
				vector.Name, normalizedEnvelopeMap(normalized), vector.Normalized)
		}
		headers := encodeContractHeaders(vector.BaseHeaders, normalized)
		if !reflect.DeepEqual(headers, vector.ExpectedHeaders) {
			t.Fatalf("valid vector %q headers = %#v, want %#v",
				vector.Name, headers, vector.ExpectedHeaders)
		}
	}
	for _, vector := range vectors.Invalid {
		assertUniqueVectorName(t, names, vector.Name)
		_, diagnostic := normalizeContractEnvelope(vector.Input)
		if diagnostic != vector.Diagnostic {
			t.Fatalf("invalid vector %q diagnostic = %q, want %q",
				vector.Name, diagnostic, vector.Diagnostic)
		}
	}
	for _, vector := range vectors.LanguageSpecificInvalid {
		assertUniqueVectorName(t, names, vector.Name)
		if vector.Field != "service" || len(vector.UTF16CodeUnits) != 1 ||
			vector.Diagnostic != "INVALID_UNICODE_SCALAR" {
			t.Fatalf("unexpected language-specific vector: %+v", vector)
		}
		raw := "\xed\xa0\x80"
		if vector.UTF16CodeUnits[0] == "DC00" {
			raw = "\xed\xb0\x80"
		}
		_, diagnostic := normalizeContractEnvelope(contractEnvelope{
			Namespace: "default",
			Service:   raw,
		})
		if diagnostic != vector.Diagnostic {
			t.Fatalf("language-specific vector %q diagnostic = %q, want %q",
				vector.Name, diagnostic, vector.Diagnostic)
		}
	}
	for _, vector := range vectors.SidecarReceive.Valid {
		assertUniqueVectorName(t, names, vector.Name)
		envelope, diagnostic := validateReceivedHeaders(vector.Headers)
		if diagnostic != "" {
			t.Fatalf("valid receive vector %q returned %s", vector.Name, diagnostic)
		}
		if got := normalizedEnvelopeMap(envelope); !reflect.DeepEqual(got, vector.ExpectedEnvelope) {
			t.Fatalf("valid receive vector %q envelope = %#v, want %#v",
				vector.Name, got, vector.ExpectedEnvelope)
		}
	}
	for _, vector := range vectors.SidecarReceive.Invalid {
		assertUniqueVectorName(t, names, vector.Name)
		_, diagnostic := validateReceivedHeaders(vector.Headers)
		if diagnostic != vector.Diagnostic {
			t.Fatalf("invalid receive vector %q diagnostic = %q, want %q",
				vector.Name, diagnostic, vector.Diagnostic)
		}
	}
}

func TestThinSDKTargetEnvelopeSchemaIdentity(t *testing.T) {
	content, err := os.ReadFile("../../thin-sdk/target-envelope/v1/schema.json")
	if err != nil {
		t.Fatalf("read TargetEnvelope schema: %v", err)
	}
	var schema struct {
		Schema               string                 `json:"$schema"`
		ID                   string                 `json:"$id"`
		AdditionalProperties bool                   `json:"additionalProperties"`
		Required             []string               `json:"required"`
		Properties           map[string]interface{} `json:"properties"`
	}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("decode TargetEnvelope schema: %v", err)
	}
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" ||
		schema.ID != "urn:pole:thin-sdk:target-envelope:1.0.0:schema" ||
		schema.AdditionalProperties ||
		!reflect.DeepEqual(schema.Required, []string{"namespace", "service"}) ||
		len(schema.Properties) != 7 {
		t.Fatalf("unexpected TargetEnvelope schema summary: %+v", schema)
	}
}

func TestThinSDKContractChecksums(t *testing.T) {
	const assetDirectory = "../../thin-sdk/target-envelope/v1"
	content, err := os.ReadFile(filepath.Join(assetDirectory, "SHA256SUMS"))
	if err != nil {
		t.Fatalf("read SHA256SUMS: %v", err)
	}
	expectedFiles := map[string]bool{
		"schema.json":      false,
		"conformance.json": false,
	}
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			t.Fatalf("invalid SHA256SUMS line %q", line)
		}
		seen, exists := expectedFiles[parts[1]]
		if !exists || seen {
			t.Fatalf("unexpected or duplicate SHA256SUMS asset %q", parts[1])
		}
		expectedFiles[parts[1]] = true
		asset, err := os.ReadFile(filepath.Join(assetDirectory, parts[1]))
		if err != nil {
			t.Fatalf("read checksummed asset %q: %v", parts[1], err)
		}
		actual := fmt.Sprintf("%x", sha256.Sum256(asset))
		if actual != parts[0] {
			t.Fatalf("%s checksum = %s, want %s", parts[1], actual, parts[0])
		}
	}
	for name, seen := range expectedFiles {
		if !seen {
			t.Fatalf("SHA256SUMS is missing %s", name)
		}
	}
}

func TestThinSDKCompatibilityStartsUnverified(t *testing.T) {
	content, err := os.ReadFile("../../thin-sdk/compatibility.json")
	if err != nil {
		t.Fatalf("read Thin SDK compatibility: %v", err)
	}
	type evidence struct {
		SpecificationCommit string `json:"specification_commit"`
		SidecarCommit       string `json:"sidecar_commit"`
		SDKCommit           string `json:"sdk_commit"`
		CIURL               string `json:"ci_url"`
	}
	var document struct {
		SchemaVersion int `json:"schema_version"`
		Contracts     []struct {
			Name                 string `json:"name"`
			ContractVersion      string `json:"contract_version"`
			EnvelopeVersion      string `json:"envelope_version"`
			Status               string `json:"status"`
			VerificationPolicy   string `json:"verification_policy"`
			VerifiedCombinations []struct {
				SidecarVersion string   `json:"sidecar_version"`
				SDKLanguage    string   `json:"sdk_language"`
				SDKVersion     string   `json:"sdk_version"`
				Evidence       evidence `json:"evidence"`
			} `json:"verified_combinations"`
		} `json:"contracts"`
	}
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatalf("decode Thin SDK compatibility: %v", err)
	}
	if document.SchemaVersion != 1 || len(document.Contracts) != 3 {
		t.Fatalf("unexpected compatibility document: %+v", document)
	}
	contract := document.Contracts[0]
	if contract.Name != "pole-target-envelope" ||
		contract.ContractVersion != "1.0.0" ||
		contract.EnvelopeVersion != "1" ||
		contract.Status != "defined" ||
		contract.VerificationPolicy == "" ||
		contract.VerifiedCombinations == nil ||
		len(contract.VerifiedCombinations) != 0 {
		t.Fatalf("unexpected initial TargetEnvelope compatibility: %+v", contract)
	}
	contract = document.Contracts[1]
	if contract.Name != "latticehub-thin-sdk-sidecar" ||
		contract.ContractVersion != "3.0.0" ||
		contract.EnvelopeVersion != "1" ||
		contract.Status != "defined" ||
		contract.VerificationPolicy == "" ||
		contract.VerifiedCombinations == nil ||
		len(contract.VerifiedCombinations) != 0 {
		t.Fatalf("unexpected initial Thin SDK v2 compatibility: %+v", contract)
	}
	contract = document.Contracts[2]
	if contract.Name != "latticehub-traffic-context" ||
		contract.ContractVersion != "1.0.0" ||
		contract.EnvelopeVersion != "1" ||
		contract.Status != "defined" ||
		contract.VerificationPolicy == "" ||
		contract.VerifiedCombinations == nil ||
		len(contract.VerifiedCombinations) != 0 {
		t.Fatalf("unexpected initial TrafficContext compatibility: %+v", contract)
	}
}

func TestThinSDKCompatibilitySchemaIdentity(t *testing.T) {
	content, err := os.ReadFile("../../thin-sdk/compatibility.schema.json")
	if err != nil {
		t.Fatalf("read compatibility schema: %v", err)
	}
	var schema struct {
		Schema               string                 `json:"$schema"`
		ID                   string                 `json:"$id"`
		AdditionalProperties bool                   `json:"additionalProperties"`
		Required             []string               `json:"required"`
		Definitions          map[string]interface{} `json:"$defs"`
	}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("decode compatibility schema: %v", err)
	}
	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" ||
		schema.ID != "urn:pole:thin-sdk:compatibility:1:schema" ||
		schema.AdditionalProperties ||
		!reflect.DeepEqual(schema.Required, []string{"schema_version", "contracts"}) ||
		schema.Definitions["verifiedCombination"] == nil {
		t.Fatalf("unexpected compatibility schema: %+v", schema)
	}
}

func assertUniqueVectorName(t *testing.T, names map[string]struct{}, name string) {
	t.Helper()
	if _, exists := names[name]; name == "" || exists {
		t.Fatalf("invalid or duplicate vector name %q", name)
	}
	names[name] = struct{}{}
}

func normalizeContractEnvelope(input contractEnvelope) (contractEnvelope, string) {
	fields := []struct {
		name     string
		value    *string
		required bool
	}{
		{name: "namespace", value: &input.Namespace, required: true},
		{name: "service", value: &input.Service, required: true},
		{name: "protocol", value: &input.Protocol},
		{name: "group", value: &input.Group},
		{name: "service_version", value: &input.ServiceVersion},
		{name: "method", value: &input.Method},
		{name: "original_endpoint", value: &input.OriginalEndpoint},
	}
	for _, field := range fields {
		if !utf8.ValidString(*field.value) {
			return contractEnvelope{}, "INVALID_UNICODE_SCALAR"
		}
		for _, character := range *field.value {
			if unicode.Is(unicode.Cc, character) {
				return contractEnvelope{}, "CONTROL_CHARACTER"
			}
		}
		*field.value = strings.TrimFunc(*field.value, isContractWhitespace)
		if field.required && *field.value == "" {
			return contractEnvelope{}, "REQUIRED_FIELD_EMPTY"
		}
	}
	if input.OriginalEndpoint != "" && !validContractEndpoint(input.OriginalEndpoint) {
		return contractEnvelope{}, "INVALID_ORIGINAL_ENDPOINT"
	}
	return input, ""
}

func isContractWhitespace(character rune) bool {
	return character == '\u0020' ||
		character == '\u00a0' ||
		character == '\u1680' ||
		character >= '\u2000' && character <= '\u200a' ||
		character == '\u2028' ||
		character == '\u2029' ||
		character == '\u202f' ||
		character == '\u205f' ||
		character == '\u3000'
}

func validContractEndpoint(endpoint string) bool {
	var host string
	var port string
	if strings.HasPrefix(endpoint, "[") {
		closingBracket := strings.LastIndex(endpoint, "]")
		if closingBracket <= 1 ||
			closingBracket+1 >= len(endpoint) ||
			endpoint[closingBracket+1] != ':' {
			return false
		}
		host = endpoint[1:closingBracket]
		if strings.Contains(host, "%") {
			return false
		}
		address, err := netip.ParseAddr(host)
		if err != nil || !address.Is6() {
			return false
		}
		port = endpoint[closingBracket+2:]
	} else {
		if strings.Count(endpoint, ":") != 1 {
			return false
		}
		host, port, _ = strings.Cut(endpoint, ":")
		if host == "" {
			return false
		}
		for _, character := range host {
			if isContractWhitespace(character) || strings.ContainsRune("/\\[]@?#%", character) {
				return false
			}
		}
	}
	if port == "" {
		return false
	}
	if len(port) > 1 && port[0] == '0' {
		return false
	}
	for _, character := range port {
		if character < '0' || character > '9' {
			return false
		}
	}
	value, err := strconv.Atoi(port)
	return err == nil && value >= 1 && value <= 65535
}

func normalizedEnvelopeMap(envelope contractEnvelope) map[string]string {
	result := map[string]string{
		"namespace": envelope.Namespace,
		"service":   envelope.Service,
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "protocol", value: envelope.Protocol},
		{name: "group", value: envelope.Group},
		{name: "service_version", value: envelope.ServiceVersion},
		{name: "method", value: envelope.Method},
		{name: "original_endpoint", value: envelope.OriginalEndpoint},
	} {
		if field.value != "" {
			result[field.name] = field.value
		}
	}
	return result
}

var contractHeaderFields = []struct {
	name  string
	value func(contractEnvelope) string
}{
	{name: "x-pole-target-envelope-version", value: func(contractEnvelope) string { return "1" }},
	{name: "x-pole-target-namespace", value: func(envelope contractEnvelope) string { return envelope.Namespace }},
	{name: "x-pole-target-service", value: func(envelope contractEnvelope) string { return envelope.Service }},
	{name: "x-pole-target-protocol", value: func(envelope contractEnvelope) string { return envelope.Protocol }},
	{name: "x-pole-target-group", value: func(envelope contractEnvelope) string { return envelope.Group }},
	{name: "x-pole-target-service-version", value: func(envelope contractEnvelope) string { return envelope.ServiceVersion }},
	{name: "x-pole-target-method", value: func(envelope contractEnvelope) string { return envelope.Method }},
	{name: "x-pole-original-endpoint", value: func(envelope contractEnvelope) string { return envelope.OriginalEndpoint }},
}

func encodeContractHeaders(base []contractHeader, envelope contractEnvelope) []contractHeader {
	internal := map[string]bool{}
	for _, field := range contractHeaderFields {
		internal[field.name] = true
	}
	result := make([]contractHeader, 0, len(base)+len(contractHeaderFields))
	for _, header := range base {
		if !internal[strings.ToLower(header[0])] {
			result = append(result, header)
		}
	}
	for _, field := range contractHeaderFields {
		value := field.value(envelope)
		if value != "" {
			result = append(result, contractHeader{field.name, encodeContractValue(value)})
		}
	}
	return result
}

func encodeContractValue(value string) string {
	var encoded strings.Builder
	for _, valueByte := range []byte(value) {
		if valueByte >= 0x20 && valueByte <= 0x7e && valueByte != '%' && valueByte != ',' {
			encoded.WriteByte(valueByte)
		} else {
			fmt.Fprintf(&encoded, "%%%02X", valueByte)
		}
	}
	return encoded.String()
}

func validateReceivedHeaders(headers []contractHeader) (contractEnvelope, string) {
	known := map[string]bool{}
	for _, field := range contractHeaderFields {
		known[field.name] = true
	}
	values := map[string]string{}
	for _, header := range headers {
		name := strings.ToLower(header[0])
		if strings.HasPrefix(name, "x-pole-target-") && !known[name] {
			return contractEnvelope{}, "UNKNOWN_INTERNAL_HEADER"
		}
		if !known[name] {
			continue
		}
		if _, exists := values[name]; exists {
			return contractEnvelope{}, "DUPLICATE_INTERNAL_HEADER"
		}
		decoded, ok := decodeCanonicalContractValue(header[1])
		if !ok {
			return contractEnvelope{}, "NON_CANONICAL_HEADER_VALUE"
		}
		values[name] = decoded
	}
	for _, name := range []string{
		"x-pole-target-envelope-version",
		"x-pole-target-namespace",
		"x-pole-target-service",
	} {
		if _, exists := values[name]; !exists {
			return contractEnvelope{}, "MISSING_REQUIRED_HEADER"
		}
	}
	if values["x-pole-target-envelope-version"] != "1" {
		return contractEnvelope{}, "UNKNOWN_ENVELOPE_VERSION"
	}
	envelope := contractEnvelope{
		Namespace:        values["x-pole-target-namespace"],
		Service:          values["x-pole-target-service"],
		Protocol:         values["x-pole-target-protocol"],
		Group:            values["x-pole-target-group"],
		ServiceVersion:   values["x-pole-target-service-version"],
		Method:           values["x-pole-target-method"],
		OriginalEndpoint: values["x-pole-original-endpoint"],
	}
	normalized, diagnostic := normalizeContractEnvelope(envelope)
	if diagnostic != "" {
		return contractEnvelope{}, diagnostic
	}
	if normalized != envelope {
		return contractEnvelope{}, "NON_CANONICAL_HEADER_VALUE"
	}
	return envelope, ""
}

func decodeCanonicalContractValue(value string) (string, bool) {
	var decoded []byte
	for index := 0; index < len(value); {
		valueByte := value[index]
		if valueByte < 0x20 || valueByte > 0x7e || valueByte == ',' {
			return "", false
		}
		if valueByte != '%' {
			decoded = append(decoded, valueByte)
			index++
			continue
		}
		if index+2 >= len(value) ||
			!isUpperHex(value[index+1]) ||
			!isUpperHex(value[index+2]) {
			return "", false
		}
		parsed, err := strconv.ParseUint(value[index+1:index+3], 16, 8)
		if err != nil {
			return "", false
		}
		decoded = append(decoded, byte(parsed))
		index += 3
	}
	if !utf8.Valid(decoded) {
		return "", false
	}
	result := string(decoded)
	return result, encodeContractValue(result) == value
}

func isUpperHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'A' && value <= 'F'
}
