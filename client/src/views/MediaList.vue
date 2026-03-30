<template>
  <div class="media-list">
    <n-card :title="`Медиа (${total} элементов)`">
      <template #header-extra>
        <n-button @click="fetchMedia" :loading="loading">
          Обновить
        </n-button>
      </template>

      <div v-if="media.length === 0 && !loading" class="empty">
        <n-empty description="Нет медиа" />
      </div>

      <div v-else class="media-grid">
        <MediaCard
          v-for="item in media"
          :key="item.id"
          :url="item.url"
          :media-type-id="item.media_type_id"
          :title="item.title"
          :file-size="item.file_size"
          @download="() => downloadMedia(item)"
        />
      </div>

      <div v-if="media.length > 0" class="pagination">
        <n-pagination
          :page="currentPage"
          :page-count="Math.ceil(total / limit)"
          @update-page="handlePageChange"
          show-size-picker
          :page-sizes="[20, 50, 100]"
        />
      </div>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { NCard, NButton, NEmpty, NPagination, useMessage } from 'naive-ui';
import { getRequestMedia, downloadMediaById } from '../api/endpoints';
import type { MediaItem } from '../types';
import MediaCard from '../components/MediaCard.vue';

const route = useRoute();
const message = useMessage();

const media = ref<MediaItem[]>([]);
const loading = ref(false);
const total = ref(0);
const limit = ref(20);
const currentPage = ref(1);

async function fetchMedia() {
  loading.value = true;
  try {
    const data = await getRequestMedia(route.params.id as string, {
      limit: limit.value,
      offset: (currentPage.value - 1) * limit.value,
    });
    media.value = data?.items ?? [];
    total.value = data?.total ?? 0;
  } catch {
    message.error('Ошибка загрузки медиа');
    media.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

function handlePageChange(page: number) {
  currentPage.value = page;
  fetchMedia();
}

async function downloadMedia(item: MediaItem) {
  try {
    await downloadMediaById(item.id);
    message.success('Скачивание началось');
  } catch (err) {
    console.error('Download error:', err);
    message.error('Ошибка скачивания: ' + (err as Error).message);
  }
}

onMounted(() => {
  fetchMedia();
});
</script>

<style scoped>
.media-list {
  max-width: 1400px;
  margin: 0 auto;
}

.empty {
  padding: 40px;
}

.media-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.pagination {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}
</style>
