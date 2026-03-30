import { describe, it, expect, beforeEach, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import StatusBadge from '../StatusBadge.vue';

const mockRequestStatuses = [
  { id: 1, code: 'pending', name: 'Ожидает' },
  { id: 2, code: 'processing', name: 'В процессе' },
  { id: 3, code: 'completed', name: 'Завершён' },
  { id: 4, code: 'failed', name: 'Ошибка' },
];

vi.mock('../../stores/dictionaries', () => ({
  useDictionariesStore: vi.fn(() => ({
    requestStatuses: mockRequestStatuses,
    getRequestStatusById: vi.fn((id: number) =>
      mockRequestStatuses.find((s) => s.id === id)
    ),
  })),
}));

describe('StatusBadge', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('should render with statusId prop', () => {
    const wrapper = shallowMount(StatusBadge, {
      props: { statusId: 1 },
    });
    expect(wrapper.exists()).toBe(true);
  });

  it('should render with statusCode prop', () => {
    const wrapper = shallowMount(StatusBadge, {
      props: { statusCode: 'pending' },
    });
    expect(wrapper.exists()).toBe(true);
  });

  it('should render with default content when no props', () => {
    const wrapper = shallowMount(StatusBadge);
    expect(wrapper.exists()).toBe(true);
  });

  it('should accept statusId prop', () => {
    const wrapper = shallowMount(StatusBadge, {
      props: { statusId: 3 },
    });
    expect(wrapper.props('statusId')).toBe(3);
  });

  it('should accept statusCode prop', () => {
    const wrapper = shallowMount(StatusBadge, {
      props: { statusCode: 'completed' },
    });
    expect(wrapper.props('statusCode')).toBe('completed');
  });
});
