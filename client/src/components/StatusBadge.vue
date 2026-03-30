<template>
  <n-tag :type="type" :bordered="false">
    {{ statusLabel }}
  </n-tag>
</template>

<script setup lang="ts">
import { NTag } from 'naive-ui';
import { computed } from 'vue';
import { useDictionariesStore } from '../stores/dictionaries';

const props = defineProps<{
  statusId?: number;
  statusCode?: string;
}>();

const dictStore = useDictionariesStore();

const status = computed(() => {
  if (props.statusId) {
    return dictStore.getRequestStatusById(props.statusId);
  }
  if (props.statusCode) {
    return dictStore.requestStatuses.find((s) => s.code === props.statusCode);
  }
  return undefined;
});

const type = computed(() => {
  const code = status.value?.code;
  switch (code) {
    case 'pending':
      return 'warning';
    case 'processing':
      return 'info';
    case 'completed':
      return 'success';
    case 'failed':
      return 'error';
    case 'partial':
      return 'warning';
    default:
      return 'default';
  }
});

const statusLabel = computed(() => {
  return status.value?.name || props.statusCode || props.statusId || '-';
});
</script>
