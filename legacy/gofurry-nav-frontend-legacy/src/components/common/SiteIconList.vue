<template>
  <div class="flex flex-wrap gap-2">
    <a
        v-for="(item, i) in items"
        :key="item.key + i"
        :href="item.value"
        target="_blank"
        rel="noopener noreferrer"
        class="w-8 h-8 flex items-center justify-center
             rounded-lg bg-orange-100 hover:bg-orange-200 transition"
        :title="item.key"
    >
      <img
          :src="getIconUrl(item.key)"
          class="w-5 h-5 object-contain"
          :alt="item.key"
      />
    </a>
  </div>
</template>

<script setup lang="ts">
import type { KvModel } from '@/types/game'

defineProps<{
  items: KvModel[]
}>()

const ICON_MAP: Record<string, string> = {
  official: 'go.svg',
  bsky: 'bsky.app-favicon.ico',
  discord: 'discord.com-favicon.ico',
  afdian: 'afdian.com-favicon.ico',
  twitter: 'x.com-favicon.ico',
  qq: 'qq.ico',
  youtube: 'www.youtube.ico',
  telegram: 't.png',
  reddit: 'www.reddit.png',
  bilibili: 'bilibili.png',
  weibo: 'weibo.ico',
  tiktok: 'www.tiktok.png',
  steamdb: 'steamdb.info-favicon.ico',
  steambase: 'steambase.io-favicon.ico',
  gamalytic: 'gamalytic.com-favicon.ico',
  vgi: 'app.sensortower.com-favicon.ico',
  pcgamingwiki: 'www.pcgamingwiki.com-favicon.ico',
  playtracker: 'playtracker.net-favicon.ico',
  metacritic: 'metacritic.com-favicon-32.png',
  gamefaqs: 'gamefaqs.gamespot.com-favicon.ico',
  instagram: 'instagram.png',
  facebook: 'facebook.png',
  twitch: 'twitch.png',
  vk: 'vk.png',
  github: 'github.png',
  gitee: 'gitee.ico',
  steam: 'steampowered.ico',
  itch: 'itch.ico',
  douyin: 'douyin.ico',
  zhihu: 'zhihu.png',
  tieba: 'tieba.baidu.ico',
  xiaohongshu: 'xiaohongshu.ico',
  music163: 'music.163.com-favicon.ico',
  pixiv: 'pixiv.png',
  xiaoheihe: 'xiaoheihe.png'
}

const BASE_URL = import.meta.env.VITE_GAME_SITE_LOGO_PREFIX_URL || ''

const getIconUrl = (key: string) => {
  return ICON_MAP[key] ? BASE_URL + ICON_MAP[key] : ''
}
</script>
