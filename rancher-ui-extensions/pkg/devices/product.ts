import { IPlugin } from "@shell/core/types";
import { DASHBOARD, PRODUCT_NAME, SCHEMA_IDS } from "./config/types";
import { createRoute, rootRoute } from "./utils/custom-routing";

export function init($plugin: IPlugin, store: any) {
  const { product, configureType, virtualType, basicType } = $plugin.DSL(
    store,
    PRODUCT_NAME
  );

  // registering a top-level product
  product({
    icon: "gear",
    inStore: "management",
    weight: 100,
    to: rootRoute(),
  });

  // creating a custom page
  virtualType({
    labelKey: "some.translation.key",
    namespaced: false,
    name: DASHBOARD,
    route: rootRoute(),
  });

  // defining a k8s resource as page
  configureType(SCHEMA_IDS.DEVICE_CLUSTERS, {
    isCreatable: true,
    isEditable: true,
    isRemovable: true,
    showAge: true,
    showState: true,
    canYaml: true,
    customRoute: createRoute("resource", {
      resource: SCHEMA_IDS.DEVICE_CLUSTERS,
    }),
  });

  // registering the defined pages as side-menu entries
  basicType([DASHBOARD, SCHEMA_IDS.DEVICE_CLUSTERS]);
}
