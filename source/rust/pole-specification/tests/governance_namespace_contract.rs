use pole_specification::v1::{
    CircuitBreakerRule, FaultDetectRule, LaneGroup, LosslessRule, RateLimit, RouteRule,
    TrafficMirror, TrafficMock, TrafficSecurityRule,
};
use prost::Message;

macro_rules! assert_namespace_round_trip {
    ($message_type:ty, $field_number:expr) => {{
        let mut message = <$message_type>::default();
        message.namespace = "production".to_owned();

        let encoded = message.encode_to_vec();
        let expected_tag = (($field_number << 3) | 2) as u8;
        assert_eq!(encoded.first().copied(), Some(expected_tag));

        let decoded = <$message_type>::decode(encoded.as_slice()).expect("decode namespace");
        assert_eq!(decoded.namespace, "production");

        // Historical payloads only contain field 1 (id). Adding namespace on a
        // new field number must keep those payloads readable.
        let legacy = [
            0x0a, 0x09, b'l', b'e', b'g', b'a', b'c', b'y', b'-', b'i', b'd',
        ];
        let decoded = <$message_type>::decode(legacy.as_slice()).expect("decode legacy payload");
        assert_eq!(decoded.id, "legacy-id");
        assert!(decoded.namespace.is_empty());
    }};
}

#[test]
fn governance_roots_preserve_namespace_and_legacy_wire_compatibility() {
    assert_namespace_round_trip!(RouteRule, 15);
    assert_namespace_round_trip!(RateLimit, 15);
    assert_namespace_round_trip!(CircuitBreakerRule, 16);
    assert_namespace_round_trip!(FaultDetectRule, 13);
    assert_namespace_round_trip!(LaneGroup, 13);
    assert_namespace_round_trip!(LosslessRule, 10);
    assert_namespace_round_trip!(TrafficSecurityRule, 15);
    assert_namespace_round_trip!(TrafficMirror, 15);
    assert_namespace_round_trip!(TrafficMock, 15);
}
