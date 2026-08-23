package models

// GameID is the stable collector target identity used by scheduling and task
// orchestration. Persistence rows remain inside the generated sqlc package.
type GameID struct {
	ID    int64 `json:"id"`
	Appid int64 `json:"appid"`
}
