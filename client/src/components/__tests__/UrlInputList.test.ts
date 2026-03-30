import { describe, it, expect, beforeEach } from 'vitest';
import { mount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import UrlInputList from '../UrlInputList.vue';

describe('UrlInputList', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('should render textarea', () => {
    const wrapper = mount(UrlInputList, {
      props: { modelValue: '' },
    });
    expect(wrapper.find('textarea').exists()).toBe(true);
  });

  it('should display URL count', () => {
    const wrapper = mount(UrlInputList, {
      props: { 
        modelValue: 'https://example.com/1\nhttps://example.com/2' 
      },
    });
    expect(wrapper.text()).toContain('URL: 2 / 100');
  });

  it('should emit update:modelValue', async () => {
    const wrapper = mount(UrlInputList, {
      props: { 
        modelValue: '',
        'onUpdate:modelValue': () => {}
      },
    });
    
    const textarea = wrapper.get('textarea');
    await textarea.setValue('https://test.com');
    expect(wrapper.emitted('update:modelValue')).toBeDefined();
  });

  it('should show correct count for empty input', () => {
    const wrapper = mount(UrlInputList, {
      props: { modelValue: '' },
    });
    expect(wrapper.text()).toContain('URL: 0 / 100');
  });
});
