import { vi } from 'vitest';

const localStorageMock = {
  store: new Map<string, string>(),
  getItem: vi.fn((key: string) => localStorageMock.store.get(key) || null),
  setItem: vi.fn((key: string, value: string) => { localStorageMock.store.set(key, value); }),
  clear: vi.fn(() => { localStorageMock.store.clear(); }),
  removeItem: vi.fn((key: string) => { localStorageMock.store.delete(key); }),
};

Object.defineProperty(global, 'localStorage', {
  value: localStorageMock,
  writable: true,
});
