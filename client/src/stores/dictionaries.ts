import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { MediaType, RequestStatus, SourceStatus } from '../types';
import { getApiClient } from '../api/client';

function getApiUrl(): string {
  return localStorage.getItem('api_url') || 'http://localhost:8080';
}

export const useDictionariesStore = defineStore('dictionaries', () => {
  const mediaTypes = ref<MediaType[]>([]);
  const requestStatuses = ref<RequestStatus[]>([]);
  const sourceStatuses = ref<SourceStatus[]>([]);
  const loaded = ref(false);
  const loading = ref(false);

  async function loadDictionaries() {
    if (loaded.value || loading.value) return;
    loading.value = true;
    try {
      const API = getApiClient();
      API.defaults.baseURL = getApiUrl();
      const response = await API.get('/api/v1/dictionaries');
      const data = response.data.data;
      mediaTypes.value = data.media_types || [];
      requestStatuses.value = data.request_statuses || [];
      sourceStatuses.value = data.source_statuses || [];
      loaded.value = true;
    } finally {
      loading.value = false;
    }
  }

  function getMediaTypeById(id: number): MediaType | undefined {
    return mediaTypes.value.find((mt) => mt.id === id);
  }

  function getMediaTypeByCode(code: string): MediaType | undefined {
    return mediaTypes.value.find((mt) => mt.code === code);
  }

  function getRequestStatusById(id: number): RequestStatus | undefined {
    return requestStatuses.value.find((rs) => rs.id === id);
  }

  function getSourceStatusById(id: number): SourceStatus | undefined {
    return sourceStatuses.value.find((ss) => ss.id === id);
  }

  return {
    mediaTypes,
    requestStatuses,
    sourceStatuses,
    loaded,
    loading,
    loadDictionaries,
    getMediaTypeById,
    getMediaTypeByCode,
    getRequestStatusById,
    getSourceStatusById,
  };
});
