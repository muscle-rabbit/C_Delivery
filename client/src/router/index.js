import Vue from 'vue'
import Router from 'vue-router'
import WorkerPanel from '@/components/WorkerPanel'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'WorkerPanel',
      component: WorkerPanel
    }
  ]
})
