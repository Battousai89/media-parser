import { describe, it, expect, beforeEach, vi } from 'vitest';
import * as endpoints from '../endpoints';
import { getApiClient } from '../client';

vi.mock('../client', () => ({
  getApiClient: vi.fn(),
}));

describe('API Endpoints', () => {
  const mockClient = {
    defaults: { baseURL: '' },
    get: vi.fn(),
    post: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.setItem('api_url', 'http://localhost:8080');
    vi.mocked(getApiClient).mockReturnValue(mockClient as unknown as ReturnType<typeof getApiClient>);
  });

  describe('getDictionaries', () => {
    it('should fetch dictionaries from API', async () => {
      const mockData = {
        media_types: [{ id: 1, code: 'image', name: 'Изображения' }],
        request_statuses: [{ id: 1, code: 'pending', name: 'Ожидает' }],
        source_statuses: [{ id: 1, code: 'active', name: 'Активен' }],
      };

      mockClient.get.mockResolvedValueOnce({ data: { success: true, data: mockData } });

      const result = await endpoints.getDictionaries();

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/dictionaries');
      expect(result).toEqual(mockData);
    });
  });

  describe('parseBatch', () => {
    it('should send batch parse request', async () => {
      const mockResponse = {
        request_id: 'req-123',
        status: 'pending' as const,
        message: 'Request created',
      };

      mockClient.post.mockResolvedValueOnce({ data: { success: true, data: mockResponse } });

      const result = await endpoints.parseBatch({
        urls: ['https://example.com'],
        limit: 10,
        offset: 0,
        priority: 5,
      });

      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/parse/batch', {
        urls: ['https://example.com'],
        limit: 10,
        offset: 0,
        priority: 5,
      });
      expect(result).toEqual(mockResponse);
    });

    it('should send batch request with media_type_ids', async () => {
      mockClient.post.mockResolvedValueOnce({
        data: { success: true, data: { request_id: 'req-123', status: 'pending', message: 'OK' } },
      });

      await endpoints.parseBatch({
        urls: ['https://example.com'],
        media_type_ids: [1, 2],
        limit: 10,
        offset: 0,
        priority: 0,
      });

      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/parse/batch', {
        urls: ['https://example.com'],
        media_type_ids: [1, 2],
        limit: 10,
        offset: 0,
        priority: 0,
      });
    });
  });

  describe('getRequests', () => {
    it('should fetch requests with pagination', async () => {
      const mockResponse = {
        items: [
          {
            id: 'req-1',
            status_id: 1,
            status: 'pending' as const,
            media_type_ids: [1],
            limit_count: 10,
            offset_count: 0,
            priority: 5,
            created_at: '2024-01-01T00:00:00Z',
            completed_at: null,
            sources_count: 1,
            parsed_count: 0,
          },
        ],
        total: 1,
        limit: 20,
        offset: 0,
      };

      mockClient.get.mockResolvedValueOnce({ data: { success: true, data: mockResponse } });

      const result = await endpoints.getRequests({ limit: 20, offset: 0 });

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/requests', {
        params: { limit: 20, offset: 0 },
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('getRequestById', () => {
    it('should fetch request by ID', async () => {
      const mockRequest = {
        id: 'req-123',
        status_id: 1,
        status: 'pending' as const,
        media_type_ids: [1],
        limit_count: 10,
        offset_count: 0,
        priority: 5,
        retry_count: 0,
        max_retries: 3,
        sources: [],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        completed_at: null,
        started_at: null,
        error_message: null,
      };

      mockClient.get.mockResolvedValueOnce({ data: { success: true, data: mockRequest } });

      const result = await endpoints.getRequestById('req-123');

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/requests/req-123');
      expect(result).toEqual(mockRequest);
    });
  });

  describe('getRequestMedia', () => {
    it('should fetch media for request', async () => {
      const mockResponse = {
        items: [
          {
            id: 'media-1',
            url: 'https://example.com/image.jpg',
            media_type_id: 1,
            media_type: 'image' as const,
            title: 'Image',
            file_size: 1024,
            mime_type: 'image/jpeg',
            source_id: 1,
            available: true,
            created_at: '2024-01-01T00:00:00Z',
          },
        ],
        total: 1,
        limit: 20,
        offset: 0,
      };

      mockClient.get.mockResolvedValueOnce({ data: { success: true, data: mockResponse } });

      const result = await endpoints.getRequestMedia('req-123', { limit: 20 });

      expect(mockClient.get).toHaveBeenCalledWith('/api/v1/requests/req-123/media', {
        params: { limit: 20 },
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('downloadMediaById', () => {
    it('should download media by ID as blob', async () => {
      const mockBlob = new Blob(['test'], { type: 'application/octet-stream' });

      mockClient.post.mockResolvedValueOnce({ data: mockBlob });

      const result = await endpoints.downloadMediaById('media-123');

      expect(mockClient.post).toHaveBeenCalledWith('/api/v1/download/media-123', null, {
        responseType: 'blob',
      });
      expect(result).toBe(mockBlob);
    });
  });

  describe('healthCheck', () => {
    it('should fetch health status', async () => {
      mockClient.get.mockResolvedValueOnce({
        data: { status: 'ok', timestamp: '2024-01-01T00:00:00Z' },
      });

      const result = await endpoints.healthCheck();

      expect(mockClient.get).toHaveBeenCalledWith('/health');
      expect(result).toEqual({ status: 'ok', timestamp: '2024-01-01T00:00:00Z' });
    });
  });
});
