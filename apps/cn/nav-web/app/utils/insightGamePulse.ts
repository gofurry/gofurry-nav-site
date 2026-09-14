import type { GameV2ListItem, GameV2PanelRecord } from '@/types/game'

export function pulseDiscountPrice(game: GameV2ListItem) {
  // The existing Panel ranks highest_discount using US prices, independently of its requested display region.
  return game.prices?.find(price => price.region === 'US') ?? (game.price?.region === 'US' ? game.price : null)
}

export function selectGamePulse(panel: GameV2PanelRecord | null) {
  const used = new Set<string>()
  const take = (games: GameV2ListItem[]) => {
    const game = games.find(item => item.id && !used.has(String(item.id))) ?? null
    if (game) used.add(String(game.id))
    return game
  }
  return {
    players: take((panel?.top_online ?? []).filter(game => game.online_count?.status === 'success' && Number.isFinite(game.online_count.count))),
    discount: take((panel?.highest_discount ?? []).filter(game => {
      const price = pulseDiscountPrice(game)
      return price?.available && !price.is_free && price.currency && Number.isFinite(price.final_amount) && price.discount_percent > 0
    })),
    latest: take(panel?.latest_games ?? []),
  }
}
