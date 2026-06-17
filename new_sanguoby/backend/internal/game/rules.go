package game

const (
	maxResourceValue = 30000
	staminaRenew     = 4
)

func (s *GameState) AdvanceMonth() {
	s.Date.Month++
	if s.Date.Month > 12 {
		s.Date.Year++
		s.Date.Month = 1
	}

	for i := range s.Generals {
		s.Generals[i].Stamina = minInt(100, s.Generals[i].Stamina+staminaRenew)
	}

	for i := range s.Cities {
		city := &s.Cities[i]
		if city.OwnerID == "" {
			continue
		}

		if s.Date.Month%3 == 0 {
			city.Money = minInt(maxResourceValue, city.Money+city.Commerce/2)
		}
		if s.Date.Month == 6 || s.Date.Month == 10 {
			city.Food = minInt(maxResourceValue, city.Food+city.Farming/4)
		}

		city.Population = minInt(city.PopulationLimit, city.Population+50)
		s.consumeMonthlyFood(city)
	}

	s.Log = append([]string{formatDate(s.Date) + " 政令已结算。"}, s.Log...)
	if len(s.Log) > 8 {
		s.Log = s.Log[:8]
	}
}

func (s *GameState) consumeMonthlyFood(city *City) {
	totalSoldiers := city.Garrison
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID {
			totalSoldiers += s.Generals[i].Soldiers
		}
	}

	upkeep := totalSoldiers / 50
	if upkeep <= 0 {
		return
	}
	if city.Food > upkeep {
		city.Food -= upkeep
		return
	}

	city.Food = 0
	city.State = CityStateFamine
	for i := range s.Generals {
		if s.Generals[i].CityID == city.ID {
			s.Generals[i].Soldiers /= 2
		}
	}
}

func formatDate(date Date) string {
	return itoa(date.Year) + "年" + itoa(date.Month) + "月"
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[i:])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
