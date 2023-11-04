package entity

type Player struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	LoginTime     string `json:"login_time"`
	MatchesPlayed int    `json:"matches_played"`
}
