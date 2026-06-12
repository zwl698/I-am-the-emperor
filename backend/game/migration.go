package game

func (s *GameState) EnsurePlayableState() {
	if s == nil {
		return
	}
	s.ensureRNG()
	if s.Phase == "" {
		s.Phase = PhasePrince
	}
	if s.Age <= 0 {
		if s.Phase == PhaseEmperor {
			s.Age = 18
		} else {
			s.Age = 1
		}
	}
	if s.Season == "" {
		s.Season = inferredSeason(s)
	}
	if s.Phase == PhaseEmperor && s.ReignYear < 1 {
		s.ReignYear = inferredReignYear(s)
	}
	if len(s.Provinces) == 0 {
		s.Provinces = startingProvinces(s.Dynasty.ID)
	}
	if len(s.Wars) == 0 {
		s.Wars = startingWars(s.Dynasty.ID)
	}
	if s.Crisis.Title == "" {
		s.Crisis = startingCrisis(s.Dynasty.ID)
	}
	if len(s.Objectives) == 0 {
		s.Objectives = startingObjectives(s.Dynasty.ID)
	}
	s.ensureCourtSystems()
	if s.Phase == PhaseEmperor && len(s.EventHand) == 0 {
		s.dealEventHand()
	}
	if s.Scene == nil && s.Ending == nil {
		s.Scene = s.nextScene()
	}
	s.updateObjectives()
}

func inferredSeason(s *GameState) string {
	seasons := []string{"春", "夏", "秋", "冬"}
	if s == nil || s.Phase != PhaseEmperor {
		return "春"
	}
	elapsed := max(0, s.Turn-5)
	return seasons[elapsed%len(seasons)]
}

func inferredReignYear(s *GameState) int {
	if s == nil {
		return 1
	}
	elapsed := max(0, s.Turn-5)
	return max(1, elapsed/4+1)
}
