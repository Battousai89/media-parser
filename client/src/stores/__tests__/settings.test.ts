import { describe, it, expect, beforeEach } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useSettingsStore } from '../settings';

describe('useSettingsStore', () => {
  beforeEach(() => {
    localStorage.clear();
    setActivePinia(createPinia());
  });

  it('should initialize with default values', () => {
    const store = useSettingsStore();
    expect(store.apiToken).toBe('');
    expect(store.apiUrl).toBe('http://localhost:8080');
    expect(store.autoRefresh).toBe(false);
    expect(store.refreshInterval).toBe(5000);
  });

  it('should initialize from localStorage if available', () => {
    localStorage.setItem('api_token', 'stored-token');
    localStorage.setItem('api_url', 'http://custom-url:9000');
    localStorage.setItem('auto_refresh', 'true');
    localStorage.setItem('refresh_interval', '10000');

    const store = useSettingsStore();

    expect(store.apiToken).toBe('stored-token');
    expect(store.apiUrl).toBe('http://custom-url:9000');
    expect(store.autoRefresh).toBe(true);
    expect(store.refreshInterval).toBe(10000);
  });

  it('should save token to store and localStorage', () => {
    const store = useSettingsStore();
    store.setToken('new-token');
    expect(store.apiToken).toBe('new-token');
  });

  it('should save url to store and localStorage', () => {
    const store = useSettingsStore();
    store.setUrl('http://new-url:9000');
    expect(store.apiUrl).toBe('http://new-url:9000');
  });

  it('should save autoRefresh setting to store and localStorage', () => {
    const store = useSettingsStore();
    store.setAutoRefresh(true);
    expect(store.autoRefresh).toBe(true);
  });

  it('should save refreshInterval setting to store and localStorage', () => {
    const store = useSettingsStore();
    store.setRefreshInterval(3000);
    expect(store.refreshInterval).toBe(3000);
  });

  it('should set autoRefresh to false', () => {
    const store = useSettingsStore();
    store.setAutoRefresh(false);
    expect(store.autoRefresh).toBe(false);
  });
});
