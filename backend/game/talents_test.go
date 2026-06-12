package game

import "testing"

func TestNewGameHasLargeHistoricallyInspiredTalentPool(t *testing.T) {
	state, err := NewGameWithDynasty("xuanshuo", 4101)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	if len(state.TalentPool) < 300 {
		t.Fatalf("expected at least 300 candidates, got %d", len(state.TalentPool))
	}

	ids := map[string]bool{}
	inspirations := map[string]bool{}
	origins := map[string]bool{}
	specialties := map[Domain]bool{}
	highAbility := 0
	highIntegrity := 0
	highAmbition := 0
	for _, talent := range state.TalentPool {
		if ids[talent.ID] {
			t.Fatalf("duplicate talent id %q", talent.ID)
		}
		ids[talent.ID] = true
		if talent.Name == "" || talent.Role == "" || talent.Trait == "" || talent.Inspiration == "" || talent.Origin == "" || talent.Specialty == "" {
			t.Fatalf("talent missing playable metadata: %+v", talent)
		}
		inspirations[talent.Inspiration] = true
		origins[talent.Origin] = true
		specialties[talent.Specialty] = true
		if talent.Ability >= 86 {
			highAbility++
		}
		if talent.Integrity >= 86 {
			highIntegrity++
		}
		if talent.Ambition >= 78 {
			highAmbition++
		}
	}
	if len(inspirations) < 80 {
		t.Fatalf("expected broad historical inspiration coverage, got %d", len(inspirations))
	}
	if len(origins) < 10 {
		t.Fatalf("expected Chinese and world origin coverage, got %d", len(origins))
	}
	if len(specialties) < 7 {
		t.Fatalf("expected all major court specialties, got %+v", specialties)
	}
	if highAbility < 40 || highIntegrity < 30 || highAmbition < 30 {
		t.Fatalf("expected varied standout attributes, ability=%d integrity=%d ambition=%d", highAbility, highIntegrity, highAmbition)
	}
}

func TestRecruitTalentMovesCandidateIntoCourtAndPreventsDuplicate(t *testing.T) {
	state, err := NewGameWithDynasty("jingyao", 4102)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	state.ForceCoronationForTest()
	beforeCourt := len(state.Court)
	beforePool := len(state.TalentPool)
	target := state.TalentPool[0]

	resolution, err := state.ApplyOrder(OrderRequest{Kind: OrderRecruitTalent, Target: target.ID})
	if err != nil {
		t.Fatalf("recruit talent: %v", err)
	}

	if len(state.Court) != beforeCourt+1 {
		t.Fatalf("expected court size to grow, before %d after %d", beforeCourt, len(state.Court))
	}
	if len(state.TalentPool) != beforePool-1 {
		t.Fatalf("expected talent pool to shrink, before %d after %d", beforePool, len(state.TalentPool))
	}
	recruited, ok := ministerByID(state, target.ID)
	if !ok {
		t.Fatalf("expected recruited talent in court")
	}
	if recruited.Inspiration != target.Inspiration || recruited.Specialty != target.Specialty {
		t.Fatalf("expected metadata preserved, got %+v want %+v", recruited, target)
	}
	if resolution.Effects == (Effects{}) {
		t.Fatalf("expected recruitment to have strategic effects")
	}
	if _, err := state.ApplyOrder(OrderRequest{Kind: OrderRecruitTalent, Target: target.ID}); err == nil {
		t.Fatalf("expected duplicate recruitment to fail")
	}
}
