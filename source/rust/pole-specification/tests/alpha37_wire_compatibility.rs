use pole_specification::v1::{
    Api, BlockConfig, DestinationService, FallbackResponse, LimitTrigger, ManagedCallerSelector,
    MirrorRule, MockResponse, MockRule, SourceService, TrafficMirror, TrafficMock,
    TrafficSecurityPolicy, TrafficSecurityRejectEffect,
};
use prost::Message;
use std::fmt::Debug;

fn assert_round_trip<M>(message: M, expected_first_tag: u8)
where
    M: Message + Default + PartialEq + Debug,
{
    let encoded = message.encode_to_vec();
    assert_eq!(encoded.first().copied(), Some(expected_first_tag));
    let decoded = M::decode(encoded.as_slice()).expect("decode alpha.37 payload");
    assert_eq!(decoded, message);
}

#[test]
fn alpha37_conflict_fields_keep_their_wire_tags() {
    assert_round_trip(
        TrafficMirror {
            caller: Some(SourceService {
                service: "caller".to_owned(),
                ..Default::default()
            }),
            ..Default::default()
        },
        0x22,
    );
    assert_round_trip(
        TrafficMirror {
            callee: Some(DestinationService {
                service: "callee".to_owned(),
                ..Default::default()
            }),
            ..Default::default()
        },
        0x2a,
    );
    assert_round_trip(
        TrafficMock {
            caller: Some(SourceService {
                service: "caller".to_owned(),
                ..Default::default()
            }),
            ..Default::default()
        },
        0x22,
    );
    assert_round_trip(
        TrafficMock {
            callee: Some(DestinationService {
                service: "callee".to_owned(),
                ..Default::default()
            }),
            ..Default::default()
        },
        0x2a,
    );
    assert_round_trip(
        MirrorRule {
            apis: vec![Api::default()],
            ..Default::default()
        },
        0x0a,
    );
    assert_round_trip(
        MirrorRule {
            duration: Some(prost_types::Duration {
                seconds: 30,
                nanos: 0,
            }),
            ..Default::default()
        },
        0x2a,
    );
    assert_round_trip(
        MockRule {
            delay: Some(prost_types::Duration {
                seconds: 1,
                nanos: 0,
            }),
            ..Default::default()
        },
        0x2a,
    );
    assert_round_trip(
        LimitTrigger {
            apis: vec![Api::default()],
            ..Default::default()
        },
        0x12,
    );
    assert_round_trip(
        BlockConfig {
            regex_separate: true,
            ..Default::default()
        },
        0x28,
    );
    assert_round_trip(
        TrafficSecurityPolicy {
            managed_caller: Some(ManagedCallerSelector {
                any_authenticated: true,
                ..Default::default()
            }),
            ..Default::default()
        },
        0x2a,
    );
    assert_round_trip(
        MockResponse {
            code: "MOCKED".to_owned(),
            ..Default::default()
        },
        0x0a,
    );
    assert_round_trip(
        TrafficSecurityRejectEffect {
            code: "DENIED".to_owned(),
            ..Default::default()
        },
        0x0a,
    );
    assert_round_trip(
        FallbackResponse {
            code: "FALLBACK".to_owned(),
            ..Default::default()
        },
        0x0a,
    );
}
