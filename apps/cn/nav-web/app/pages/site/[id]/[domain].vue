<script setup lang="ts">
import { siteTargetPath } from '~/utils/siteRoutes'

const route = useRoute()
const localePath = useLocalePath()
const rawDomain = Array.isArray(route.params.domain) ? route.params.domain[0] : route.params.domain

await navigateTo(localePath(siteTargetPath(String(route.params.id ?? ''), String(rawDomain ?? ''))), {
  redirectCode: 301,
  replace: true,
})
</script>
