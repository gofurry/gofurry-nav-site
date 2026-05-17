<template>
  <section class="rag-prompt-panel" :aria-label="title">
    <div class="rag-prompt-panel__header">
      <p>{{ title }}</p>
      <span>{{ description }}</span>
    </div>

    <div class="rag-prompt-panel__templates">
      <button
        v-for="template in templates"
        :key="template.id"
        type="button"
        @click="ask(template.prompt)"
      >
        <span>{{ template.title }}</span>
        <small>{{ template.description }}</small>
      </button>
    </div>

    <form class="rag-prompt-panel__custom" @submit.prevent="ask(customPrompt)">
      <input
        v-model.trim="customPrompt"
        type="text"
        :placeholder="placeholder"
      />
      <button type="submit" :disabled="!customPrompt">
        {{ submitLabel }}
      </button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

type RagPromptTemplate = {
  id: string
  title: string
  description: string
  prompt: string
}

defineProps<{
  title: string
  description: string
  placeholder: string
  submitLabel: string
  templates: RagPromptTemplate[]
}>()

const emit = defineEmits<{
  (event: 'ask', prompt: string): void
}>()

const customPrompt = ref('')

function ask(prompt: string) {
  const question = prompt.trim()
  if (!question) {
    return
  }
  emit('ask', question)
  customPrompt.value = ''
}
</script>

<style scoped>
.rag-prompt-panel {
  width: min(21rem, calc(100vw - 5rem));
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.56);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.76);
  box-shadow: 0 18px 44px rgba(76, 42, 18, 0.16);
  color: #1f2937;
  backdrop-filter: blur(18px);
}

.rag-prompt-panel__header {
  padding: 0.85rem 0.95rem 0.75rem;
  border-bottom: 1px solid rgba(120, 113, 108, 0.16);
}

.rag-prompt-panel__header p {
  margin: 0;
  font-size: 0.82rem;
  font-weight: 700;
  color: #292524;
}

.rag-prompt-panel__header span {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.72rem;
  line-height: 1.45;
  color: rgba(68, 64, 60, 0.72);
}

.rag-prompt-panel__templates {
  display: grid;
  gap: 0.35rem;
  padding: 0.55rem;
}

.rag-prompt-panel__templates button {
  display: grid;
  gap: 0.18rem;
  width: 100%;
  border: 0;
  border-radius: 0.55rem;
  background: rgba(255, 247, 237, 0.7);
  padding: 0.58rem 0.65rem;
  text-align: left;
  cursor: pointer;
  transition: background 160ms ease, color 160ms ease;
}

.rag-prompt-panel__templates button:hover {
  background: rgba(254, 215, 170, 0.56);
  color: #9a3412;
}

.rag-prompt-panel__templates span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.78rem;
  font-weight: 700;
}

.rag-prompt-panel__templates small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.68rem;
  color: rgba(68, 64, 60, 0.62);
}

.rag-prompt-panel__custom {
  display: flex;
  gap: 0.4rem;
  padding: 0.65rem;
  border-top: 1px solid rgba(120, 113, 108, 0.14);
}

.rag-prompt-panel__custom input {
  min-width: 0;
  flex: 1;
  border: 1px solid rgba(120, 113, 108, 0.18);
  border-radius: 0.55rem;
  background: rgba(255, 255, 255, 0.72);
  padding: 0.52rem 0.62rem;
  font-size: 0.76rem;
  outline: none;
}

.rag-prompt-panel__custom input:focus {
  border-color: rgba(251, 146, 60, 0.56);
}

.rag-prompt-panel__custom button {
  border: 0;
  border-radius: 0.55rem;
  background: #292524;
  padding: 0.52rem 0.72rem;
  color: #fff7ed;
  font-size: 0.74rem;
  cursor: pointer;
  transition: opacity 160ms ease;
}

.rag-prompt-panel__custom button:disabled {
  cursor: not-allowed;
  opacity: 0.36;
}
</style>
