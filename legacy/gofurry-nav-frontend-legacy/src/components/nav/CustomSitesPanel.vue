<template>
  <section v-if="isVisible" class="px-6">
    <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
      <div class="flex items-center justify-center" :style="panelStyle">
        <iframe
          v-if="hasCustomPanelCode"
          :srcdoc="customPanelCode"
          sandbox="allow-scripts"
          class="w-full h-full"
          :style="panelStyle"
          :title="t('customSites.customPanelFrameTitle')"
        ></iframe>

        <div
          v-else
          class="w-full rounded-lg border-2 border-black/30 bg-[rgba(18,24,37,0.55)] p-5"
          :style="panelStyle"
        >
          <div class="flex h-full flex-col">
            <div class="mb-4">
              <p
                class="text-xs font-medium uppercase tracking-[0.18em] text-slate-200"
              >
                {{ t("customSites.panelTitle") }}
              </p>
              <h2 class="mt-2 text-2xl font-semibold text-gray-200">
                {{ t(greetingKey) }}
              </h2>
            </div>

            <div class="mb-4">
              <p class="text-sm font-medium text-gray-300">
                {{ t("customSites.recentTitle") }}
              </p>
            </div>

            <SiteIconStrip
              :sites="recentSites"
              :empty-title="t('customSites.recentEmptyTitle')"
              :empty-description="t('customSites.recentEmptyDescription')"
            />
          </div>
        </div>
      </div>

      <div
        class="rounded-lg border-2 border-black/30 bg-[rgba(18,24,37,0.55)] p-5"
        :style="panelStyle"
      >
        <SiteIconStrip
          :sites="fixedSites"
          :empty-title="t('customSites.fixedEmptyTitle')"
          :empty-description="t('customSites.fixedEmptyDescription')"
        />
      </div>

      <div
        class="rounded-lg border-2 border-black/30 bg-[rgba(18,24,37,0.55)] p-5"
        :style="panelStyle"
      >
        <div v-if="showForm" class="mb-4 rounded-lg bg-white/40 p-4">
          <div class="grid gap-3">
            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700">
                {{ t("customSites.nameLabel") }}
              </label>
              <input
                v-model="form.name"
                type="text"
                class="w-full rounded-lg bg-white/40 px-4 py-3 text-sm text-slate-700 outline-none focus:ring-2 ring-orange-800/70 duration-300"
                :placeholder="t('customSites.namePlaceholder')"
              />
            </div>

            <div class="space-y-2">
              <label class="text-sm font-medium text-slate-700">
                {{ t("customSites.urlLabel") }}
              </label>
              <input
                v-model="form.url"
                type="text"
                class="w-full rounded-lg bg-white/40 px-4 py-3 text-sm text-slate-700 outline-none focus:ring-2 ring-orange-800/70 duration-300"
                :placeholder="t('customSites.urlPlaceholder')"
              />
            </div>
          </div>

          <p v-if="formError" class="mt-3 text-sm text-red-500">
            {{ formError }}
          </p>

          <div class="mt-4 flex justify-end gap-3">
            <button
              class="rounded-lg bg-white/40 px-4 py-2 text-sm text-slate-600 transition hover:bg-white/60 duration-300"
              @click="closeForm"
            >
              {{ t("common.cancel") }}
            </button>
            <button
              class="rounded-lg bg-orange-800/80 px-4 py-2 text-sm font-medium text-slate-300 hover:bg-orange-800 duration-300"
              @click="saveSite"
            >
              {{ t("common.save") }}
            </button>
          </div>
        </div>

        <SiteIconStrip
          :sites="customSites"
          :empty-title="t('customSites.emptyTitle')"
          :empty-description="t('customSites.emptyDescription')"
          editable
          reorderable
          show-add-tile
          :add-title="t('customSites.addSite')"
          @add="openCreateForm"
          @edit="openEditForm"
          @remove="removeSite"
          @reorder="reorderSites"
        />
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import SiteIconStrip, {
  type SiteStripItem,
} from "@/components/nav/SiteIconStrip.vue";
import {
  loadRecentSites,
  RECENT_SITES_EVENT,
  RECENT_SITES_STORAGE_KEY,
} from "@/utils/recentSites";
import {
  CUSTOM_PANEL_CODE_EVENT,
  CUSTOM_PANEL_HEIGHT_EVENT,
  CUSTOM_PANEL_CODE_STORAGE_KEY,
  CUSTOM_PANEL_HEIGHT_STORAGE_KEY,
  loadCustomPanelHeight,
  loadCustomPanelCode,
} from "@/utils/customPanel";

const { t } = useI18n();

const SITES_STORAGE_KEY = "customSites";
const VISIBILITY_STORAGE_KEY = "showCustomSites";
const VISIBILITY_EVENT = "custom-sites-visibility-change";

const isVisible = ref(true);
const customPanelCode = ref("");
const customPanelHeight = ref(loadCustomPanelHeight());
const recentSites = ref<SiteStripItem[]>([]);
const customSites = ref<SiteStripItem[]>([]);
const fixedSites: SiteStripItem[] = [
  {
    id: "Chatgpt",
    name: "Chatgpt",
    url: "chatgpt.com",
  },
  {
    id: "Gemini",
    name: "Gemini",
    url: "gemini.google.com",
  },
  {
    id: "Deepseek",
    name: "Deepseek",
    url: "chat.deepseek.com",
  },
  {
    id: "Doubao",
    name: "Doubao",
    url: "www.doubao.com",
  },

  {
    id: "Bilibili",
    name: "Bilibili",
    url: "www.bilibili.com",
  },
  {
    id: "Youtube",
    name: "Youtube",
    url: "www.youtube.com",
  },
  {
    id: "Steam",
    name: "Steam",
    url: "store.steampowered.com",
  },
  {
    id: "Itch",
    name: "itch",
    url: "itch.io",
  },
  {
    id: "Byrutgame",
    name: "Byrutgame",
    url: "byrutgame.org",
  },
  {
    id: "Github",
    name: "Github",
    url: "github.com",
  },
  {
    id: "Pixiv",
    name: "Pixiv",
    url: "www.pixiv.net",
  },
];
const showForm = ref(false);
const editingId = ref<string | null>(null);
const formError = ref("");
const form = reactive({
  name: "",
  url: "",
});

const greetingKey = computed(() => {
  const hour = new Date().getHours();
  if (hour < 6) return "customSites.greetings.lateNight";
  if (hour < 12) return "customSites.greetings.morning";
  if (hour < 18) return "customSites.greetings.afternoon";
  return "customSites.greetings.evening";
});

const hasCustomPanelCode = computed(() => Boolean(customPanelCode.value));
const panelStyle = computed(() => ({
  minHeight: `${customPanelHeight.value}px`,
}));

function syncVisibility() {
  const saved = localStorage.getItem(VISIBILITY_STORAGE_KEY);
  isVisible.value = saved !== "false";
}

function syncCustomPanelCode() {
  customPanelCode.value = loadCustomPanelCode();
}

function syncCustomPanelHeight() {
  customPanelHeight.value = loadCustomPanelHeight();
}

function loadRecentVisitedSites() {
  recentSites.value = loadRecentSites();
}

function loadSites() {
  const saved = localStorage.getItem(SITES_STORAGE_KEY);
  if (!saved) {
    customSites.value = [];
    return;
  }

  try {
    const parsed = JSON.parse(saved);
    customSites.value = Array.isArray(parsed)
      ? parsed.filter((item) => item?.id && item?.name && item?.url)
      : [];
  } catch {
    customSites.value = [];
  }
}

function persistSites() {
  localStorage.setItem(SITES_STORAGE_KEY, JSON.stringify(customSites.value));
}

function resetForm() {
  form.name = "";
  form.url = "";
  formError.value = "";
  editingId.value = null;
}

function closeForm() {
  showForm.value = false;
  resetForm();
}

function openCreateForm() {
  resetForm();
  showForm.value = true;
}

function openEditForm(site: SiteStripItem) {
  editingId.value = site.id;
  form.name = site.name;
  form.url = site.url;
  formError.value = "";
  showForm.value = true;
}

function normalizeUrl(url: string) {
  const trimmed = url.trim();
  if (!trimmed) {
    return "";
  }

  const withProtocol = /^https?:\/\//i.test(trimmed)
    ? trimmed
    : `https://${trimmed}`;
  return new URL(withProtocol).toString();
}

function generateId() {
  if (typeof crypto !== "undefined" && "randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function saveSite() {
  formError.value = "";

  const name = form.name.trim();
  if (!name) {
    formError.value = t("customSites.nameRequired");
    return;
  }

  let normalizedUrl = "";
  try {
    normalizedUrl = normalizeUrl(form.url);
  } catch {
    formError.value = t("customSites.urlInvalid");
    return;
  }

  if (!normalizedUrl) {
    formError.value = t("customSites.urlRequired");
    return;
  }

  if (editingId.value) {
    customSites.value = customSites.value.map((site) =>
      site.id === editingId.value ? { ...site, name, url: normalizedUrl } : site
    );
  } else {
    customSites.value.unshift({
      id: generateId(),
      name,
      url: normalizedUrl,
    });
  }

  persistSites();
  closeForm();
}

function removeSite(id: string) {
  customSites.value = customSites.value.filter((site) => site.id !== id);
  persistSites();
}

function reorderSites(payload: { draggedId: string; targetId: string }) {
  const fromIndex = customSites.value.findIndex(
    (site) => site.id === payload.draggedId
  );
  const toIndex = customSites.value.findIndex(
    (site) => site.id === payload.targetId
  );

  if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) {
    return;
  }

  const nextSites = [...customSites.value];
  const [movedSite] = nextSites.splice(fromIndex, 1);
  if (!movedSite) {
    return;
  }
  nextSites.splice(toIndex, 0, movedSite);
  customSites.value = nextSites;
  persistSites();
}

function handleStorage(event: StorageEvent) {
  if (event.key === SITES_STORAGE_KEY) {
    loadSites();
  }

  if (event.key === VISIBILITY_STORAGE_KEY) {
    syncVisibility();
  }

  if (event.key === RECENT_SITES_STORAGE_KEY) {
    loadRecentVisitedSites();
  }

  if (event.key === CUSTOM_PANEL_CODE_STORAGE_KEY) {
    syncCustomPanelCode();
  }

  if (event.key === CUSTOM_PANEL_HEIGHT_STORAGE_KEY) {
    syncCustomPanelHeight();
  }
}

onMounted(() => {
  syncVisibility();
  syncCustomPanelCode();
  syncCustomPanelHeight();
  loadSites();
  loadRecentVisitedSites();
  window.addEventListener("storage", handleStorage);
  window.addEventListener(VISIBILITY_EVENT, syncVisibility);
  window.addEventListener(RECENT_SITES_EVENT, loadRecentVisitedSites);
  window.addEventListener(CUSTOM_PANEL_CODE_EVENT, syncCustomPanelCode);
  window.addEventListener(CUSTOM_PANEL_HEIGHT_EVENT, syncCustomPanelHeight);
});

onUnmounted(() => {
  window.removeEventListener("storage", handleStorage);
  window.removeEventListener(VISIBILITY_EVENT, syncVisibility);
  window.removeEventListener(RECENT_SITES_EVENT, loadRecentVisitedSites);
  window.removeEventListener(CUSTOM_PANEL_CODE_EVENT, syncCustomPanelCode);
  window.removeEventListener(CUSTOM_PANEL_HEIGHT_EVENT, syncCustomPanelHeight);
});
</script>
