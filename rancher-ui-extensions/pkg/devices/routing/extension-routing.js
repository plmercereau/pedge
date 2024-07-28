import ViewResource from '@shell/pages/c/_cluster/_product/_resource/_id.vue';
import ViewNamespacedResource from '@shell/pages/c/_cluster/_product/_resource/_namespace/_id.vue';
import CreateResource from '@shell/pages/c/_cluster/_product/_resource/create.vue';
import ListResource from '@shell/pages/c/_cluster/_product/_resource/index.vue';
import { BLANK_CLUSTER, PRODUCT_NAME } from '../config/types';
import Dashboard from '../pages/index.vue';

const meta = {
  product: PRODUCT_NAME,
  cluster: BLANK_CLUSTER,
  pkg:     PRODUCT_NAME
}

const routes = [
  {
    name:      `${ PRODUCT_NAME }-c-cluster`,
    path:      `/${ PRODUCT_NAME }/c/:cluster/dashboard`,
    component: Dashboard,
    meta
  },
  {
    name:      `${ PRODUCT_NAME }-c-cluster-resource`,
    path:      `/${ PRODUCT_NAME }/c/:cluster/:resource`,
    component: ListResource,
    meta
  },
  {
    name:      `${ PRODUCT_NAME }-c-cluster-resource-create`,
    path:      `/${ PRODUCT_NAME }/c/:cluster/:resource/create`,
    component: CreateResource,
    meta
  },
  {
    name:      `${ PRODUCT_NAME }-c-cluster-resource-id`,
    path:      `/${ PRODUCT_NAME }/c/:cluster/:resource/:id`,
    component: ViewResource,
    meta
  },
  {
    name:      `${ PRODUCT_NAME }-c-cluster-resource-namespace-id`,
    path:      `/${ PRODUCT_NAME }/c/:cluster/:resource/:namespace/:id`,
    component: ViewNamespacedResource,
    meta
  },
];

export default routes;