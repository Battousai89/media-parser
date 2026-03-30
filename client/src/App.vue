<template>
  <n-config-provider :theme="theme">
    <n-message-provider>
      <n-dialog-provider>
        <DefaultLayout />
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { NConfigProvider, NMessageProvider, NDialogProvider } from 'naive-ui';
import DefaultLayout from './layouts/DefaultLayout.vue';

const theme = computed(() => null);
</script>

<style>
:root {
  --app-min-width: 800px;
  --app-min-height: 600px;
}

body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  min-width: var(--app-min-width);
  min-height: var(--app-min-height);
  overflow-x: hidden;
}

#app {
  min-height: 100vh;
  min-width: var(--app-min-width);
}

/* Адаптивность для маленьких экранов */
@media (max-width: 768px) {
  body {
    min-width: 100%;
  }
  
  #app {
    min-width: 100%;
  }
}

/* Масштабирование контента */
.app-container {
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 16px;
}

/* Унификация отступов */
.page-content {
  padding: 24px;
}

@media (max-width: 768px) {
  .page-content {
    padding: 16px;
  }
}

@media (max-width: 480px) {
  .page-content {
    padding: 12px;
  }
}
</style>
