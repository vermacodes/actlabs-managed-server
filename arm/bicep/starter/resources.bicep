targetScope = 'resourceGroup'

param storageAccountName string = 'storageaccount'
param location string = 'eastus'

resource storageAccount 'Microsoft.Storage/storageAccounts@2019-06-01' = {
  name: storageAccountName
  location: location
  kind: 'StorageV2'
  sku: {
    name: 'Standard_LRS'
  }
}
