use pole_specification::v1::{Namespace, NamespaceKind};
use prost::Message;

#[test]
fn namespace_kind_round_trips_and_treats_legacy_payload_as_business() {
    let namespace = Namespace {
        name: "pole-system".to_owned(),
        kind: NamespaceKind::System as i32,
        ..Default::default()
    };

    let encoded = namespace.encode_to_vec();
    assert!(encoded.ends_with(&[(10 << 3) as u8, 1]));
    let decoded = Namespace::decode(encoded.as_slice()).expect("decode namespace");
    assert_eq!(decoded.kind, NamespaceKind::System as i32);
    let legacy =
        Namespace::decode(b"\x0a\x09legacy-id".as_slice()).expect("decode legacy namespace");
    assert_eq!(legacy.id, "legacy-id");
    assert_eq!(legacy.kind, NamespaceKind::Business as i32);
}
