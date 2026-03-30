import { describe, it, expect, beforeEach, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import MediaTypeSelect from '../MediaTypeSelect.vue';

vi.mock('../../stores/dictionaries', () => ({
  useDictionariesStore: vi.fn(() => ({
    mediaTypes: [
      { id: 1, code: 'image', name: 'Изображения' },
      { id: 2, code: 'video', name: 'Видео' },
      { id: 3, code: 'audio', name: 'Аудио' },
    ],
    loadDictionaries: vi.fn(),
  })),
}));

describe('MediaTypeSelect', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('should render select component', () => {
    const wrapper = shallowMount(MediaTypeSelect);
    expect(wrapper.exists()).toBe(true);
  });

  it('should accept modelValue prop', () => {
    const wrapper = shallowMount(MediaTypeSelect, {
      props: { modelValue: [1, 2] },
    });

    expect(wrapper.props('modelValue')).toEqual([1, 2]);
  });

  it('should accept placeholder prop', () => {
    const wrapper = shallowMount(MediaTypeSelect, {
      props: { placeholder: 'Выберите типы' },
    });

    expect(wrapper.props('placeholder')).toBe('Выберите типы');
  });

  it('should emit update:modelValue event', () => {
    const wrapper = shallowMount(MediaTypeSelect);

    wrapper.vm.$emit('update:modelValue', [1, 2, 3]);

    expect(wrapper.emitted('update:modelValue')).toBeTruthy();
    expect(wrapper.emitted('update:modelValue')?.[0]?.[0]).toEqual([1, 2, 3]);
  });
});
