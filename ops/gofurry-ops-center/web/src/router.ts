import { createRouter, createWebHashHistory } from 'vue-router'
import AlertsView from './views/AlertsView.vue'
import DeploymentsView from './views/DeploymentsView.vue'
import NodesView from './views/NodesView.vue'
import NodeDetailView from './views/NodeDetailView.vue'
import OverviewView from './views/OverviewView.vue'
import PeersView from './views/PeersView.vue'
import ServicesView from './views/ServicesView.vue'
import SyncView from './views/SyncView.vue'

const router = createRouter({
  history: createWebHashHistory('/admin/'),
  routes: [
    { path: '/', name: 'overview', component: OverviewView },
    { path: '/nodes', name: 'nodes', component: NodesView },
    { path: '/nodes/:id', name: 'node-detail', component: NodeDetailView, props: true },
    { path: '/services', name: 'services', component: ServicesView },
    { path: '/alerts', name: 'alerts', component: AlertsView },
    { path: '/sync', name: 'sync', component: SyncView },
    { path: '/peers', name: 'peers', component: PeersView },
    { path: '/deployments', name: 'deployments', component: DeploymentsView },
  ],
})

export default router
