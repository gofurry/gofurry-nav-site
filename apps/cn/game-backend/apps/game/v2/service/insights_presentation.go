package service

import v2models "github.com/gofurry/gofurry-game-backend/apps/game/v2/models"

// Presentation identity only; ranking and price evidence remain owned by their facts.
func insightGameEntity(id int64, name, asset string) v2models.InsightEntityRef {
	entity := v2models.InsightEntityRef{ID: id, Name: name}
	if asset = normalizeSteamAssetURL(asset); asset != "" {
		entity.Visual = &v2models.InsightEntityVisual{Kind: "game_header", Asset: asset}
	}
	return entity
}
