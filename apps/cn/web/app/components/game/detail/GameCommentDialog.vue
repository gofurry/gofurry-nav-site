<template>
  <GameReviewDialog
    :visible="visible"
    :game-id="gameId"
    :game-name="gameName"
    @close="emit('close')"
    @submitted="emit('submitted')"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GameReviewDialog from '@/components/game/common/GameReviewDialog.vue'

const props = defineProps<{
  visible: boolean
  gameName: string
  gameId?: string
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'submitted'): void
}>()

const route = useRoute()
const gameId = computed(() => props.gameId || String(route.params.id || ''))
</script>
