use pole_specification::v1::{ResourceType, StrategyResourceEntry, StrategyResources};
use prost::Message;

#[test]
fn skill_resource_authorization_keeps_stable_wire_tags() {
    assert_eq!(ResourceType::SkillResources as i32, 32);

    let resources = StrategyResources {
        strategy_id: "policy-1".to_string(),
        skills: vec![StrategyResourceEntry {
            id: "skill-1".to_string(),
            namespace: String::new(),
            name: "publisher/example".to_string(),
        }],
        ..Default::default()
    };
    let bytes = resources.encode_to_vec();
    let decoded = StrategyResources::decode(bytes.as_slice()).expect("decode StrategyResources");

    assert_eq!(decoded.skills.len(), 1);
    assert_eq!(decoded.skills[0].id, "skill-1");
}
