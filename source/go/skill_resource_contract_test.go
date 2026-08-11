package specification

import (
	"testing"

	"google.golang.org/protobuf/proto"

	security "github.com/pole-io/specification/source/go/api/v1/security"
)

func TestSkillResourceAuthorizationWireContract(t *testing.T) {
	if got := int32(security.ResourceType_SkillResources); got != 32 {
		t.Fatalf("SkillResources enum = %d, want 32", got)
	}

	original := &security.StrategyResources{
		StrategyId: "policy-1",
		Skills: []*security.StrategyResourceEntry{{
			Id:   "skill-1",
			Name: "publisher/example",
		}},
	}
	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal StrategyResources: %v", err)
	}
	decoded := &security.StrategyResources{}
	if err := proto.Unmarshal(wire, decoded); err != nil {
		t.Fatalf("unmarshal StrategyResources: %v", err)
	}
	if len(decoded.GetSkills()) != 1 || decoded.GetSkills()[0].GetId() != "skill-1" {
		t.Fatalf("skills did not round-trip: %#v", decoded.GetSkills())
	}
}
