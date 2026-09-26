package collaboration

import "context"

// At most one query per domain, independent of the number of reference cards.
func (s *Service) boardReferences(ctx context.Context, nodes []BoardNode) ([]BoardReference, error) {
	ids := map[string][]int64{}
	seen := map[string]map[int64]bool{}
	for _, node := range nodes {
		if node.ReferenceKind == nil || node.ReferenceID == nil {
			continue
		}
		kind, id := *node.ReferenceKind, *node.ReferenceID
		if seen[kind] == nil {
			seen[kind] = map[int64]bool{}
		}
		if !seen[kind][id] {
			ids[kind] = append(ids[kind], id)
			seen[kind][id] = true
		}
	}
	result := []BoardReference{}
	for _, kind := range []string{"idea", "game", "site"} {
		if len(ids[kind]) == 0 {
			continue
		}
		found := map[int64]BoardReference{}
		switch kind {
		case "idea":
			rows, err := s.queries.ListBoardIdeaReferences(ctx, ids[kind])
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				found[row.ID] = BoardReference{Kind: kind, ID: row.ID, Title: row.Title, Status: row.Status, ContentKind: row.Kind}
			}
		case "game":
			rows, err := s.game.ListBoardGameReferences(ctx, ids[kind])
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				found[row.ID] = BoardReference{Kind: kind, ID: row.ID, Title: row.Name, ContentKind: kind}
			}
		case "site":
			rows, err := s.nav.ListBoardSiteReferences(ctx, ids[kind])
			if err != nil {
				return nil, err
			}
			for _, row := range rows {
				found[row.ID] = BoardReference{Kind: kind, ID: row.ID, Title: row.Name, ContentKind: kind}
			}
		}
		for _, id := range ids[kind] {
			ref, ok := found[id]
			if !ok {
				ref = BoardReference{Kind: kind, ID: id, Missing: true}
			}
			result = append(result, ref)
		}
	}
	return result, nil
}
