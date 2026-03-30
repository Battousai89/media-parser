import type { RequestStatusCode, MediaTypeCode, SourceStatusCode } from './dictionaries';

// Source — Источник
export interface Source {
  id: number;
  name: string;
  base_url: string;
  status_id: number;
  status: SourceStatusCode;
  updated_at: string;
  created_at: string;
}

export interface SourceItem {
  id: number;
  name: string;
  base_url: string;
  status_id: number;
  status: string;
  updated_at: string;
}

export interface SourceDetail extends SourceItem {
  patterns: PatternItem[] | null;
  created_at: string;
}

// Pattern — Паттерн поиска
export interface Pattern {
  id: number;
  name: string;
  regex: string;
  media_type_id: number;
  media_type: MediaTypeCode;
  priority: number;
  created_at: string;
}

export interface PatternItem {
  id: number;
  name: string;
  regex: string;
  media_type_id: number;
  media_type: MediaTypeCode;
  priority: number;
  created_at: string;
}

// Request — Запрос на парсинг
export interface Request {
  id: string;
  status_id: number;
  status: RequestStatusCode;
  media_type_id: number | null;
  media_type: MediaTypeCode | null;
  limit_count: number;
  offset_count: number;
  priority: number;
  retry_count: number;
  max_retries: number;
  error_message: string | null;
  started_at: string | null;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
  sources?: RequestSourceItem[];
}

export interface RequestSourceItem {
  source_id: number;
  source_name: string;
  base_url: string;
  status_id: number;
  status: string;
  media_count: number;
  parsed_count: number;
  retry_count: number;
  max_retries: number;
  error_message: string | null;
}

// Media — Медиа
export interface Media {
  id: string;
  url: string;
  media_type_id: number;
  media_type: MediaTypeCode;
  title: string | null;
  description: string | null;
  file_size: number | null;
  mime_type: string | null;
  hash: string | null;
  meta: Record<string, unknown> | null;
  available: boolean;
  checked_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface MediaItem {
  id: string;
  url: string;
  media_type_id: number;
  media_type: MediaTypeCode;
  title: string | null;
  file_size: number | null;
  mime_type: string | null;
  source_id: number;
  available: boolean;
  created_at: string;
}

export interface MediaDetail extends MediaItem {
  description: string | null;
  hash: string | null;
  meta: Record<string, unknown> | null;
  checked_at: string | null;
  updated_at: string;
  sources: SourceBrief[] | null;
}

export interface SourceBrief {
  id: number;
  name: string;
  base_url: string;
}

// APIToken — Токен авторизации
export interface ApiToken {
  id: number;
  token: string;
  name: string | null;
  active: boolean;
  expires_at: string | null;
  permissions: TokenPermissions;
  created_at: string;
  last_used_at: string | null;
}

export interface TokenPermissions {
  parse: boolean;
  media_read: boolean;
  media_write: boolean;
  sources_manage: boolean;
  requests_view: boolean;
}

// URLCache — Кэш URL
export interface UrlCache {
  id: number;
  url: string;
  hash: string;
  parsed_at: string | null;
  expires_at: string;
}
