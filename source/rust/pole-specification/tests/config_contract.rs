use pole_specification::pole::config::v1::{
    ConfigDefinition, ConfigDraft, ConfigPresence, ConfigRelease, ConfigResult, ConfigState,
    CopyReleaseToDraftRequest, Cursor, DiscoverRequest, GetStateRequest, ProtocolCompatibility,
    ReleaseKind, ReleaseSummary, RenderSnapshot, Scalar, ValueType,
};
use pole_specification::v1::config_grpc_client::ConfigGrpcClient;
use prost::Message;
use std::collections::HashMap;

#[test]
fn config_binary_and_uint64_share_go_wire_vector() {
    let wire = [
        0x08, 0x01, 0x2a, 0x05, 0x0a, 0x03, 0x00, 0xff, 0x80, 0x32, 0x0b, 0x18, 0xff, 0xff, 0xff,
        0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x01,
    ];
    let result = ConfigResult::decode(wire.as_slice()).expect("decode fixed Go/Rust wire");
    assert_eq!(result.presence, ConfigPresence::Present as i32);
    assert_eq!(result.snapshot.as_ref().unwrap().content, [0, 255, 128]);
    assert_eq!(
        result.observation_cursor.as_ref().unwrap().position,
        u64::MAX
    );
    assert_eq!(result.encode_to_vec(), wire);
}

#[test]
fn config_state_round_trips_exact_scalar_and_revisions() {
    let state = ConfigState {
        draft: Some(ConfigDraft {
            revision: u64::MAX,
            definition: Some(ConfigDefinition {
                content: b"amount: {{{amount}}}\r\n".to_vec(),
                engine: "pole-mustache-v1".into(),
                default_values: HashMap::from([(
                    "amount".into(),
                    Scalar {
                        r#type: ValueType::Decimal as i32,
                        value: "9007199254740993.0001".into(),
                    },
                )]),
                ..Default::default()
            }),
            ..Default::default()
        }),
        release_set_revision: u64::MAX,
        ..Default::default()
    };
    let decoded = ConfigState::decode(state.encode_to_vec().as_slice()).expect("decode state");
    assert_eq!(decoded, state);
    let release = ConfigRelease {
        sequence: u64::MAX,
        kind: ReleaseKind::Gray as i32,
        ..Default::default()
    };
    assert_eq!(
        ConfigRelease::decode(release.encode_to_vec().as_slice()).unwrap(),
        release
    );
}

#[test]
fn config_exports_rpc_client_and_explicit_statuses() {
    // 编译验证公开模块与 gRPC stub；不启动网络服务。
    let client: Option<ConfigGrpcClient<tonic::transport::Channel>> = None;
    if let Some(mut client) = client {
        drop(client.discover(tonic::codegen::tokio_stream::empty::<DiscoverRequest>()));
    }
    let _ = Cursor::default();
    for number in 0..=2 {
        let result = ConfigResult::decode([0x08, number].as_slice()).unwrap();
        assert_eq!(result.presence, i32::from(number));
        assert!(ConfigPresence::try_from(result.presence).is_ok());
    }
    assert_eq!(ReleaseKind::Unspecified as i32, 0);
    assert_eq!(ReleaseKind::Normal as i32, 1);
    assert_eq!(ReleaseKind::Gray as i32, 2);
    assert_ne!(
        ConfigPresence::Present as i32,
        ConfigPresence::Absent as i32
    );
}

#[test]
fn config_messages_and_client_service_have_separate_contracts() {
    let messages = include_str!("../proto/config.proto");
    assert!(!messages
        .lines()
        .any(|line| line.trim().starts_with("service ")));
    assert!(!messages.contains("message ResolveRequest"));
    assert!(!messages.contains("message WatchRequest"));
    let grpc = include_str!("../proto/grpc_config_api.proto");
    let methods: Vec<&str> = grpc
        .lines()
        .map(str::trim)
        .filter_map(|line| line.strip_prefix("rpc "))
        .map(|line| line.split('(').next().unwrap())
        .collect();
    assert_eq!(methods, ["Discover"]);
    assert!(grpc
        .contains("rpc Discover(stream pole.config.v1.DiscoverRequest) returns (stream Response)"));
    assert_eq!(
        grpc.lines()
            .filter(|line| line.trim().starts_with("service "))
            .count(),
        1
    );
}

#[test]
fn config_paged_state_shares_go_wire_tags() {
    let request_wire = [0x10, 0x14, 0x1a, 0x01, b'p', 0x22, 0x01, b'r'];
    let request = GetStateRequest::decode(request_wire.as_slice()).unwrap();
    assert_eq!(request.page_size, 20);
    assert_eq!(request.page_token, "p");
    assert_eq!(request.release_id, "r");
    assert_eq!(request.encode_to_vec(), request_wire);

    let response_wire = [
        0x3a, 0x01, b'p', 0x42, 0x03, 0x0a, 0x01, b'r', 0x4a, 0x01, b'a',
    ];
    let response = ConfigState::decode(response_wire.as_slice()).unwrap();
    assert_eq!(response.next_page_token, "p");
    assert_eq!(response.selected_release.as_ref().unwrap().id, "r");
    assert_eq!(response.active_release_ids, ["a"]);
    assert_eq!(response.encode_to_vec(), response_wire);
}

#[test]
fn scalar_shares_go_exact_wire_and_management_views_are_distinct() {
    let mut wire = vec![0x08, 0x04, 0x12, 0x15];
    wire.extend_from_slice(b"9007199254740993.0001");
    let scalar = Scalar::decode(wire.as_slice()).unwrap();
    assert_eq!(scalar.r#type, ValueType::Decimal as i32);
    assert_eq!(scalar.value, "9007199254740993.0001");
    assert_eq!(scalar.encode_to_vec(), wire);
    let preview = RenderSnapshot {
        content: vec![0, 255],
        engine: "literal".into(),
        ..Default::default()
    };
    assert_eq!(
        RenderSnapshot::decode(preview.encode_to_vec().as_slice()).unwrap(),
        preview
    );
    let summary = ReleaseSummary {
        id: "gray-1".into(),
        active: true,
        ..Default::default()
    };
    let state = ConfigState {
        releases: vec![summary],
        ..Default::default()
    };
    assert_eq!(
        ConfigState::decode(state.encode_to_vec().as_slice()).unwrap(),
        state
    );
    let _ = CopyReleaseToDraftRequest::default();
}

#[test]
fn config_uses_common_response_and_protocol_compatibility_defaults_off() {
    use pole_specification::v1::{Code, Response};
    let result = ConfigResult {
        presence: ConfigPresence::Present as i32,
        snapshot: Some(RenderSnapshot {
            content: b"{{{name}}}".to_vec(),
            engine: "pole-mustache-v1".into(),
            ..Default::default()
        }),
        ..Default::default()
    };
    let response = Response {
        code: Code::ExecuteSuccess as u32,
        data: Some(prost_types::Any {
            type_url: "type.googleapis.com/pole.config.v1.ConfigResult".into(),
            value: result.encode_to_vec(),
        }),
        ..Default::default()
    };
    let decoded = Response::decode(response.encode_to_vec().as_slice()).unwrap();
    assert_eq!(decoded.code, Code::ExecuteSuccess as u32);
    assert_eq!(
        ConfigResult::decode(decoded.data.unwrap().value.as_slice()).unwrap(),
        result
    );
    let disabled = ProtocolCompatibility::default();
    assert!(!disabled.apollo && !disabled.nacos && !disabled.consul);
    let definition = ConfigDefinition {
        compatibility: Some(ProtocolCompatibility {
            nacos: true,
            ..Default::default()
        }),
        ..Default::default()
    };
    let enabled = ConfigDefinition::decode(definition.encode_to_vec().as_slice())
        .unwrap()
        .compatibility
        .unwrap();
    assert!(enabled.nacos && !enabled.apollo && !enabled.consul);
    assert_eq!(Code::ConfigCursorReset as u32, 409020);
    assert_eq!(Code::ConfigRenderFailed as u32, 422020);
    assert_eq!(Code::ConfigUnavailable as u32, 503020);
}

#[test]
fn audience_conditions_share_go_wire_vector() {
    use pole_specification::pole::config::v1::{Audience, MatchType};
    let wire = [
        0x12, 0x0b, 0x0a, 0x01, b'k', 0x10, 0x06, 0x1a, 0x04, b'1', b'~', b'1', b'0',
    ];
    let audience = Audience::decode(wire.as_slice()).unwrap();
    assert_eq!(audience.conditions.len(), 1);
    assert_eq!(audience.conditions[0].key, "k");
    assert_eq!(audience.conditions[0].r#type, MatchType::Range as i32);
    assert_eq!(audience.conditions[0].value, "1~10");
    assert_eq!(audience.encode_to_vec(), wire);
}

#[test]
fn bindings_presence_and_promote_request_wire_round_trip() {
    use pole_specification::pole::config::v1::{ConfigKey, ParameterBindings, PromoteGrayRequest};
    let def = ConfigDefinition {
        parameter_bindings: vec![
            ParameterBindings {
                name: "empty".into(),
                fallback: Some(Scalar {
                    r#type: ValueType::String as i32,
                    value: String::new(),
                }),
                ..Default::default()
            },
            ParameterBindings {
                name: "missing".into(),
                ..Default::default()
            },
        ],
        ..Default::default()
    };
    let decoded = ConfigDefinition::decode(def.encode_to_vec().as_slice()).unwrap();
    assert_eq!(
        decoded.parameter_bindings[0]
            .fallback
            .as_ref()
            .unwrap()
            .value,
        ""
    );
    assert!(decoded.parameter_bindings[1].fallback.is_none());
    let request = PromoteGrayRequest {
        key: Some(ConfigKey {
            environment: "prod".into(),
            group: "g".into(),
            name: "n".into(),
        }),
        release_id: "r".into(),
        expected_release_set_revision: u64::MAX,
    };
    assert_eq!(
        PromoteGrayRequest::decode(request.encode_to_vec().as_slice()).unwrap(),
        request
    );
}
