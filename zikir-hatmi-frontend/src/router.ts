import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('./views/HomeView.vue'),
    },
    {
      path: '/h/:shareCode',
      name: 'hatim',
      component: () => import('./views/HatimView.vue'),
      props: true,
    },
    {
      path: '/h/:shareCode/join',
      name: 'hatim-join',
      component: () => import('./views/JoinView.vue'),
      props: true,
    },
    {
      path: '/y',
      name: 'manage',
      component: () => import('./views/ManageView.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

export default router
