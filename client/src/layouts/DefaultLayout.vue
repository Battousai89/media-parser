<template>
  <n-layout has-sider class="app-layout">
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="220"
      :native-scrollbar="false"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <div class="logo-container">
        <h1 v-if="!collapsed" class="logo">Media Parser</h1>
      </div>
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="currentRoute"
        @update:value="handleMenuSelect"
      />
    </n-layout-sider>

    <n-layout-content class="app-content">
      <div class="content-wrapper">
        <router-view />
      </div>
    </n-layout-content>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, h, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import {
  NLayout,
  NLayoutSider,
  NLayoutContent,
  NMenu,
  type MenuOption,
} from 'naive-ui';
import {
  HomeOutline,
  AddCircleOutline,
  SettingsOutline,
} from '@vicons/ionicons5';

const router = useRouter();
const route = useRoute();
const collapsed = ref(false);

const currentRoute = computed(() => String(route.name));

const menuOptions: MenuOption[] = [
  {
    label: 'Запросы',
    key: 'Dashboard',
    icon: () => h(HomeOutline),
  },
  {
    label: 'Новый запрос',
    key: 'NewRequest',
    icon: () => h(AddCircleOutline),
  },
  {
    label: 'Настройки',
    key: 'Settings',
    icon: () => h(SettingsOutline),
  },
];

const handleMenuSelect = (key: string) => {
  router.push({ name: key });
};
</script>

<style scoped>
.app-layout {
  min-height: 100vh;
  height: 100%;
}

.app-content {
  background-color: #f5f5f5;
  min-height: 100vh;
}

.content-wrapper {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
  min-height: calc(100vh - 48px);
}

.logo-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.logo {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #18a058;
  white-space: nowrap;
}

/* Адаптивность для мобильных */
@media (max-width: 768px) {
  .content-wrapper {
    padding: 16px;
    min-height: calc(100vh - 32px);
  }
  
  .logo {
    font-size: 16px;
  }
}

@media (max-width: 480px) {
  .content-wrapper {
    padding: 12px;
    min-height: calc(100vh - 24px);
  }
}
</style>
