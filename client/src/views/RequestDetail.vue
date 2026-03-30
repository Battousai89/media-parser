<template>
  <div class="request-detail">
    <n-card title="Детали запроса" class="mb-4">
      <template #header-extra>
        <div class="header-actions">
          <span class="last-updated">Обновлено: {{ lastUpdated }}</span>
          <n-button @click="fetchRequest" :loading="loading">
            <template #icon>
              <n-icon :component="Refresh" :class="{ spinning: loading }" />
            </template>
            Обновить
          </n-button>
        </div>
      </template>

      <div class="request-info">
        <div class="info-row">
          <div class="info-label">ID:</div>
          <div class="info-value">{{ request?.id }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Статус:</div>
          <div class="info-value">
            <StatusBadge :status-id="request?.status_id" />
          </div>
        </div>
        <div class="info-row" v-if="request?.media_type_ids && request.media_type_ids.length > 0">
          <div class="info-label">Типы медиа:</div>
          <div class="info-value media-types">
            <n-tag
              v-for="id in request.media_type_ids"
              :key="id"
              size="small"
              style="margin-right: 8px;"
            >
              {{ getMediaTypeName(id) }}
            </n-tag>
          </div>
        </div>
        <div class="info-row">
          <div class="info-label">Лимит:</div>
          <div class="info-value">{{ request?.limit_count }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Смещение:</div>
          <div class="info-value">{{ request?.offset_count }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Приоритет:</div>
          <div class="info-value">{{ request?.priority }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Источников:</div>
          <div class="info-value">{{ request?.sources?.length ?? 0 }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Создан:</div>
          <div class="info-value">{{ request?.created_at ? new Date(request.created_at).toLocaleString() : '-' }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Начат:</div>
          <div class="info-value">{{ request?.started_at ? new Date(request.started_at).toLocaleString() : '-' }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Завершён:</div>
          <div class="info-value">{{ request?.completed_at ? new Date(request.completed_at).toLocaleString() : '-' }}</div>
        </div>
        <div class="info-row">
          <div class="info-label">Обновлён:</div>
          <div class="info-value">{{ request?.updated_at ? new Date(request.updated_at).toLocaleString() : '-' }}</div>
        </div>
      </div>

      <div v-if="request?.error_message" class="error-message">
        <n-alert type="error" :title="request.error_message" />
      </div>
    </n-card>

    <div v-if="request?.sources" class="sources-section">
      <h2 class="section-title">
        Источники и медиа
        <span v-if="request.sources.length === 0" class="no-sources">(нет источников)</span>
      </h2>
      <template v-if="request.sources.length > 0">
        <RequestSourceMedia
          v-for="sourceWithMedia in sourcesWithMedia"
          :key="sourceWithMedia.source.source_id"
          :source="sourceWithMedia.source"
          :media="sourceWithMedia.media"
          @download="handleDownload"
        />
      </template>
      <n-empty v-else description="Источники не найдены" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import {
  NCard,
  NAlert,
  NButton,
  NIcon,
  NEmpty,
  NTag,
  useMessage,
} from 'naive-ui';
import { Refresh } from '@vicons/ionicons5';
import { getRequestById, getRequestMedia } from '../api/endpoints';
import { getApiClient } from '../api/client';
import type { RequestDetailResponse, MediaItem } from '../types';
import StatusBadge from '../components/StatusBadge.vue';
import RequestSourceMedia from '../components/RequestSourceMedia.vue';
import { useDictionariesStore } from '../stores/dictionaries';

const route = useRoute();
const message = useMessage();
const dictStore = useDictionariesStore();

const request = ref<RequestDetailResponse | null>(null);
const loading = ref(false);
const allMedia = ref<MediaItem[]>([]);
const mediaLoading = ref(false);
const lastUpdated = ref('');

let pollingInterval: number | undefined;

const sourcesWithMedia = computed(() => {
  if (!request.value?.sources) {
    return [];
  }

  const mediaList = allMedia.value ?? [];
  return request.value.sources.map(source => ({
    source,
    media: mediaList.filter(media => media.source_id === source.source_id),
  }));
});

const getMediaTypeName = (id: number): string => {
  const mt = dictStore.getMediaTypeById(id);
  return mt?.name || `Type #${id}`;
};

async function fetchRequest() {
  loading.value = true;
  try {
    request.value = await getRequestById(route.params.id as string);
    lastUpdated.value = new Date().toLocaleTimeString();
    await fetchMedia();
  } catch (err) {
    console.error('Error fetching request:', err);
    message.error('Ошибка загрузки запроса');
  } finally {
    loading.value = false;
  }
}

async function fetchMedia() {
  mediaLoading.value = true;
  try {
    const data = await getRequestMedia(route.params.id as string, { limit: 1000 });
    console.log('Fetched media:', data);
    allMedia.value = data?.items ?? [];
    console.log('allMedia:', allMedia.value);
  } catch (err) {
    console.error('Error fetching media:', err);
    message.error('Ошибка загрузки медиа');
    allMedia.value = [];
  } finally {
    mediaLoading.value = false;
  }
}

async function handleDownload(id: string) {
  try {
    const API = getApiClient();
    const apiUrl = localStorage.getItem('api_url') || 'http://localhost:8080';
    API.defaults.baseURL = apiUrl;
    
    // Получаем имя файла из заголовка
    const response = await API.post(`/api/v1/download/${id}`, null, {
      responseType: 'blob',
    });
    
    const contentDisposition = response.headers['content-disposition'];
    let filename = `media_${id}`;
    
    if (contentDisposition) {
      const filenameMatch = contentDisposition.match(/filename\*=UTF-8''([^;]+)|filename="?([^"]+)"?/);
      if (filenameMatch) {
        filename = decodeURIComponent(filenameMatch[1] || filenameMatch[2]);
        filename = filename.replace(/"/g, '').trim();
      }
    }
    
    // Стандартное скачивание через blob URL
    const url = window.URL.createObjectURL(response.data);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
    message.success('Скачивание началось');
  } catch (err) {
    console.error('Download error:', err);
    message.error('Ошибка скачивания: ' + (err as Error).message);
  }
}

onMounted(async () => {
  await dictStore.loadDictionaries();
  fetchRequest();
  pollingInterval = window.setInterval(fetchRequest, 3000);
});

onUnmounted(() => {
  if (pollingInterval) {
    clearInterval(pollingInterval);
  }
});
</script>

<style scoped>
.request-detail {
  max-width: 1200px;
  margin: 0 auto;
}

.mb-4 {
  margin-bottom: 16px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}

.last-updated {
  color: #999;
  font-size: 13px;
}

.request-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.info-row:last-child {
  border-bottom: none;
}

.info-label {
  font-weight: 600;
  color: #666;
  min-width: 140px;
  flex-shrink: 0;
}

.info-value {
  flex: 1;
  color: #333;
}

.info-value.media-types {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.error-message {
  margin-top: 16px;
}

.sources-section {
  margin-top: 24px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  color: #333;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
