<template>
  <div class="bg-white/50 backdrop-blur-md rounded-2xl shadow p-6 min-h-[900px]">
    <!-- 顶部搜索 -->
    <div class="mb-2">
      <GameSidebarSearch />
    </div>

    <!-- 快捷功能 -->
    <GameSidebarActions>
      <!-- 默认前6个网站 -->
      <template #default-sites>
        <GameSiteItem href="https://store.steampowered.com" :icon="icons.steam" title="Steam" subtitle="steam" />
        <GameSiteItem href="https://itch.io" :icon="icons.itch" title="itch.io" subtitle="itch" />
        <GameSiteItem href="https://srk.games/" :icon="icons.srk" title="兽人控游戏库" subtitle="srkGames" />
        <GameSiteItem href="https://furrygames.top/" :icon="icons.defaultLogo" title="兽人控游戏索引" subtitle="furryGames" />
        <GameSiteItem href="https://f95zone.to.it/tag/furry/" :icon="icons.f95" title="F95Zone" subtitle="f95zoneFurry" />
        <GameSiteItem href="https://byrutgame.org/" :icon="icons.byrut" title="ByRUT Game" subtitle="byrut" />
      </template>

      <!-- 可展开的更多网站 -->
      <template #extra-sites>
        <GameSiteItem href="https://www.xiaoheihe.cn/" :icon="icons.xiaoheihe" title="小黑盒" subtitle="xiaoheihe" />
        <GameSiteItem href="https://afdian.com/" :icon="icons.afdian" title="爱发电" subtitle="afdian" />
        <GameSiteItem href="https://www.patreon.com/" :icon="icons.patreon" title="Patreon" subtitle="patreon" />
        <GameSiteItem href="https://www.kickstarter.com/" :icon="icons.kickstarter" title="Kickstarter" subtitle="kickstarter" />
        <GameSiteItem href="https://steamdb.info" :icon="icons.steamdb" title="SteamDB" subtitle="steamDB" />
        <GameSiteItem href="https://steambase.io" :icon="icons.steambase" title="SteamBase" subtitle="steamBase" />
        <GameSiteItem href="https://playtracker.net" :icon="icons.playtracker" title="PlayTracker" subtitle="playTracker" />
        <GameSiteItem href="https://app.sensortower.com/vgi" :icon="icons.sensortower" title="Sensor Tower" subtitle="sensorTower" />
        <GameSiteItem href="https://gamalytic.com" :icon="icons.gamalytic" title="Gamalytic" subtitle="gamalytic" />
        <GameSiteItem href="https://www.metacritic.com/" :icon="icons.metacritic" title="Metacritic" subtitle="metacritic" />
        <GameSiteItem href="https://www.pcgamingwiki.com/" :icon="icons.pcgamingwiki" title="PCGamingWiki" subtitle="pcGamingWiki" />
        <GameSiteItem href="https://gamefaqs.gamespot.com/" :icon="icons.gamefaqs" title="GameFAQs" subtitle="gameFAQs" />
      </template>
    </GameSidebarActions>

    <!-- 近期评论 -->
    <GameSidebarLatestReviews />
  </div>
</template>

<script setup lang="ts">
import GameSidebarActions from "@/components/game/main/sidebar/GameSidebarActions.vue";
import GameSidebarSearch from "@/components/game/main/sidebar/GameSidebarSearch.vue";
import GameSidebarLatestReviews from "@/components/game/main/sidebar/GameSidebarLatestReviews.vue";
import GameSiteItem from "@/components/game/main/sidebar/GameSiteItem.vue";

const LOGO_PREFIX = import.meta.env.VITE_GAME_SITE_LOGO_PREFIX_URL || '';
const defaultLogo = import.meta.env.VITE_SITE_DEFAULT_LOGO || '';

const icons = {
  defaultLogo,
  steam: `${LOGO_PREFIX}store.steampowered.com.png`,
  itch: `${LOGO_PREFIX}itch.io-favicon.ico`,
  srk: `${LOGO_PREFIX}srk.games-favicon.ico`,
  f95: `${LOGO_PREFIX}f95zone.to.it-favicon.ico`,
  byrut: `${LOGO_PREFIX}byrutgame.org-favicon.ico`,
  afdian: `${LOGO_PREFIX}afdian.com-favicon.ico`,
  patreon: `${LOGO_PREFIX}www.patreon.com-favicon.ico`,
  kickstarter: `${LOGO_PREFIX}www.kickstarter.com.jpg`,
  steamdb: `${LOGO_PREFIX}steamdb.info-favicon.ico`,
  steambase: `${LOGO_PREFIX}steambase.io-favicon.ico`,
  playtracker: `${LOGO_PREFIX}playtracker.net-favicon.png`,
  sensortower: `${LOGO_PREFIX}app.sensortower.com-favicon.ico`,
  gamalytic: `${LOGO_PREFIX}gamalytic.com-favicon.ico`,
  metacritic: `${LOGO_PREFIX}metacritic.com-favicon-32.png`,
  pcgamingwiki: `${LOGO_PREFIX}www.pcgamingwiki.com-favicon.ico`,
  gamefaqs: `${LOGO_PREFIX}gamefaqs.gamespot.com-favicon.ico`,
  xiaoheihe: `${LOGO_PREFIX}xiaoheihe.png`
};
</script>
