<template>
  <div class="nav-home-page flex w-full flex-1 flex-col" data-transition-home>
    <NavHeader :desktop-object-key="hero?.desktop?.object_key ?? null" :mobile-object-key="hero?.mobile?.object_key ?? null" />
    <p class="transition-notice relative px-4 py-6 text-center">{{ transitionNotice }}</p>
  </div>
</template>

<script setup lang="ts">
import NavHeader from '~/components/nav/NavHeader.vue'
import { getNavHomeHero } from '~/services/nav'
import { heroPreferenceKey } from '~/utils/heroPreferences'
import { transitionNotice } from '#shared/utils/transitionMode'

const settings = useHeroPreferences()
const preference = { ...settings.preference.value }
const emptyHero = () => ({ desktop: null, mobile: null })
const { data: hero } = await useAsyncData(`transition-hero:${heroPreferenceKey(preference)}`, () =>
  preference.mode === 'local' ? Promise.resolve(emptyHero()) : getNavHomeHero(preference).catch(emptyHero))
</script>
