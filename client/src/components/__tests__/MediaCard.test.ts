import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import MediaCard from '../MediaCard.vue';

vi.mock('../../stores/dictionaries', () => ({
  useDictionariesStore: vi.fn(() => ({
    mediaTypes: [
      { id: 1, code: 'image', name: 'Изображения' },
      { id: 2, code: 'video', name: 'Видео' },
    ],
    getMediaTypeById: vi.fn((id: number) => {
      if (id === 1) return { id: 1, code: 'image', name: 'Изображения' };
      if (id === 2) return { id: 2, code: 'video', name: 'Видео' };
      return undefined;
    }),
  })),
}));

describe('MediaCard', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('should render media card with image type', () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/image.jpg',
        mediaTypeId: 1,
        title: 'Test Image',
      },
    });
    expect(wrapper.text()).toContain('Test Image');
    expect(wrapper.text()).toContain('Изображения');
  });

  it('should render media card with video type', () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/video.mp4',
        mediaTypeId: 2,
      },
    });
    expect(wrapper.text()).toContain('Видео');
  });

  it('should display file size in KB', () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/file.jpg',
        mediaTypeId: 1,
        fileSize: 2048,
      },
    });
    expect(wrapper.text()).toContain('2.0 KB');
  });

  it('should display file size in MB', () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/file.jpg',
        mediaTypeId: 1,
        fileSize: 2097152,
      },
    });
    expect(wrapper.text()).toContain('2.0 MB');
  });

  it('should emit download event on button click', async () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/file.jpg',
        mediaTypeId: 1,
      },
    });

    await wrapper.find('button').trigger('click');
    expect(wrapper.emitted('download')).toBeTruthy();
  });

  it('should show default title for null title', () => {
    const wrapper = mount(MediaCard, {
      props: {
        url: 'https://example.com/file.jpg',
        mediaTypeId: 1,
        title: null,
      },
    });
    expect(wrapper.text()).toContain('Без названия');
  });
});
