// import { ANNOTATIONS_TO_IGNORE_REGEX, LABELS_TO_IGNORE_REGEX } from '@shell/config/labels-annotations';
import { _CREATE } from '@shell/config/query-params';
// import { matchesSomeRegex } from '@shell/utils/string';
// import omitBy from 'lodash/omitBy';
// import pickBy from 'lodash/pickBy';
import { set } from 'vue';
import { DEFAULT_NAMESPACE } from '../config/types';
import Resource from './resource';

export default class DeviceCluster extends Resource {
  applyDefaults(_vm, mode) {
    if ( !this.spec || mode === _CREATE ) {
      set(this, 'spec', {});
    }
    if ( !this.metadata || mode === _CREATE ) {
      set(this, 'metadata', { namespace: DEFAULT_NAMESPACE });
    }
  }

  //   setLabels(val, prop = 'labels', isSpec = false) {
  //     if (isSpec && !this.spec) {
  //       this.spec = {};
  //     } else if ( !this.metadata ) {
  //       this.metadata = {};
  //     }

  //     let all = this.metadata[prop] || {};

  //     if (isSpec) {
  //       all = this.spec[prop] || {};
  //     }

  //     const wasIgnored = pickBy(all, (value, key) => {
  //       return matchesSomeRegex(key, LABELS_TO_IGNORE_REGEX);
  //     });

  //     if (isSpec) {
  //       set(this.spec, prop, { ...wasIgnored, ...val });
  //     } else {
  //       set(this.metadata, prop, { ...wasIgnored, ...val });
  //     }
  //   }

  //   setAnnotations(val, prop = 'annotations', isSpec = false) {
  //     if (isSpec && !this.spec) {
  //       this.spec = {};
  //     } else if ( !this.metadata ) {
  //       this.metadata = {};
  //     }

  //     let all = this.metadata[prop] || {};

  //     if (isSpec) {
  //       all = this.spec[prop] || {};
  //     }

  //     const wasIgnored = pickBy(all, (value, key) => {
  //       return matchesSomeRegex(key, ANNOTATIONS_TO_IGNORE_REGEX);
  //     });

  //     if (isSpec) {
  //       set(this.spec, prop, { ...wasIgnored, ...val });
  //     } else {
  //       set(this.metadata, prop, { ...wasIgnored, ...val });
  //     }
  //   }

  //   get machineInventoryLabels() {
  //     const all = this.spec?.machineInventoryLabels || {};

  //     return omitBy(all, (value, key) => {
  //       return matchesSomeRegex(key, LABELS_TO_IGNORE_REGEX);
  //     });
  //   }

  //   get machineInventoryAnnotations() {
  //     const all = this.spec?.machineInventoryAnnotations || {};

  //     return omitBy(all, (value, key) => {
  //       return matchesSomeRegex(key, ANNOTATIONS_TO_IGNORE_REGEX);
  //     });
  //   }

  //   async getMachineRegistrationData() {
  //     const url = `/elemental/registration/${ this.status.registrationToken }`;

  //     try {
  //       const inStore = this.$rootGetters['currentStore']();
  //       const res = await this.$dispatch(`${ inStore }/request`, { url, responseType: 'blob' }, { root: true });
  //       const machineRegFileName = `${ this.metadata.name }_registrationURL.yaml`;

  //       return {
  //         data:     res.data,
  //         fileName: machineRegFileName
  //       };
  //     } catch (e) {
  //       return Promise.reject(e);
  //     }
  //   }

  //   async downloadMachineRegistration() {
  //     try {
  //       const machineReg = await this.getMachineRegistrationData();

//       return downloadFile(machineReg.fileName, machineReg.data, 'text/markdown; charset=UTF-8');
//     } catch (e) {
//       return Promise.reject(e);
//     }
//   }
}
