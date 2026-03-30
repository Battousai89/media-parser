# План исправлений клиента (frontend)

## Обзор

Документ описывает необходимые изменения в frontend-приложении для соответствия обновлённому API.

---

## 1. Страница списка запросов (Dashboard.vue)

### Проблемы
- **Не получает данные по токену**: API требует `X-Auth-Token`, но токен может не передаваться
- **Некорректное отображение источников**: поле `sources` может отсутствовать в ответе `/api/v1/requests`
- **Некорректные статусы**: статусы должны отображаться через ID, а не строки
- **Нет пагинации**: список не поддерживает `limit`/`offset`
- **Нет обновления по таймеру**: polling работает, но нет индикации последнего обновления

### Решение

**1.1. Обновить типы**
```typescript
// types/api.ts
export interface RequestListResponse {
  id: string;
  status_id: number;
  status: string;
  media_type_ids: number[];
  limit_count: number;
  offset_count: number;
  priority: number;
  created_at: string;
  completed_at: string | null;
  parsed_count?: number;
  sources_count?: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}
```

**1.2. Обновить API endpoint**
```typescript
// api/endpoints.ts
export async function getRequests(params?: {
  limit?: number;
  offset?: number;
  status_id?: number;
}): Promise<PaginatedResponse<RequestListResponse>> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{
    success: boolean;
    data: { items: RequestListResponse[]; total: number }
  }>('/api/v1/requests', { params });
  return response.data.data;
}
```

**1.3. Обновить компонент Dashboard**
- Добавить пагинацию (`n-pagination`)
- Отображать `sources_count` вместо `sources.length`
- Использовать `status_id` для определения цвета статуса
- Добавить индикатор "Последнее обновление: ..."
- Кнопка "Обновить" с анимацией загрузки

---

## 2. Страница нового запроса (NewRequest.vue)

### Проблемы
- **Нет поля `offset`**: нельзя указать смещение
- **Нет поля `priority`**: нельзя задать приоритет (0-10)
- **Нет поля `media_type_ids`**: нельзя выбрать несколько типов через ID
- **Есть чекбокс `download`**: удалён из API
- **Не переходит на детальную страницу**: после создания должен открывать RequestDetail

### Решение

**2.1. Обновить форму**
```vue
<n-form-item label="Типы медиа" path="mediaTypeIds">
  <n-select
    v-model:value="form.mediaTypeIds"
    :options="mediaTypeOptions"
    multiple
    placeholder="Все типы"
  />
</n-form-item>

<n-form-item label="Лимит" path="limit">
  <n-input-number v-model:value="form.limit" :min="1" :max="1000" />
</n-form-item>

<n-form-item label="Offset" path="offset">
  <n-input-number v-model:value="form.offset" :min="0" />
</n-form-item>

<n-form-item label="Приоритет" path="priority">
  <n-slider v-model:value="form.priority" :min="0" :max="10" :marks="{0: 'Низкий', 10: 'Высокий'}" />
</n-form-item>
```

**2.2. Обновить запрос**
```typescript
// Удалить download
const form = reactive({
  urlsText: '',
  mediaTypeIds: [] as number[],
  limit: 10,
  offset: 0,
  priority: 5,
});

// Запрос с media_type_ids
await parseBatch({
  urls,
  media_type_ids: form.mediaTypeIds,
  limit: form.limit,
  offset: form.offset,
  priority: form.priority,
});
```

**2.3. Переход на детальную страницу**
```typescript
// После успешного создания
const result = await parseBatch(...);
router.push({ name: 'RequestDetail', params: { id: result.request_id } });
```

---

## 3. Страница детали запроса (RequestDetail.vue)

### Проблемы
- **Нет автообновления**: статусы не обновляются в реальном времени
- **Нет кнопки обновления**: пользователь не может вручную обновить
- **Нет таблицы медиа**: не показывает найденные медиа
- **Некорректные статусы источников**: должны использовать ID
- **Нет кнопки скачивания**: у каждого медиа должна быть кнопка Download

### Решение

**3.1. Автообновление + кнопка**
```vue
<template>
  <div class="actions">
    <n-button @click="fetchRequest" :loading="loading">
      <template #icon><n-icon :component="Refresh" /></template>
      Обновить
    </n-button>
    <span class="last-updated">Обновлено: {{ lastUpdated }}</span>
  </div>
</template>

<script setup>
let pollingInterval: number | undefined;

onMounted(() => {
  fetchRequest();
  pollingInterval = window.setInterval(fetchRequest, 3000);
});

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval);
});
</script>
```

**3.2. Таблица медиа**
```vue
<n-card title="Найденные медиа" v-if="media.length">
  <n-data-table
    :columns="mediaColumns"
    :data="media"
    :pagination="{ pageSize: 20 }"
  />
</n-card>

<script setup>
const mediaColumns: DataTableColumns = [
  {
    title: 'Тип',
    key: 'media_type_id',
    render: (row) => getMediaTypeName(row.media_type_id),
  },
  {
    title: 'URL',
    key: 'url',
    ellipsis: { tooltip: true },
  },
  {
    title: 'Размер',
    key: 'file_size',
    render: (row) => formatFileSize(row.file_size),
  },
  {
    title: 'Действия',
    key: 'actions',
    render: (row) => h('a', {
      href: '#',
      onClick: () => downloadMedia(row.id),
    }, 'Скачать'),
  },
];
</script>
```

**3.3. API для медиа**
```typescript
// api/endpoints.ts
export async function getRequestMedia(
  requestId: string,
  params?: { limit?: number; offset?: number }
): Promise<PaginatedResponse<MediaItem>> {
  const API = getApiClient();
  const response = await API.get(`/api/v1/requests/${requestId}/media`, { params });
  return response.data.data;
}

export async function downloadMediaById(id: string): Promise<Blob> {
  const API = getApiClient();
  const response = await API.post(`/api/v1/download/${id}`, null, {
    responseType: 'blob',
  });
  return response.data;
}
```

---

## 4. Справочники (Dictionaries)

### Проблемы
- **Хардкод значений**: типы медиа и статусы захардкожены в компонентах
- **Нет синхронизации с бэком**: если на бэке изменятся ID, клиент сломается
- **Использование строк вместо ID**: API теперь работает только с ID

### Решение

**4.1. Загрузка справочников при старте**
```typescript
// stores/dictionaries.ts
export const useDictionariesStore = defineStore('dictionaries', () => {
  const mediaTypes = ref<MediaType[]>([]);
  const requestStatuses = ref<RequestStatus[]>([]);
  const sourceStatuses = ref<SourceStatus[]>([]);
  const loaded = ref(false);

  async function loadDictionaries() {
    if (loaded.value) return;
    const API = getApiClient();
    const response = await API.get('/api/v1/dictionaries');
    const data = response.data.data;
    mediaTypes.value = data.media_types;
    requestStatuses.value = data.request_statuses;
    sourceStatuses.value = data.source_statuses;
    loaded.value = true;
  }

  function getMediaTypeById(id: number): MediaType | undefined {
    return mediaTypes.value.find(mt => mt.id === id);
  }

  function getRequestStatusById(id: number): RequestStatus | undefined {
    return requestStatuses.value.find(rs => rs.id === id);
  }

  function getSourceStatusById(id: number): SourceStatus | undefined {
    return sourceStatuses.value.find(ss => ss.id === id);
  }

  return {
    mediaTypes,
    requestStatuses,
    sourceStatuses,
    loadDictionaries,
    getMediaTypeById,
    getRequestStatusById,
    getSourceStatusById,
  };
});
```

**4.2. Обновить MediaTypeSelect**
```vue
<script setup lang="ts">
import { useDictionariesStore } from '../stores/dictionaries';

const dictStore = useDictionariesStore();
await dictStore.loadDictionaries();

const options = computed(() =>
  dictStore.mediaTypes.map(mt => ({
    label: mt.name,
    value: mt.id,
  }))
);
</script>
```

**4.3. Обновить StatusBadge**
```vue
<script setup lang="ts">
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
  return dictStore.requestStatuses.find(s => s.code === props.statusCode);
});

const color = computed(() => {
  switch (status.value?.code) {
    case 'pending': return 'gray';
    case 'processing': return 'blue';
    case 'completed': return 'green';
    case 'failed': return 'red';
    case 'partial': return 'orange';
    default: return 'default';
  }
});
</script>
```

---

## 5. Stores (Pinia)

### 5.1. Обновить requests store
```typescript
// stores/requests.ts
export const useRequestsStore = defineStore('requests', () => {
  const requests = ref<RequestListResponse[]>([]);
  const total = ref(0);
  const limit = ref(20);
  const offset = ref(0);
  const loading = ref(false);

  async function fetchRequests(params?: { limit?: number; offset?: number }) {
    loading.value = true;
    try {
      const { getRequests } = await import('../api/endpoints');
      const data = await getRequests(params);
      requests.value = data.items;
      total.value = data.total;
      limit.value = data.limit;
      offset.value = data.offset;
    } finally {
      loading.value = false;
    }
  }

  async function fetchRequestById(id: string): Promise<RequestDetailResponse | null> {
    const { getRequestById } = await import('../api/endpoints');
    return await getRequestById(id);
  }

  return {
    requests,
    total,
    limit,
    offset,
    loading,
    fetchRequests,
    fetchRequestById,
  };
});
```

### 5.2. Создать dictionaries store (см. выше)

### 5.3. Обновить settings store
```typescript
// stores/settings.ts
export const useSettingsStore = defineStore('settings', () => {
  const apiToken = ref(localStorage.getItem('api_token') || '');
  const apiUrl = ref(localStorage.getItem('api_url') || 'http://localhost:8080');
  const autoRefresh = ref(localStorage.getItem('auto_refresh') === 'true');
  const refreshInterval = ref(Number(localStorage.getItem('refresh_interval') || '5000'));

  function setToken(token: string) {
    apiToken.value = token;
    localStorage.setItem('api_token', token);
  }

  function setUrl(url: string) {
    apiUrl.value = url;
    localStorage.setItem('api_url', url);
  }

  function setAutoRefresh(value: boolean) {
    autoRefresh.value = value;
    localStorage.setItem('auto_refresh', String(value));
  }

  function setRefreshInterval(ms: number) {
    refreshInterval.value = ms;
    localStorage.setItem('refresh_interval', String(ms));
  }

  return {
    apiToken,
    apiUrl,
    autoRefresh,
    refreshInterval,
    setToken,
    setUrl,
    setAutoRefresh,
    setRefreshInterval,
  };
});
```

---

## 6. Компоненты

### 6.1. Обновить StatusBadge
- Поддержка `statusId` и `statusCode`
- Цвета из справочника

### 6.2. Обновить MediaTypeSelect
- Загрузка опций из справочника
- Возврат ID вместо кодов

### 6.3. Создать MediaTable
```vue
<!-- components/MediaTable.vue -->
<template>
  <n-data-table
    :columns="columns"
    :data="media"
    :pagination="pagination"
    :loading="loading"
  />
</template>

<script setup lang="ts">
import { h } from 'vue';
import { NButton, NTag } from 'naive-ui';
import type { MediaItem } from '../types';

const props = defineProps<{
  media: MediaItem[];
  loading?: boolean;
}>();

const emit = defineEmits<{
  download: [id: string];
}>();

const columns = [
  {
    title: 'Тип',
    key: 'media_type_id',
    render: (row: MediaItem) => getMediaTypeName(row.media_type_id),
  },
  {
    title: 'URL',
    key: 'url',
    ellipsis: { tooltip: true },
  },
  {
    title: 'Размер',
    key: 'file_size',
    render: (row: MediaItem) => formatBytes(row.file_size),
  },
  {
    title: 'Действия',
    key: 'actions',
    render: (row: MediaItem) => h(NButton, {
      size: 'small',
      onClick: () => emit('download', row.id),
    }, { default: () => 'Скачать' }),
  },
];
</script>
```

### 6.4. Создать RequestSourcesTable
```vue
<!-- components/RequestSourcesTable.vue -->
<template>
  <n-data-table
    :columns="columns"
    :data="sources"
    :row-key="(row) => row.id"
  />
</template>

<script setup lang="ts">
import { h } from 'vue';
import { NTag } from 'naive-ui';
import StatusBadge from './StatusBadge.vue';
import type { RequestSourceItem } from '../types';

const props = defineProps<{
  sources: RequestSourceItem[];
}>();

const columns = [
  {
    title: 'Источник',
    key: 'source_id',
    render: (row: RequestSourceItem) => `Source #${row.source_id}`,
  },
  {
    title: 'Статус',
    key: 'status_id',
    render: (row: RequestSourceItem) => h(StatusBadge, { statusId: row.status_id }),
  },
  {
    title: 'Медиа',
    key: 'counts',
    render: (row: RequestSourceItem) => `${row.parsed_count} / ${row.media_count}`,
  },
];
</script>
```

---

## 7. API Client

### 7.1. Обновить endpoints.ts
```typescript
// api/endpoints.ts

// Dictionaries
export async function getDictionaries(): Promise<{
  media_types: MediaType[];
  request_statuses: RequestStatus[];
  source_statuses: SourceStatus[];
}> {
  const API = getApiClient();
  const response = await API.get('/api/v1/dictionaries');
  return response.data.data;
}

export async function getMediaTypes(): Promise<MediaType[]> {
  const API = getApiClient();
  const response = await API.get('/api/v1/dictionaries/media-types');
  return response.data.data;
}

export async function getRequestStatuses(): Promise<RequestStatus[]> {
  const API = getApiClient();
  const response = await API.get('/api/v1/dictionaries/request-statuses');
  return response.data.data;
}

export async function getSourceStatuses(): Promise<SourceStatus[]> {
  const API = getApiClient();
  const response = await API.get('/api/v1/dictionaries/source-statuses');
  return response.data.data;
}

// Parse
export async function parseBatch(data: {
  urls: string[];
  media_type_ids?: number[];
  limit: number;
  offset?: number;
  priority?: number;
}): Promise<ParseResponse> {
  const API = getApiClient();
  const response = await API.post('/api/v1/parse/batch', data);
  return response.data.data;
}

export async function parseUrl(data: {
  url: string;
  media_type_ids?: number[];
  limit: number;
  offset?: number;
  priority?: number;
}): Promise<ParseResponse> {
  const API = getApiClient();
  const response = await API.post('/api/v1/parse/url', data);
  return response.data.data;
}

// Download
export async function downloadMediaById(id: string): Promise<Blob> {
  const API = getApiClient();
  const response = await API.post(`/api/v1/download/${id}`, null, {
    responseType: 'blob',
  });
  return response.data;
}

export async function downloadMediaByUrl(data: { url: string }): Promise<Blob> {
  const API = getApiClient();
  const response = await API.post('/api/v1/download/url', data, {
    responseType: 'blob',
  });
  return response.data;
}
```

---

## 8. Порядок выполнения

| № | Задача | Файлы |
|---|--------|-------|
| 1 | Создать dictionaries store | `stores/dictionaries.ts` |
| 2 | Обновить types | `types/api.ts`, `types/models.ts` |
| 3 | Обновить API endpoints | `api/endpoints.ts` |
| 4 | Обновить Dashboard | `views/Dashboard.vue` |
| 5 | Обновить NewRequest | `views/NewRequest.vue` |
| 6 | Обновить RequestDetail | `views/RequestDetail.vue` |
| 7 | Обновить StatusBadge | `components/StatusBadge.vue` |
| 8 | Обновить MediaTypeSelect | `components/MediaTypeSelect.vue` |
| 9 | Создать MediaTable | `components/MediaTable.vue` |
| 10 | Создать RequestSourcesTable | `components/RequestSourcesTable.vue` |
| 11 | Обновить stores | `stores/requests.ts`, `stores/settings.ts` |
| 12 | Тестирование | |

---

## 9. Тестовые сценарии

### 9.1. Список запросов
1. Открыть Dashboard
2. Проверить отображение статусов через ID
3. Проверить пагинацию
4. Клик на запрос → переход на RequestDetail

### 9.2. Новый запрос
1. Открыть NewRequest
2. Ввести URL (1-100)
3. Выбрать типы медиа (через ID)
4. Указать limit, offset, priority
5. Нажать "Запустить"
6. Проверить переход на RequestDetail

### 9.3. Детали запроса
1. Открыть RequestDetail
2. Проверить автообновление (каждые 3 сек)
3. Нажать "Обновить" вручную
4. Проверить таблицу источников
5. Проверить таблицу медиа
6. Нажать "Скачать" у медиа

### 9.4. Справочники
1. Проверить загрузку `/api/v1/dictionaries`
2. Проверить использование ID в Select
3. Проверить отображение статусов
