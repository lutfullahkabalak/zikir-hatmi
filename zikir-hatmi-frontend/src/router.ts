import { createRouter, createWebHistory } from 'vue-router'
import HomeView from './views/HomeView.vue'
import HatimView from './views/HatimView.vue'
import JoinView from './views/JoinView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/h/:shareCode',
      name: 'hatim',
      component: HatimView,
      props: true,
    },
    {
      path: '/h/:shareCode/join',
      name: 'hatim-join',
      component: JoinView,
      props: true,
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

export default router
