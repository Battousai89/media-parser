import { defineStore } from 'pinia';
import { ref, watch } from 'vue';

export const useSettingsStore = defineStore('settings', () => {
  const apiToken = ref(localStorage.getItem('api_token') || '');
  const apiUrl = ref(localStorage.getItem('api_url') || 'http://localhost:8080');
  const autoRefresh = ref(localStorage.getItem('auto_refresh') === 'true');
  const refreshInterval = ref(Number(localStorage.getItem('refresh_interval') || '5000'));

  watch(apiToken, (newToken) => {
    localStorage.setItem('api_token', newToken);
  });

  watch(apiUrl, (newUrl) => {
    localStorage.setItem('api_url', newUrl);
  });

  watch(autoRefresh, (value) => {
    localStorage.setItem('auto_refresh', String(value));
  });

  watch(refreshInterval, (ms) => {
    localStorage.setItem('refresh_interval', String(ms));
  });

  function setToken(token: string) {
    apiToken.value = token;
  }

  function setUrl(url: string) {
    apiUrl.value = url;
  }

  function setAutoRefresh(value: boolean) {
    autoRefresh.value = value;
  }

  function setRefreshInterval(ms: number) {
    refreshInterval.value = ms;
  }

  return {
    apiToken,
    apiUrl,
    autoRefresh,
    refreshInterval,
    setToken,
    setUrl,
    setAutoRefresh,
    setRefreshInterval,
  };
});
