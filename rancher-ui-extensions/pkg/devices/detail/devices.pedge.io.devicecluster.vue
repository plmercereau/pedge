<script>
import CruResource from '@shell/components/CruResource.vue';
import DetailText from '@shell/components/DetailText';
import Loading from '@shell/components/Loading.vue';
import CreateEditView from '@shell/mixins/create-edit-view';
import Checkbox from '@shell/rancher-components/Form/Checkbox/Checkbox.vue';
import { saferDump } from '@shell/utils/create-yaml';
export default {
  name: 'DeviceClusterDetailView',
  components: {
    Loading,
    CruResource,
    DetailText,
    Checkbox,
  },
  mixins: [CreateEditView],
  props: {
    value: {
      type: Object,
      required: true
    },
    mode: {
      type: String,
      required: true
    },
    resource: {
      type: String,
      required: true
    },
  },
  data() {
    return {
      cloudConfig: typeof this.value.spec === 'string' ? this.value.spec : saferDump(this.value.spec),
      yamlErrors: null
    };
  },
  computed: {
    specType() {
      return this.value?.schema?.resourceFields?.spec?.type;
    },
    
    
  },
  methods: {
    getSchema(path) {
      const spec = this.value?.schema?.resourceFields?.spec;

      return this.$store.getters['management/schemaFor'](`${spec?.type}.${path}`);
    },

  },
};
</script>

<template>
  <Loading v-if="!value" />
  <CruResource v-else :done-route="doneRoute" :can-yaml="true" :mode="mode" :resource="value" :errors="errors"
    @error="e => errors = e" @finish="save" @cancel="done">
    <div class="row mt-40 mb-40">
      <div class="col span-8">
        <h3>{{ t('devices.deviceCluster.artefacts.image.label') }}</h3>
        <DetailText :value="value?.spec?.artefacts?.image?.repository"
          :label="t('devices.deviceCluster.artefacts.image.repository.label')" />
        <DetailText :value="value?.spec?.artefacts?.image?.tag"
          :label="t('devices.deviceCluster.artefacts.image.tag.label')" />

      </div>
    </div>

    <div class="row mt-40 mb-40">
      <div class="col span-8">
        <h3>{{ t('devices.deviceCluster.artefacts.ingress.label') }}</h3>
        <Checkbox data-testid="psp-checkbox" :value="value?.spec?.artefacts?.ingress?.enabled"
          :label="t('devices.deviceCluster.artefacts.ingress.enabled.label')" :mode="mode" />
        <DetailText :value="value?.spec?.artefacts?.ingress?.hostname"
          :label="t('devices.deviceCluster.artefacts.ingress.hostname.label')" />
      </div>
    </div>


  </CruResource>
</template>

<style lang="scss" scoped>
.flex {
  display: flex;
}
</style>
