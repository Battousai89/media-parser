import { createRouter, createWebHistory } from 'vue-router';
import Dashboard from '../views/Dashboard.vue';
import NewRequest from '../views/NewRequest.vue';
import RequestDetail from '../views/RequestDetail.vue';
import MediaList from '../views/MediaList.vue';
import Settings from '../views/Settings.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Dashboard', component: Dashboard },
    { path: '/new', name: 'NewRequest', component: NewRequest },
    { path: '/request/:id', name: 'RequestDetail', component: RequestDetail },
    { path: '/request/:id/media', name: 'MediaList', component: MediaList },
    { path: '/settings', name: 'Settings', component: Settings },
  ],
});

export default router;
