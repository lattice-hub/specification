use pole_specification::v1::{
    config_template_value, ConfigDiscoverFilter, ConfigDiscoverResponse, ConfigTemplateBinding,
    ConfigTemplateEngine, ConfigTemplateRelease, ConfigTemplateRenderInput, ConfigTemplateValue,
    NamespaceTemplateValueRelease, NamespaceTemplateValueReleaseType, RenderPreview,
    RenderPreviewRequest, RenderSnapshot,
};
use prost::Message;
use std::collections::HashMap;

#[test]
fn shared_vectors_are_packaged_for_every_sdk() {
    let vectors = include_str!("../../../../CONFIG_TEMPLATE_TEST_VECTORS.json");
    assert!(vectors.contains("\"profile\": \"pole-mustache-v1\""));
    assert!(vectors.contains("\"name\": \"typed-scalars\""));
    assert!(vectors.contains("\"name\": \"missing-required-value\""));
    assert!(vectors.contains("\"expected_sha256\""));
}

#[test]
fn render_snapshot_round_trips_typed_value_in_discover_response() {
    let values = HashMap::from([(
        "server.port".to_owned(),
        ConfigTemplateValue {
            value: Some(config_template_value::Value::IntegerValue(8080)),
        },
    )]);
    let snapshot = RenderSnapshot {
        template_binding: Some(ConfigTemplateBinding {
            template_id: 42,
            template_release_id: "template-release-7".to_owned(),
            binding_release_id: "binding-release-3".to_owned(),
        }),
        template_release: Some(ConfigTemplateRelease {
            id: "template-release-7".to_owned(),
            template_id: 42,
            content: "port: {{{server.port}}}\n".to_owned(),
            format: "yaml".to_owned(),
            engine: Some(ConfigTemplateEngine {
                name: "pole-mustache".to_owned(),
                version: "v1".to_owned(),
            }),
            ..Default::default()
        }),
        value_release: Some(NamespaceTemplateValueRelease {
            id: "value-release-9".to_owned(),
            namespace: "production".to_owned(),
            template_id: 42,
            template_release_id: "template-release-7".to_owned(),
            values,
            release_type: NamespaceTemplateValueReleaseType::TemplateValueReleaseGray as i32,
            priority: 100,
            active: true,
            version: 9,
            ..Default::default()
        }),
        revision: "combined-revision".to_owned(),
        expected_rendered_sha256: "reference-sha256".to_owned(),
    };
    let response = ConfigDiscoverResponse {
        revision: snapshot.revision.clone(),
        render_snapshot: Some(snapshot),
        ..Default::default()
    };

    let encoded = response.encode_to_vec();
    let decoded = ConfigDiscoverResponse::decode(encoded.as_slice()).expect("decode response");
    let decoded_snapshot = decoded.render_snapshot.expect("render snapshot");
    let decoded_value = decoded_snapshot
        .value_release
        .expect("value release")
        .values
        .remove("server.port")
        .expect("server.port");
    assert_eq!(
        decoded_value.value,
        Some(config_template_value::Value::IntegerValue(8080))
    );
    assert_eq!(
        decoded_snapshot.expected_rendered_sha256,
        "reference-sha256"
    );
}

#[test]
fn client_capability_and_preview_have_stable_wire_tags() {
    let capability = ConfigDiscoverFilter {
        supported_template_engines: vec![ConfigTemplateEngine {
            name: "pole-mustache".to_owned(),
            version: "v1".to_owned(),
        }],
        ..Default::default()
    };
    let encoded = capability.encode_to_vec();
    assert_eq!(encoded.first().copied(), Some((3 << 3 | 2) as u8));

    let preview = RenderPreview {
        rendered_content: "port: 8080\n".to_owned(),
        format: "yaml".to_owned(),
        template_release_id: "template-release-7".to_owned(),
        value_release_id: "value-release-9".to_owned(),
        engine: Some(ConfigTemplateEngine {
            name: "pole-mustache".to_owned(),
            version: "v1".to_owned(),
        }),
        rendered_sha256: "reference-sha256".to_owned(),
        valid: true,
        code: 200000,
        info: "ExecuteSuccess".to_owned(),
        ..Default::default()
    };
    let decoded =
        RenderPreview::decode(preview.encode_to_vec().as_slice()).expect("decode preview");
    assert!(decoded.valid);
    assert_eq!(decoded.rendered_content, "port: 8080\n");
    assert_eq!(decoded.code, 200000);

    let request = RenderPreviewRequest {
        input: Some(ConfigTemplateRenderInput {
            content: "port: {{{server.port}}}\n".to_owned(),
            format: "yaml".to_owned(),
            engine: Some(ConfigTemplateEngine {
                name: "pole-mustache".to_owned(),
                version: "v1".to_owned(),
            }),
            values: HashMap::from([(
                "server.port".to_owned(),
                ConfigTemplateValue {
                    value: Some(config_template_value::Value::IntegerValue(8080)),
                },
            )]),
            ..Default::default()
        }),
        ..Default::default()
    };
    let decoded_request =
        RenderPreviewRequest::decode(request.encode_to_vec().as_slice()).expect("decode request");
    assert_eq!(
        decoded_request.input.expect("draft input").content,
        "port: {{{server.port}}}\n"
    );
}
