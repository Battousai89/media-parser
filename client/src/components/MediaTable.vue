<template>
  <n-data-table
    :columns="columns"
    :data="media"
    :pagination="pagination"
    :loading="loading"
    :row-key="(row: MediaItem) => row.id"
  />
</template>

<script setup lang="ts">
import { h } from 'vue';
import { NButton, NTag } from 'naive-ui';
import type { DataTableColumns, DataTableBaseColumn, PaginationProps } from 'naive-ui';
import type { MediaItem } from '../types';
import { useDictionariesStore } from '../stores/dictionaries';

defineProps<{
  media: MediaItem[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  download: [id: string];
  open: [id: string];
}>();

const dictStore = useDictionariesStore();

const getMediaTypeName = (mediaTypeId: number): string => {
  const mt = dictStore.getMediaTypeById(mediaTypeId);
  return mt?.name || `Type #${mediaTypeId}`;
};

const formatFileSize = (size: number | null): string => {
  if (!size) return '-';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let i = 0;
  let s = size;
  while (s >= 1024 && i < units.length - 1) {
    s /= 1024;
    i++;
  }
  return `${s.toFixed(1)} ${units[i]}`;
};

const columns: DataTableColumns<MediaItem> = [
  {
    title: 'Тип',
    key: 'media_type_id',
    width: 120,
    render: (row) => {
      const typeName = getMediaTypeName(row.media_type_id);
      return h(NTag, { type: 'info', size: 'small' }, { default: () => typeName });
    },
  } as DataTableBaseColumn<MediaItem>,
  {
    title: 'URL',
    key: 'url',
    ellipsis: { tooltip: true },
    render: (row) =>
      h(
        'a',
        {
          href: row.url,
          target: '_blank',
          rel: 'noopener noreferrer',
          style: 'color: #18a058; text-decoration: none;',
        },
        { default: () => row.url }
      ),
  },
  {
    title: 'Название',
    key: 'title',
    ellipsis: { tooltip: true },
    render: (row) => row.title || '-',
  },
  {
    title: 'Размер',
    key: 'file_size',
    width: 100,
    render: (row) => formatFileSize(row.file_size),
  },
  {
    title: 'Действия',
    key: 'actions',
    width: 100,
    render: (row) =>
      h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          onClick: () => emit('download', row.id),
        },
        { default: () => 'Скачать' }
      ),
  },
];

const pagination: PaginationProps = {
  pageSize: 20,
  pageSizes: [10, 20, 50, 100],
  showSizePicker: true,
};
</script>
