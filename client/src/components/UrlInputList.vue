<template>
  <n-input
    :value="modelValue"
    type="textarea"
    :rows="rows"
    :placeholder="placeholder"
    @update:value="emit('update:modelValue', $event)"
  />
  <div class="url-count">
    URL: {{ count }} / {{ max }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { NInput } from 'naive-ui';

const props = withDefaults(defineProps<{
  modelValue: string;
  rows?: number;
  max?: number;
  placeholder?: string;
}>(), {
  rows: 8,
  max: 100,
  placeholder: 'https://example.com/page1\nhttps://example.com/page2',
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const count = computed(() => {
  return props.modelValue.split('\n').filter((u: string) => u.trim()).length;
});
</script>

<style scoped>
.url-count {
  margin-top: 4px;
  font-size: 12px;
  color: #666;
  text-align: right;
}
</style>
