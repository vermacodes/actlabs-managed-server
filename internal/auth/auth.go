package auth

import (
	"actlabs-managed-server/internal/config"
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

type Auth struct {
	Cred                      azcore.TokenCredential
	ActlabsServersTableClient *aztables.Client
}

func NewAuth(appConfig *config.Config) (*Auth, error) {
	var cred azcore.TokenCredential
	var err error

	if appConfig.UseMsi {
		cred, err = azidentity.NewManagedIdentityCredential(&azidentity.ManagedIdentityCredentialOptions{
			ID: azidentity.ClientID(appConfig.ServerManagerClientID),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize managed identity auth: %v", err)
		}
	} else {
		cred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize default auth: %v", err)
		}
	}

	tableClient, err := GetTableClient(
		appConfig.ActlabsSubscriptionID,
		cred,
		appConfig.ActlabsResourceGroup,
		appConfig.ActlabsStorageAccount,
		appConfig.ActlabsServerTableName,
	)
	if err != nil {
		return nil, fmt.Errorf("not able to create table client %w", err)
	}

	return &Auth{
		Cred:                      cred,
		ActlabsServersTableClient: tableClient,
	}, nil
}

func GetStorageAccountKey(subscriptionId string, cred azcore.TokenCredential, resourceGroup string, storageAccountName string) (string, error) {
	client, err := armstorage.NewAccountsClient(subscriptionId, cred, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.ListKeys(context.Background(), resourceGroup, storageAccountName, nil)
	if err != nil {
		return "", err
	}

	if len(resp.Keys) == 0 {
		return "", fmt.Errorf("no storage account key found")
	}

	return *resp.Keys[0].Value, nil
}

func GetTableClient(subscriptionId string, cred azcore.TokenCredential, resourceGroup string, storageAccountName string, tableName string) (*aztables.Client, error) {
	accountKey, err := GetStorageAccountKey(subscriptionId, cred, resourceGroup, storageAccountName)
	if err != nil {
		return &aztables.Client{}, fmt.Errorf("error getting storage account key %w", err)
	}

	sharedKeyCred, err := aztables.NewSharedKeyCredential(storageAccountName, accountKey)
	if err != nil {
		return &aztables.Client{}, fmt.Errorf("error creating shared key credential %w", err)
	}

	tableUrl := "https://" + storageAccountName + ".table.core.windows.net/" + tableName

	return aztables.NewClientWithSharedKey(tableUrl, sharedKeyCred, nil)
}
