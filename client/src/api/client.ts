import axios, { type AxiosInstance } from 'axios';

let apiInstance: AxiosInstance | null = null;

export function getApiClient(): AxiosInstance {
  if (apiInstance) {
    return apiInstance;
  }

  apiInstance = axios.create({
    baseURL: 'http://localhost:8080',
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  apiInstance.interceptors.request.use((config) => {
    const token = localStorage.getItem('api_token');
    if (token) {
      config.headers['X-Auth-Token'] = token;
    }
    return config;
  });

  apiInstance.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error.response?.status === 401) {
        console.error('Unauthorized');
      }
      return Promise.reject(error);
    }
  );

  return apiInstance;
}
