<template>
  <div class="insight-compare-matrix-scroll" role="region" :aria-label="label" tabindex="0">
    <table class="insight-compare-matrix" :style="{ '--compare-columns': entities.length }">
      <caption class="sr-only">{{ label }}</caption>
      <colgroup><col class="insight-compare-matrix__fact" /><col v-for="entity in entities" :key="entity.id" /></colgroup>
      <thead><tr>
        <th scope="col">{{ $t('insights.compare.fact') }}</th>
        <th v-for="entity in entities" :key="entity.id" scope="col" :data-compare-entity-id="entity.id">
          <NuxtLink :to="localePath((domain === 'site' ? '/site/' : '/games/') + entity.id)">
            <InsightEntityMedia :domain="domain" :entity="entity" />
            <strong>{{ entity.name || '#' + entity.id }}</strong><small>#{{ entity.id }}</small>
          </NuxtLink>
        </th>
      </tr></thead>
      <tbody v-for="group in groups" :key="group.key" :data-compare-group="group.key">
        <tr class="insight-compare-matrix__group"><th :colspan="entities.length + 1" scope="rowgroup"><span>{{ group.label }}</span></th></tr>
        <tr v-for="row in group.rows" :key="row.key" :data-compare-fact="row.key">
          <th scope="row">{{ row.label }}</th>
          <td v-for="(cell, index) in row.cells" :key="entities[index]?.id" v-bind="cell.attributes" :data-capability-state="cell.state">
            <span v-if="cell.state" class="insight-compare-matrix__mark" aria-hidden="true">{{ cell.state === 'supported' ? '●' : cell.state === 'unsupported' ? '○' : '–' }}</span>{{ cell.text }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup lang="ts">
import InsightEntityMedia from '@/components/insights/entity/InsightEntityMedia.vue'
import type { InsightDomain, InsightEntityRef } from '@/types/insights'
import type { CompareMatrixGroup } from '@/types/insightCompare'
defineProps<{ domain: InsightDomain; entities: InsightEntityRef[]; groups: CompareMatrixGroup[]; label: string }>()
const localePath = useLocalePath()
</script>
