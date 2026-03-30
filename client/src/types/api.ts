import type { RequestStatusCode, MediaTypeCode } from './dictionaries';
import type { RequestSourceItem } from './models';

export interface ApiResponse<T = unknown> {
  success: boolean;
  data: T | null;
  error: ApiError | null;
}

export interface ApiError {
  code: string;
  message: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  limit: number;
  offset: number;
}

export interface RequestListResponse {
  id: string;
  status_id: number;
  status: RequestStatusCode;
  media_type_ids: number[];
  limit_count: number;
  offset_count: number;
  priority: number;
  parsed_count: number;
  sources_count: number;
  created_at: string;
  completed_at: string | null;
}

export interface ParseURLRequest {
  url: string;
  media_type_ids?: number[];
  limit: number;
  offset?: number;
  priority?: number;
}

export interface ParseBatchRequest {
  urls: string[];
  media_type_ids?: number[];
  limit: number;
  offset?: number;
  priority?: number;
}

export interface ParseResponse {
  request_id: string;
  status: RequestStatusCode;
  message: string;
}

export interface RequestDetailResponse {
  id: string;
  status_id: number;
  status: RequestStatusCode;
  media_type_ids: number[];
  limit_count: number;
  offset_count: number;
  priority: number;
  retry_count: number;
  max_retries: number;
  started_at: string | null;
  sources: RequestSourceItem[];
  created_at: string;
  updated_at: string;
  completed_at: string | null;
  error_message: string | null;
}

export interface DictionariesResponse {
  media_types: MediaType[];
  request_statuses: RequestStatus[];
  source_statuses: SourceStatus[];
}

export interface MediaType {
  id: number;
  code: MediaTypeCode;
  name: string;
}

export interface RequestStatus {
  id: number;
  code: RequestStatusCode;
  name: string;
}

export interface SourceStatus {
  id: number;
  code: string;
  name: string;
}

export interface DownloadByUrlRequest {
  url: string;
}
