import SteveModel from '@shell/plugins/steve/steve-class';
import { clone } from '@shell/utils/object';
import { PRODUCT_NAME } from '../config/types';
import { createRoute } from '../utils/custom-routing';

export default class Resource extends SteveModel {
  get listLocation() {
    return createRoute('resource', { resource: this.type });
  }

  get detailLocation() {
    const schema = this.$getters['schemaFor'](this.type);
    const detailLocation = clone(this._detailLocation);

    detailLocation.params.resource = this.type;

    if (schema?.attributes?.namespaced) {
      detailLocation.name = `${ PRODUCT_NAME }-c-cluster-resource-namespace-id`;
    } else {
      detailLocation.name = `${ PRODUCT_NAME }-c-cluster-resource-id`;
    }

    return detailLocation;
  }

  get parentLocationOverride() {
    return this.listLocation;
  }

  get doneRoute() {
    return this.listLocation.name;
  }
}
