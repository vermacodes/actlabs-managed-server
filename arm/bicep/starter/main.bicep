param rgName string = 'myResourceGroup'
param storageAccountName string = 'mystorageaccount'
param location string = 'eastus'

targetScope = 'subscription'

resource rg 'Microsoft.Resources/resourceGroups@2021-04-01' = {
  name: rgName
  location: location
}
