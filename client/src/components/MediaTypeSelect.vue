<template>
  <n-select
    :value="modelValue"
    :options="options"
    :placeholder="placeholder"
    clearable
    multiple
    @update:value="$emit('update:modelValue', $event)"
  />
</template>

<script setup lang="ts">
import { NSelect } from 'naive-ui';
import { computed, onMounted } from 'vue';
import { useDictionariesStore } from '../stores/dictionaries';

defineProps<{
  modelValue?: number[] | null;
  placeholder?: string;
}>();

defineEmits<{
  'update:modelValue': [value: number[] | null];
}>();

const dictStore = useDictionariesStore();

const options = computed(() =>
  dictStore.mediaTypes.map((mt) => ({
    label: mt.name,
    value: mt.id,
  }))
);

onMounted(async () => {
  await dictStore.loadDictionaries();
});
</script>
