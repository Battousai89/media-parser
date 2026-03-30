<template>
  <n-data-table
    :columns="columns"
    :data="sources"
    :row-key="(row: RequestSourceItem) => row.source_id"
  />
</template>

<script setup lang="ts">
import { h } from 'vue';
import type { DataTableColumns, DataTableBaseColumn } from 'naive-ui';
import type { RequestSourceItem } from '../types';
import StatusBadge from './StatusBadge.vue';

defineProps<{
  sources: RequestSourceItem[];
}>();

const columns: DataTableColumns<RequestSourceItem> = [
  {
    title: 'Источник',
    key: 'source_name',
    width: 200,
    render: (row) => row.source_name || `Source #${row.source_id}`,
  } as DataTableBaseColumn<RequestSourceItem>,
  {
    title: 'Статус',
    key: 'status_id',
    width: 120,
    render: (row) => h(StatusBadge, { statusId: row.status_id }),
  } as DataTableBaseColumn<RequestSourceItem>,
  {
    title: 'Найдено / Запрошено',
    key: 'counts',
    width: 150,
    render: (row) => `${row.parsed_count} / ${row.media_count}`,
  } as DataTableBaseColumn<RequestSourceItem>,
  {
    title: 'Ошибка',
    key: 'error_message',
    ellipsis: { tooltip: true },
    render: (row) => row.error_message || '-',
  },
];
</script>
