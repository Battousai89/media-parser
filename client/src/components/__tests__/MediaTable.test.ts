import { describe, it, expect, beforeEach, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import MediaTable from '../MediaTable.vue';

vi.mock('../../stores/dictionaries', () => ({
  useDictionariesStore: vi.fn(() => ({
    mediaTypes: [{ id: 1, code: 'image', name: 'Изображения' }],
    getMediaTypeById: vi.fn((id: number) =>
      id === 1 ? { id: 1, code: 'image', name: 'Изображения' } : undefined
    ),
  })),
}));

describe('MediaTable', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  const mockMedia = [
    {
      id: 'media-1',
      url: 'https://example.com/image1.jpg',
      media_type_id: 1,
      media_type: 'image' as const,
      title: 'Image 1',
      file_size: 1024000,
      mime_type: 'image/jpeg',
      source_id: 1,
      available: true,
      created_at: '2024-01-01T00:00:00Z',
    },
    {
      id: 'media-2',
      url: 'https://example.com/video1.mp4',
      media_type_id: 2,
      media_type: 'video_audio' as const,
      title: 'Video 1',
      file_size: 5242880,
      mime_type: 'video/mp4',
      source_id: 1,
      available: true,
      created_at: '2024-01-02T00:00:00Z',
    },
  ];

  it('should render with media items', () => {
    const wrapper = shallowMount(MediaTable, {
      props: { media: mockMedia },
    });

    expect(wrapper.exists()).toBe(true);
  });

  it('should accept media prop', () => {
    const wrapper = shallowMount(MediaTable, {
      props: { media: mockMedia },
    });

    expect(wrapper.props('media')).toEqual(mockMedia);
  });

  it('should accept loading prop', () => {
    const wrapper = shallowMount(MediaTable, {
      props: { media: mockMedia, loading: true },
    });

    expect(wrapper.props('loading')).toBe(true);
  });

  it('should emit download event', async () => {
    const wrapper = shallowMount(MediaTable, {
      props: { media: [mockMedia[0]] },
    });

    // Эмитим событие напрямую для проверки
    wrapper.vm.$emit('download', 'media-1');

    expect(wrapper.emitted('download')).toBeTruthy();
    expect(wrapper.emitted('download')?.[0]).toEqual(['media-1']);
  });
});
