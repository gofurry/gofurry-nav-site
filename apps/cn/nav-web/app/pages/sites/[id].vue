<script setup lang="ts">
import { siteEntityPath } from '~/utils/siteRoutes'

const route = useRoute()
const localePath = useLocalePath()

await navigateTo(localePath(siteEntityPath(String(route.params.id ?? ''))), {
  redirectCode: 301,
  replace: true,
})
</script>
