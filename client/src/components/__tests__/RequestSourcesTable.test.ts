import { describe, it, expect, beforeEach, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import { createPinia, setActivePinia } from 'pinia';
import RequestSourcesTable from '../RequestSourcesTable.vue';

vi.mock('../../stores/dictionaries', () => ({
  useDictionariesStore: vi.fn(() => ({
    requestStatuses: [
      { id: 1, code: 'pending', name: 'Ожидает' },
      { id: 2, code: 'processing', name: 'В процессе' },
    ],
    getRequestStatusById: vi.fn((id: number) =>
      id === 1 ? { id: 1, code: 'pending', name: 'Ожидает' } :
      id === 2 ? { id: 2, code: 'processing', name: 'В процессе' } : undefined
    ),
  })),
}));

vi.mock('../StatusBadge.vue', () => ({
  default: {
    name: 'StatusBadge',
    props: ['statusId'],
    template: '<span class="status-badge">Status</span>',
  },
}));

describe('RequestSourcesTable', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  const mockSources = [
    {
      source_id: 101,
      source_name: 'Example.com',
      base_url: 'https://example.com',
      status_id: 1,
      status: 'pending',
      media_count: 100,
      parsed_count: 50,
      retry_count: 0,
      max_retries: 3,
      error_message: null,
    },
    {
      source_id: 102,
      source_name: 'Test.org',
      base_url: 'https://test.org',
      status_id: 2,
      status: 'processing',
      media_count: 200,
      parsed_count: 180,
      retry_count: 1,
      max_retries: 3,
      error_message: 'Connection timeout',
    },
  ];

  it('should render with sources prop', () => {
    const wrapper = shallowMount(RequestSourcesTable, {
      props: { sources: mockSources },
    });

    expect(wrapper.exists()).toBe(true);
  });

  it('should accept sources prop', () => {
    const wrapper = shallowMount(RequestSourcesTable, {
      props: { sources: mockSources },
    });

    expect(wrapper.props('sources')).toEqual(mockSources);
  });
});
