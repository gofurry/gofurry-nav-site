<script setup lang="ts">
const props = defineProps<{ modelValue: string[] }>()
const emit = defineEmits<{ 'update:modelValue': [value: string[]] }>()

function update(index: number, value: string) {
  const next = [...(props.modelValue ?? [])]
  next[index] = value
  emit('update:modelValue', next)
}

function addRow() {
  emit('update:modelValue', [...(props.modelValue ?? []), ''])
}

function removeRow(index: number) {
  const next = [...(props.modelValue ?? [])]
  next.splice(index, 1)
  emit('update:modelValue', next.length ? next : [''])
}
</script>

<template>
  <div class="space-y-2">
    <div v-for="(item, index) in modelValue" :key="index" class="flex gap-2">
      <input
        class="ui-control w-full px-3 py-2 text-sm"
        :value="item"
        @input="update(index, ($event.target as HTMLInputElement).value)"
      />
      <button type="button" class="ui-button ui-button--quiet px-3 text-sm" @click="removeRow(index)">删</button>
    </div>
    <button type="button" class="ui-button px-3 py-2 text-sm" @click="addRow">添加一项</button>
  </div>
</template>
