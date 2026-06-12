package game

import "testing"

func TestEventDeckCatalogHasAtLeast120DistinctCards(t *testing.T) {
	cards := EventDeckCatalog()
	if len(cards) < 120 {
		t.Fatalf("expected at least 120 event cards, got %d", len(cards))
	}
	ids := map[string]bool{}
	categories := map[string]int{}
	for _, card := range cards {
		if card.ID == "" || card.Title == "" || card.Category == "" || card.Summary == "" || card.Hook == "" {
			t.Fatalf("event card should expose full narrative identity: %+v", card)
		}
		if ids[card.ID] {
			t.Fatalf("duplicate card id %q", card.ID)
		}
		ids[card.ID] = true
		categories[card.Category]++
	}
	if len(categories) < 12 {
		t.Fatalf("expected at least 12 event categories, got %+v", categories)
	}
	for category, count := range categories {
		if count < 8 {
			t.Fatalf("expected category %q to have depth, got %d", category, count)
		}
	}
}

func TestEmperorStateDealsDynamicEventHand(t *testing.T) {
	state := NewGame(1401)
	state.ForceCoronationForTest()

	if len(state.EventHand) != 5 {
		t.Fatalf("expected 5 event cards in hand, got %+v", state.EventHand)
	}
	seen := map[string]bool{}
	for _, card := range state.EventHand {
		if seen[card.ID] {
			t.Fatalf("event hand should not repeat cards: %+v", state.EventHand)
		}
		seen[card.ID] = true
		if card.Urgency <= 0 || card.Severity <= 0 {
			t.Fatalf("event hand should expose urgency and severity: %+v", card)
		}
	}
}

func TestEventHandChangesAfterSeasonAdvance(t *testing.T) {
	state, err := NewGameWithDynasty("chengping", 1402)
	if err != nil {
		t.Fatalf("new dynasty game: %v", err)
	}
	state.ForceCoronationForTest()
	first := eventHandSignature(state.EventHand)

	if _, err := state.ApplyChoice(state.Scene.Choices[0].ID); err != nil {
		t.Fatalf("advance season: %v", err)
	}
	second := eventHandSignature(state.EventHand)

	if first == second {
		t.Fatalf("expected event hand to change after season, got %s", first)
	}
}

func eventHandSignature(cards []EventCard) string {
	out := ""
	for _, card := range cards {
		out += card.ID + "|"
	}
	return out
}
