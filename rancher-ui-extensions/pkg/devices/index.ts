import { importTypes } from '@rancher/auto-import';
import { IPlugin, TableColumnLocation } from '@shell/core/types';
import extensionRouting from './routing/extension-routing';

// Init the package
export default function(plugin: IPlugin) {
  // Auto-import model, detail, edit from the folders
  importTypes(plugin);
  // Provide extension metadata from package.json
  // it will grab information such as `name` and `description`
  plugin.metadata = require('./package.json');
  // Load a product
  plugin.addProduct(require('./product'));
  // Add Vue Routes
  plugin.addRoutes(extensionRouting);

  plugin.addTableColumn(
    TableColumnLocation.RESOURCE,
    { resource: ['devices.pedge.io.device'] },
    {
      name:     'device-class-col',
      labelKey: 'devices.deviceClass.label',
      getValue: (row: any) => {
        return row.spec?.deviceClassReference?.name;
      },
      width:  150,
      sort:   ['stateSort', 'nameSort'],
      search: ['stateSort', 'nameSort'],
    }
  );

  plugin.addTableColumn(
    TableColumnLocation.RESOURCE,
    { resource: ['devices.pedge.io.deviceclass'] },
    {
      name:     'device-cluster-col',
      labelKey: 'devices.deviceCluster.label',
      getValue: (row: any) => {
        return row.spec?.deviceClusterReference?.name;
      },
      width:  150,
      sort:   ['stateSort', 'nameSort'],
      search: ['stateSort', 'nameSort'],
    }
  );
}
