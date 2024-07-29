import { IPlugin } from '@shell/core/types';
import { DASHBOARD, PRODUCT_NAME, SCHEMA_IDS } from './config/types';
import { createRoute, rootRoute } from './utils/custom-routing';

export function init($plugin: IPlugin, store: any) {
  const {
    product, configureType, virtualType, basicType, weightType
  } = $plugin.DSL(
    store,
    PRODUCT_NAME
  );

  // registering a top-level product
  product({
    icon:    'gear',
    inStore: 'management',
    weight:  100,
    to:      rootRoute(),
  });

  // creating a custom page
  virtualType({
    labelKey:   'some.translation.key',
    weight:     10,
    namespaced: false,
    name:       DASHBOARD,
    route:      rootRoute(),
  });

  weightType(SCHEMA_IDS.DEVICE_CLUSTERS, 9, true);
  // defining a k8s resource as page
  configureType(SCHEMA_IDS.DEVICE_CLUSTERS, {
    isCreatable: true,
    isEditable:  true,
    isRemovable: true,
    showAge:     true,
    showState:   true,
    canYaml:     true,
    customRoute: createRoute('resource', { resource: SCHEMA_IDS.DEVICE_CLUSTERS }),
  });

  weightType(SCHEMA_IDS.DEVICE_CLASSES, 8, true);
  configureType(SCHEMA_IDS.DEVICE_CLASSES, {
    isCreatable: true,
    isEditable:  true,
    isRemovable: true,
    showAge:     true,
    showState:   true,
    canYaml:     true,
    customRoute: createRoute('resource', { resource: SCHEMA_IDS.DEVICE_CLASSES }),
  });

  weightType(SCHEMA_IDS.DEVICES, 7, true);
  configureType(SCHEMA_IDS.DEVICES, {
    isCreatable: true,
    isEditable:  true,
    isRemovable: true,
    showAge:     true,
    showState:   true,
    canYaml:     true,
    customRoute: createRoute('resource', { resource: SCHEMA_IDS.DEVICES }),
  });

  basicType([DASHBOARD, SCHEMA_IDS.DEVICE_CLUSTERS, SCHEMA_IDS.DEVICE_CLASSES, SCHEMA_IDS.DEVICES]);
}
