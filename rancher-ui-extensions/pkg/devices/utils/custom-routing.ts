import { BLANK_CLUSTER, PRODUCT_NAME } from '../config/types';

export const rootRoute = () => ({
  name:    `${ PRODUCT_NAME }-c-cluster`,
  params: { product: PRODUCT_NAME, cluster: BLANK_CLUSTER },
  meta:   {
    product: PRODUCT_NAME,
    cluster: BLANK_CLUSTER,
    pkg:     PRODUCT_NAME
  },
});

export const createRoute = (name: string, params: Object = {}, meta: Object = {}) => ({
  name:   `${ rootRoute().name }-${ name }`,
  params: {
    ...rootRoute().params,
    ...params
  },
  meta: {
    ...rootRoute().meta,
    ...meta
  }
});
