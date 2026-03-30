import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { RequestListResponse, RequestDetailResponse } from '../types';

export const useRequestsStore = defineStore('requests', () => {
  const requests = ref<RequestListResponse[]>([]);
  const total = ref(0);
  const limit = ref(20);
  const offset = ref(0);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchRequests(params?: { limit?: number; offset?: number }) {
    loading.value = true;
    error.value = null;
    try {
      const { getRequests } = await import('../api/endpoints');
      const data = await getRequests(params);
      requests.value = data.items;
      total.value = data.total;
      limit.value = data.limit;
      offset.value = data.offset;
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Failed to fetch requests';
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
    error,
    fetchRequests,
    fetchRequestById,
  };
});
