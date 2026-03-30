import { getApiClient } from './client';
import type {
  ParseBatchRequest,
  ParseResponse,
  PaginatedResponse,
  RequestDetailResponse,
  MediaItem,
  RequestListResponse,
  DictionariesResponse,
  MediaType,
  RequestStatus,
  SourceStatus,
  DownloadByUrlRequest,
} from '../types';

function getApiUrl(): string {
  return localStorage.getItem('api_url') || 'http://localhost:8080';
}

export async function getDictionaries(): Promise<DictionariesResponse> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: DictionariesResponse }>('/api/v1/dictionaries');
  return response.data.data;
}

export async function getMediaTypes(): Promise<MediaType[]> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: MediaType[] }>('/api/v1/dictionaries/media-types');
  return response.data.data;
}

export async function getRequestStatuses(): Promise<RequestStatus[]> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: RequestStatus[] }>('/api/v1/dictionaries/request-statuses');
  return response.data.data;
}

export async function getSourceStatuses(): Promise<SourceStatus[]> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: SourceStatus[] }>('/api/v1/dictionaries/source-statuses');
  return response.data.data;
}

export async function parseBatch(data: ParseBatchRequest): Promise<ParseResponse> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.post<{ success: boolean; data: ParseResponse }>('/api/v1/parse/batch', data);
  return response.data.data;
}

export async function parseUrl(data: {
  url: string;
  media_type_ids?: number[];
  limit?: number;
  offset?: number;
  priority?: number;
}): Promise<ParseResponse> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.post<{ success: boolean; data: ParseResponse }>('/api/v1/parse/url', data);
  return response.data.data;
}

export async function getRequests(params?: {
  limit?: number;
  offset?: number;
  status_id?: number;
}): Promise<PaginatedResponse<RequestListResponse>> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{
    success: boolean;
    data: { items: RequestListResponse[]; total: number; limit: number; offset: number };
  }>('/api/v1/requests', { params });
  return response.data.data;
}

export async function getRequestById(id: string): Promise<RequestDetailResponse> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: RequestDetailResponse }>(`/api/v1/requests/${id}`);
  return response.data.data;
}

export async function getRequestMedia(
  requestId: string,
  params?: { limit?: number; offset?: number; media_type_id?: number }
): Promise<PaginatedResponse<MediaItem>> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{
    success: boolean;
    data: { items: MediaItem[]; total: number; limit: number; offset: number };
  }>(`/api/v1/requests/${requestId}/media`, { params });
  return response.data.data;
}

export async function getMediaById(id: string): Promise<MediaItem> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ success: boolean; data: MediaItem }>(`/api/v1/media/${id}`);
  return response.data.data;
}

/**
 * Скачать медиа по ID.
 * Бэкенд отдаёт файл напрямую из Minio.
 */
export async function downloadMediaById(id: string): Promise<Blob> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.post(`/api/v1/download/${id}`, null, {
    responseType: 'blob',
  });
  return response.data;
}

export async function downloadMediaByUrl(data: DownloadByUrlRequest): Promise<Blob> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.post('/api/v1/download/url', data, {
    responseType: 'blob',
  });
  return response.data;
}

export async function healthCheck(): Promise<{ status: string; timestamp: string }> {
  const API = getApiClient();
  API.defaults.baseURL = getApiUrl();
  const response = await API.get<{ status: string; timestamp: string }>('/health');
  return response.data;
}
