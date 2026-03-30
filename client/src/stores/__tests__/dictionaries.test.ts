import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useDictionariesStore } from '../dictionaries';

vi.mock('../../api/client', () => ({
  getApiClient: vi.fn(() => ({
    defaults: { baseURL: '' },
    get: vi.fn(),
  })),
}));

describe('useDictionariesStore', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
  });

  it('should initialize with empty values', () => {
    const store = useDictionariesStore();
    expect(store.mediaTypes).toEqual([]);
    expect(store.requestStatuses).toEqual([]);
    expect(store.sourceStatuses).toEqual([]);
    expect(store.loaded).toBe(false);
    expect(store.loading).toBe(false);
  });

  it('should load dictionaries from API', async () => {
    const { getApiClient } = await import('../../api/client');
    const mockGet = vi.fn().mockResolvedValue({
      data: {
        data: {
          media_types: [{ id: 1, code: 'image', name: 'Изображения' }],
          request_statuses: [{ id: 1, code: 'pending', name: 'Ожидает' }],
          source_statuses: [{ id: 1, code: 'active', name: 'Активен' }],
        },
      },
    });
    vi.mocked(getApiClient).mockReturnValue({
      defaults: { baseURL: '' },
      get: mockGet,
    } as unknown as ReturnType<typeof getApiClient>);

    const store = useDictionariesStore();
    await store.loadDictionaries();

    expect(store.mediaTypes).toHaveLength(1);
    expect(store.requestStatuses).toHaveLength(1);
    expect(store.sourceStatuses).toHaveLength(1);
    expect(store.loaded).toBe(true);
  });

  it('should not load dictionaries twice', async () => {
    const store = useDictionariesStore();
    store.loaded = true;

    await store.loadDictionaries();

    expect(store.loading).toBe(false);
  });

  it('should get media type by ID', () => {
    const store = useDictionariesStore();
    store.mediaTypes = [
      { id: 1, code: 'image', name: 'Изображения' },
      { id: 2, code: 'video_audio', name: 'Видео/Аудио' },
    ];

    const result = store.getMediaTypeById(2);

    expect(result).toEqual({ id: 2, code: 'video_audio', name: 'Видео/Аудио' });
  });

  it('should return undefined for unknown media type ID', () => {
    const store = useDictionariesStore();
    store.mediaTypes = [{ id: 1, code: 'image', name: 'Изображения' }];

    const result = store.getMediaTypeById(999);

    expect(result).toBeUndefined();
  });

  it('should get request status by ID', () => {
    const store = useDictionariesStore();
    store.requestStatuses = [
      { id: 1, code: 'pending', name: 'Ожидает' },
      { id: 2, code: 'completed', name: 'Завершён' },
    ];

    const result = store.getRequestStatusById(2);

    expect(result).toEqual({ id: 2, code: 'completed', name: 'Завершён' });
  });

  it('should return undefined for unknown request status ID', () => {
    const store = useDictionariesStore();
    store.requestStatuses = [{ id: 1, code: 'pending', name: 'Ожидает' }];

    const result = store.getRequestStatusById(999);

    expect(result).toBeUndefined();
  });

  it('should get source status by ID', () => {
    const store = useDictionariesStore();
    store.sourceStatuses = [
      { id: 1, code: 'active', name: 'Активен' },
      { id: 2, code: 'inactive', name: 'Неактивен' },
    ];

    const result = store.getSourceStatusById(1);

    expect(result).toEqual({ id: 1, code: 'active', name: 'Активен' });
  });

  it('should return undefined for unknown source status ID', () => {
    const store = useDictionariesStore();
    store.sourceStatuses = [{ id: 1, code: 'active', name: 'Активен' }];

    const result = store.getSourceStatusById(999);

    expect(result).toBeUndefined();
  });
});
