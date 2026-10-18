import { defineConfig } from 'vitest/config'
import { defineVitestProject } from '@nuxt/test-utils/config'

export default defineConfig({
  test: {
    projects: [
      {
        test: {
          name: 'unit',
          include: ['tests/unit/**/*.test.ts'],
          environment: 'node',
          restoreMocks: true,
          unstubGlobals: true,
        },
      },
      await defineVitestProject({
        test: {
          name: 'nuxt',
          include: ['tests/nuxt/**/*.nuxt.test.ts'],
          environment: 'nuxt',
          // First mount also transforms the real app/router on a cold CI cache.
          testTimeout: 30_000,
          hookTimeout: 30_000,
          setupFiles: ['./tests/nuxt/setup.ts'],
          environmentOptions: { nuxt: { domEnvironment: 'happy-dom' } },
        },
      }),
    ],
  },
})
