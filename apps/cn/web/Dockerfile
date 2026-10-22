FROM node:24.22.0-bookworm-slim AS base

# Temporary Linux AMD64 bridge until the official 24.22.1 Bookworm Slim image is
# available. Then restore that FROM tag and remove this ENV/RUN installation.
ENV NODE_VERSION=24.22.1
# SHA-256 is pinned from https://nodejs.org/dist/v24.22.1/SHASUMS256.txt.
RUN set -eux; \
    test "$(dpkg --print-architecture)" = amd64; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates curl; \
    curl --fail --show-error --location --proto '=https' --tlsv1.2 \
      --output /tmp/node-v24.22.1-linux-x64.tar.gz \
      https://nodejs.org/dist/v24.22.1/node-v24.22.1-linux-x64.tar.gz; \
    echo 'eb6728e77c576a35162c122c45ccadf224d89ff00abc8dae46cc2897c6002d7c  /tmp/node-v24.22.1-linux-x64.tar.gz' | sha256sum --check --strict -; \
    tar -xzf /tmp/node-v24.22.1-linux-x64.tar.gz -C /usr/local --strip-components=1 --no-same-owner; \
    rm /tmp/node-v24.22.1-linux-x64.tar.gz; \
    apt-get purge -y --auto-remove curl; \
    rm -rf /var/lib/apt/lists/*; \
    test "$(node --version)" = v24.22.1; \
    npm --version

WORKDIR /app
ENV NUXT_TELEMETRY_DISABLED=1

FROM base AS deps
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
# npm only bootstraps the exact pnpm declared by this standalone project.
RUN npm install --global "$(node -p 'require("./package.json").packageManager')" \
    && pnpm install --frozen-lockfile

FROM deps AS build

ARG NODE_BUILD_HEAP_MB=4096
ENV NODE_OPTIONS=--max-old-space-size=${NODE_BUILD_HEAP_MB}

ARG NUXT_PUBLIC_SITE_URL=https://go-furry.com
ARG NUXT_PUBLIC_NAV_API_BASE=https://nav.go-furry.com/api/v1
ARG NUXT_PUBLIC_NAV_V2_API_BASE=https://nav.go-furry.com/api/v2
ARG NUXT_PUBLIC_GAME_API_BASE=https://game.go-furry.com/api/v1
ARG NUXT_PUBLIC_UPTIME_URL=https://status.go-furry.com
ARG NUXT_PUBLIC_ASSET_PRIMARY_BASE=https://assets.go-furry.com
ARG NUXT_PUBLIC_ASSET_MIRROR_BASE=https://assets.gofurry.com
ARG NUXT_PUBLIC_STEAM_APP_PREFIX_URL=https://store.steampowered.com/app/
ARG NAV_API_INTERNAL_BASE=http://10.6.0.11:9999/api/v1
ARG NAV_V2_API_INTERNAL_BASE=http://10.6.0.11:9999/api/v2
ARG GAME_API_INTERNAL_BASE=http://10.6.0.11:9998/api/v1

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NUXT_PUBLIC_SITE_URL=${NUXT_PUBLIC_SITE_URL}
ENV NUXT_PUBLIC_NAV_API_BASE=${NUXT_PUBLIC_NAV_API_BASE}
ENV NUXT_PUBLIC_NAV_V2_API_BASE=${NUXT_PUBLIC_NAV_V2_API_BASE}
ENV NUXT_PUBLIC_GAME_API_BASE=${NUXT_PUBLIC_GAME_API_BASE}
ENV NUXT_PUBLIC_UPTIME_URL=${NUXT_PUBLIC_UPTIME_URL}
ENV NUXT_PUBLIC_ASSET_PRIMARY_BASE=${NUXT_PUBLIC_ASSET_PRIMARY_BASE}
ENV NUXT_PUBLIC_ASSET_MIRROR_BASE=${NUXT_PUBLIC_ASSET_MIRROR_BASE}
ENV NUXT_PUBLIC_STEAM_APP_PREFIX_URL=${NUXT_PUBLIC_STEAM_APP_PREFIX_URL}
ENV NAV_API_INTERNAL_BASE=${NAV_API_INTERNAL_BASE}
ENV NAV_V2_API_INTERNAL_BASE=${NAV_V2_API_INTERNAL_BASE}
ENV GAME_API_INTERNAL_BASE=${GAME_API_INTERNAL_BASE}

COPY . .
# The frozen install prepared an incomplete tree in deps. Refresh Nuxt's paths/types now
# that app/ and nuxt.config.ts exist, before Vue loads imported prop types.
RUN pnpm exec nuxt prepare && pnpm run build

FROM deps AS runner

WORKDIR /app

ARG NUXT_PUBLIC_SITE_URL=https://go-furry.com
ARG NUXT_PUBLIC_NAV_API_BASE=https://nav.go-furry.com/api/v1
ARG NUXT_PUBLIC_NAV_V2_API_BASE=https://nav.go-furry.com/api/v2
ARG NUXT_PUBLIC_GAME_API_BASE=https://game.go-furry.com/api/v1
ARG NUXT_PUBLIC_UPTIME_URL=https://status.go-furry.com
ARG NUXT_PUBLIC_ASSET_PRIMARY_BASE=https://assets.go-furry.com
ARG NUXT_PUBLIC_ASSET_MIRROR_BASE=https://assets.gofurry.com
ARG NUXT_PUBLIC_STEAM_APP_PREFIX_URL=https://store.steampowered.com/app/
ARG NAV_API_INTERNAL_BASE=http://10.6.0.11:9999/api/v1
ARG NAV_V2_API_INTERNAL_BASE=http://10.6.0.11:9999/api/v2
ARG GAME_API_INTERNAL_BASE=http://10.6.0.11:9998/api/v1

ENV NODE_ENV=production
ENV NUXT_TELEMETRY_DISABLED=1
ENV HOST=0.0.0.0
ENV PORT=3000
ENV NUXT_PUBLIC_SITE_URL=${NUXT_PUBLIC_SITE_URL}
ENV NUXT_PUBLIC_NAV_API_BASE=${NUXT_PUBLIC_NAV_API_BASE}
ENV NUXT_PUBLIC_NAV_V2_API_BASE=${NUXT_PUBLIC_NAV_V2_API_BASE}
ENV NUXT_PUBLIC_GAME_API_BASE=${NUXT_PUBLIC_GAME_API_BASE}
ENV NUXT_PUBLIC_UPTIME_URL=${NUXT_PUBLIC_UPTIME_URL}
ENV NUXT_PUBLIC_ASSET_PRIMARY_BASE=${NUXT_PUBLIC_ASSET_PRIMARY_BASE}
ENV NUXT_PUBLIC_ASSET_MIRROR_BASE=${NUXT_PUBLIC_ASSET_MIRROR_BASE}
ENV NUXT_PUBLIC_STEAM_APP_PREFIX_URL=${NUXT_PUBLIC_STEAM_APP_PREFIX_URL}
ENV NAV_API_INTERNAL_BASE=${NAV_API_INTERNAL_BASE}
ENV NAV_V2_API_INTERNAL_BASE=${NAV_V2_API_INTERNAL_BASE}
ENV GAME_API_INTERNAL_BASE=${GAME_API_INTERNAL_BASE}

COPY --from=build /app/.output ./.output

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
