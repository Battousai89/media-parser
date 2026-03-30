// Статусы запросов
export type RequestStatusCode = 'pending' | 'processing' | 'completed' | 'failed' | 'partial';

export interface RequestStatus {
  id: number;
  code: RequestStatusCode;
  name: string;
}

// Типы медиа
export type MediaTypeCode = 'image' | 'video_audio' | 'audio' | 'document' | 'archive' | 'other';

export interface MediaType {
  id: number;
  code: MediaTypeCode;
  name: string;
}

// Статусы источников
export type SourceStatusCode = 'active' | 'inactive' | 'error' | 'blocked';

export interface SourceStatus {
  id: number;
  code: SourceStatusCode;
  name: string;
}
