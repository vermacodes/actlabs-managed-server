package repository

import (
	"actlabs-managed-server/internal/config"
	"actlabs-managed-server/internal/entity"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appcontainers/armappcontainers/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization/v3"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerinstance/armcontainerinstance"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/msi/armmsi"
	"golang.org/x/exp/slog"
)

type serverRepository struct {
	// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity#DefaultAzureCredential
	cred      *azidentity.DefaultAzureCredential
	appConfig *config.Config
}

func NewServerRepository(
	appConfig *config.Config,
) (entity.ServerRepository, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	return &serverRepository{
		cred:      cred,
		appConfig: appConfig,
	}, nil
}

func (s *serverRepository) GetServerEnv(server entity.Server) (string, error) {
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return "", err
	}

	ctx := context.Background()
	res, err := clientFactory.NewManagedEnvironmentsClient().Get(ctx, server.ResourceGroup, server.UserAlias+"-env", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return "", err
	}

	// You could use response here. We use blank identifier for just demo purposes.
	return *res.ID, nil
}

func (s *serverRepository) GetServer(server entity.Server) (entity.Server, error) {
	ctx := context.Background()
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}

	res, err := clientFactory.NewContainerAppsClient().Get(ctx, server.ResourceGroup, server.UserAlias+"-app", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	server.Endpoint = "https://" + *res.Properties.Configuration.Ingress.Fqdn
	server.Status = string(*res.Properties.ProvisioningState)

	return server, nil
}

func (s *serverRepository) GetAzureContainerGroup(server entity.Server) (entity.Server, error) {
	ctx := context.Background()
	clientFactory, err := armcontainerinstance.NewContainerGroupsClient(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}

	res, err := clientFactory.Get(ctx, server.ResourceGroup, server.UserAlias+"-aci", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	server.Endpoint = *res.Properties.IPAddress.Fqdn
	server.Status = string(*res.Properties.ProvisioningState)

	return server, nil
}

func (s *serverRepository) GetUserAssignedManagedIdentity(server entity.Server) (entity.Server, error) {
	ctx := context.Background()
	clientFactory, err := armmsi.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}

	res, err := clientFactory.NewUserAssignedIdentitiesClient().Get(ctx, server.ResourceGroup, server.UserAlias+"-msi", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	server.ManagedIdentityClientId = *res.Properties.ClientID
	server.ManagedIdentityPrincipalId = *res.Properties.PrincipalID
	server.ManagedIdentityResourceId = *res.ID

	return server, nil
}

// https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/azidentity/README.md
// Use for authentication.
func (s *serverRepository) GetSecret(secretName string) (string, error) {
	client, err := azsecrets.NewClient("https://actlabs-kv.vault.azure.net", s.cred, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.GetSecret(context.Background(), secretName, "", nil)
	if err != nil {
		return "", err
	}

	return *resp.Value, nil
}

// https://learn.microsoft.com/en-us/rest/api/containerapps/container-apps/create-or-update?view=rest-containerapps-2023-05-01&tabs=Go
func (s *serverRepository) DeployServer(server entity.Server) (entity.Server, error) {

	ctx := context.Background()
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}
	poller, err := clientFactory.NewContainerAppsClient().BeginCreateOrUpdate(ctx, server.ResourceGroup, server.UserAlias+"-app", armappcontainers.ContainerApp{
		Location: &server.Region,
		Identity: &armappcontainers.ManagedServiceIdentity{
			Type: to.Ptr(armappcontainers.ManagedServiceIdentityTypeUserAssigned),
			UserAssignedIdentities: map[string]*armappcontainers.UserAssignedIdentity{
				server.ManagedIdentityResourceId: {},
			},
		},
		Properties: &armappcontainers.ContainerAppProperties{
			Configuration: &armappcontainers.Configuration{
				Ingress: &armappcontainers.Ingress{
					External:   to.Ptr(true),
					TargetPort: to.Ptr[int32](80),
				},
				// Secrets: []*armappcontainers.Secret{
				// 	{
				// 		Name:  to.Ptr("arm-client-secret"),
				// 		Value: &server.ClientSecret,
				// 	},
				// },
			},
			EnvironmentID: to.Ptr(server.ServerEnvId),
			Template: &armappcontainers.Template{
				Containers: []*armappcontainers.Container{
					{
						Name:  to.Ptr(server.UserAlias + "-app"),
						Image: to.Ptr("ashishvermapu/repro:alpha"),
						Resources: &armappcontainers.ContainerResources{
							CPU:    to.Ptr(1.0),
							Memory: to.Ptr("2Gi"),
						},
						Env: []*armappcontainers.EnvironmentVar{
							{
								Name:  to.Ptr("ARM_USE_MSI"),
								Value: to.Ptr("true"),
							},
							{
								Name:  to.Ptr("USE_MSI"),
								Value: to.Ptr("true"),
							},
							{
								Name:  to.Ptr("PROTECTED_LAB_SECRET"),
								Value: to.Ptr("odEhRZrt!&!94*haZw4YGz7KqzAM!CYg@MxtvbJWu5k9Q!My6Hp$3ffCMwXp$Lo8cgPSM7FSHisi%NxBcYnJ5WVKykReSFDC97^6ZXQ!&HX9SKncaRT4S79ALYp8aFS&"),
							},
							{
								Name:  to.Ptr("ACTLABS_AUTH_URL"),
								Value: to.Ptr("https://actlabs-auth.azurewebsites.net/"),
							},
							{
								Name:  to.Ptr("PORT"),
								Value: to.Ptr("80"),
							},
							{
								Name:  to.Ptr("ROOT_DIR"),
								Value: to.Ptr("/app"),
							},
							{
								Name:  to.Ptr("AZURE_CLIENT_ID"), // https://github.com/microsoft/azure-container-apps/issues/442
								Value: &server.ManagedIdentityClientId,
							},
							// {
							// 	Name:      to.Ptr("ARM_CLIENT_SECRET"),
							// 	SecretRef: to.Ptr("arm-client-secret"),
							// },
							{
								Name:  to.Ptr("ARM_SUBSCRIPTION_ID"),
								Value: &server.SubscriptionId,
							},
							{
								Name:  to.Ptr("AZURE_SUBSCRIPTION_ID"),
								Value: &server.SubscriptionId,
							},
							{
								Name:  to.Ptr("ARM_TENANT_ID"),
								Value: to.Ptr("72f988bf-86f1-41af-91ab-2d7cd011db47"),
							},
							{
								Name:  to.Ptr("ARM_USER_PRINCIPAL_NAME"),
								Value: to.Ptr(server.UserPrincipalName),
							},
							{
								Name:  to.Ptr("LOG_LEVEL"),
								Value: to.Ptr(server.LogLevel),
							},
							{
								Name:  to.Ptr("AUTH_TOKEN_ISS"),
								Value: to.Ptr("https://login.microsoftonline.com/72f988bf-86f1-41af-91ab-2d7cd011db47/v2.0"),
							},
							{
								Name:  to.Ptr("AUTH_TOKEN_AUD"),
								Value: to.Ptr("00399ddd-434c-4b8a-84be-d096cff4f494"),
							},
							{
								Name:  to.Ptr("APPSETTING_WEBSITE_SITE_NAME"), //https://github.com/microsoft/azure-container-apps/issues/502
								Value: to.Ptr("azcli-workaround"),
							},
						},
					}},
				Scale: &armappcontainers.Scale{
					MaxReplicas: to.Ptr[int32](1),
					MinReplicas: to.Ptr[int32](1),
				},
			},
		},
	}, nil)

	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	resp, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return server, err
	}

	server.Endpoint = "https://" + *resp.Properties.Configuration.Ingress.Fqdn
	server.Status = "running"

	return server, nil
}

func (s *serverRepository) DeployServerEnv(server entity.Server) (string, error) {

	ctx := context.Background()
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return "", err
	}
	poller, err := clientFactory.NewManagedEnvironmentsClient().BeginCreateOrUpdate(ctx, server.ResourceGroup, server.UserAlias+"-env", armappcontainers.ManagedEnvironment{
		Location:   to.Ptr(server.Region),
		Properties: &armappcontainers.ManagedEnvironmentProperties{},
	}, nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return "", err
	}

	res, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return "", err
	}
	// You could use response here. We use blank identifier for just demo purposes.
	return *res.ID, nil
}

func (s *serverRepository) DeployAzureContainerGroup(server entity.Server) (entity.Server, error) {

	ctx := context.Background()

	clientFactory, err := armcontainerinstance.NewContainerGroupsClient(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}

	poller, err := clientFactory.BeginCreateOrUpdate(ctx,
		server.ResourceGroup,
		server.UserAlias+"-aci", armcontainerinstance.ContainerGroup{
			Location: to.Ptr(server.Region),
			Identity: &armcontainerinstance.ContainerGroupIdentity{
				Type: to.Ptr(armcontainerinstance.ResourceIdentityTypeUserAssigned),
				UserAssignedIdentities: map[string]*armcontainerinstance.Components10Wh5UdSchemasContainergroupidentityPropertiesUserassignedidentitiesAdditionalproperties{
					server.ManagedIdentityResourceId: {},
				},
			},
			Properties: &armcontainerinstance.ContainerGroupProperties{
				// https://learn.microsoft.com/en-us/azure/container-instances/container-instances-init-container
				InitContainers: []*armcontainerinstance.InitContainerDefinition{
					{
						Name: to.Ptr("init"),
						Properties: &armcontainerinstance.InitContainerPropertiesDefinition{
							Image: to.Ptr("busybox"),
							EnvironmentVariables: []*armcontainerinstance.EnvironmentVariable{
								{
									Name:  to.Ptr("USER_ALIAS"),
									Value: to.Ptr(server.UserAlias),
								},
							},
							VolumeMounts: []*armcontainerinstance.VolumeMount{
								{
									Name:      to.Ptr("emptydir"),
									MountPath: to.Ptr("/etc/caddy"),
								},
							},
							Command: []*string{
								to.Ptr("/bin/sh"),
								to.Ptr("-c"),
								to.Ptr("echo -e \"${USER_ALIAS}-actlabs-aci.eastus.azurecontainer.io {\n\treverse_proxy http://localhost:8881\n}\" > /etc/caddy/Caddyfile"),
							},
						},
					},
				},
				Containers: []*armcontainerinstance.Container{
					{
						Name: to.Ptr("caddy"),
						Properties: &armcontainerinstance.ContainerProperties{
							Image: to.Ptr("ashishvermapu/caddy:latest"),
							Ports: []*armcontainerinstance.ContainerPort{
								{
									Port:     to.Ptr[int32](s.appConfig.HttpPort),
									Protocol: to.Ptr(armcontainerinstance.ContainerNetworkProtocolTCP),
								},
								{
									Port:     to.Ptr[int32](s.appConfig.HttpsPort),
									Protocol: to.Ptr(armcontainerinstance.ContainerNetworkProtocolTCP),
								},
							},
							Resources: &armcontainerinstance.ResourceRequirements{
								Requests: &armcontainerinstance.ResourceRequests{
									CPU:        to.Ptr[float64](s.appConfig.CaddyCPU),
									MemoryInGB: to.Ptr[float64](s.appConfig.CaddyMemory),
								},
							},
							VolumeMounts: []*armcontainerinstance.VolumeMount{
								{
									Name:      to.Ptr("emptydir"),
									MountPath: to.Ptr("/etc/caddy"),
								},
							},
						},
					},
					{
						Name: to.Ptr("actlabs"),
						Properties: &armcontainerinstance.ContainerProperties{
							Image: to.Ptr("ashishvermapu/repro:alpha"),
							Ports: []*armcontainerinstance.ContainerPort{
								{
									Port:     to.Ptr[int32](s.appConfig.ActlabsPort),
									Protocol: to.Ptr(armcontainerinstance.ContainerNetworkProtocolTCP),
								},
							},
							Resources: &armcontainerinstance.ResourceRequirements{
								Requests: &armcontainerinstance.ResourceRequests{
									CPU:        to.Ptr[float64](s.appConfig.ActlabsCPU),
									MemoryInGB: to.Ptr[float64](s.appConfig.ActlabsMemory),
								},
							},
							ReadinessProbe: &armcontainerinstance.ContainerProbe{
								InitialDelaySeconds: to.Ptr[int32](s.appConfig.ActlabsReadinessProbeInitialDelaySeconds),
								PeriodSeconds:       to.Ptr[int32](s.appConfig.ActlabsReadinessProbePeriodSeconds),
								FailureThreshold:    to.Ptr[int32](s.appConfig.ActlabsReadinessProbeFailureThreshold),
								SuccessThreshold:    to.Ptr[int32](s.appConfig.ActlabsReadinessProbeSuccessThreshold),
								TimeoutSeconds:      to.Ptr[int32](s.appConfig.ActlabsReadinessProbeTimeoutSeconds),
								HTTPGet: &armcontainerinstance.ContainerHTTPGet{
									Path:   to.Ptr(s.appConfig.ReadinessProbePath),
									Port:   to.Ptr[int32](s.appConfig.ActlabsPort),
									Scheme: to.Ptr(armcontainerinstance.SchemeHTTP),
								},
							},
							EnvironmentVariables: []*armcontainerinstance.EnvironmentVariable{
								{
									Name:  to.Ptr("ARM_USE_MSI"),
									Value: to.Ptr(strconv.FormatBool(s.appConfig.UseMsi)),
								},
								{
									Name:  to.Ptr("USE_MSI"),
									Value: to.Ptr(strconv.FormatBool(s.appConfig.UseMsi)),
								},
								{
									Name:  to.Ptr("PROTECTED_LAB_SECRET"),
									Value: to.Ptr(s.appConfig.ProtectedLabSecret),
								},
								{
									Name:  to.Ptr("ACTLABS_AUTH_URL"),
									Value: to.Ptr(s.appConfig.ActlabsAuthURL),
								},
								{
									Name:  to.Ptr("PORT"),
									Value: to.Ptr(strconv.Itoa(int(s.appConfig.ActlabsPort))),
								},
								{
									Name:  to.Ptr("ROOT_DIR"),
									Value: to.Ptr(s.appConfig.ActlabsRootDir),
								},
								{
									Name:  to.Ptr("AZURE_CLIENT_ID"), // https://github.com/microsoft/azure-container-apps/issues/442
									Value: &server.ManagedIdentityClientId,
								},
								{
									Name:  to.Ptr("ARM_SUBSCRIPTION_ID"),
									Value: &server.SubscriptionId,
								},
								{
									Name:  to.Ptr("AZURE_SUBSCRIPTION_ID"),
									Value: &server.SubscriptionId,
								},
								{
									Name:  to.Ptr("ARM_TENANT_ID"),
									Value: to.Ptr(s.appConfig.TenantID),
								},
								{
									Name:  to.Ptr("ARM_USER_PRINCIPAL_NAME"),
									Value: to.Ptr(server.UserPrincipalName),
								},
								{
									Name:  to.Ptr("LOG_LEVEL"),
									Value: to.Ptr(server.LogLevel),
								},
								{
									Name:  to.Ptr("AUTH_TOKEN_ISS"),
									Value: to.Ptr(s.appConfig.AuthTokenIss),
								},
								{
									Name:  to.Ptr("AUTH_TOKEN_AUD"),
									Value: to.Ptr(s.appConfig.AuthTokenAud),
								},
							},
							VolumeMounts: []*armcontainerinstance.VolumeMount{
								{
									Name:      to.Ptr("emptydir"),
									MountPath: to.Ptr("/mnt/emptydir"),
								},
							},
						},
					},
				},
				OSType:        to.Ptr(armcontainerinstance.OperatingSystemTypesLinux),
				RestartPolicy: to.Ptr(armcontainerinstance.ContainerGroupRestartPolicyAlways),
				IPAddress: &armcontainerinstance.IPAddress{
					Ports: []*armcontainerinstance.Port{
						{
							Port:     to.Ptr[int32](s.appConfig.HttpPort),
							Protocol: to.Ptr(armcontainerinstance.ContainerGroupNetworkProtocolTCP),
						},
						{
							Port:     to.Ptr[int32](s.appConfig.HttpsPort),
							Protocol: to.Ptr(armcontainerinstance.ContainerGroupNetworkProtocolTCP),
						},
					},
					Type:         to.Ptr(armcontainerinstance.ContainerGroupIPAddressTypePublic),
					DNSNameLabel: to.Ptr(server.UserAlias + "-actlabs-aci"),
				},
				Volumes: []*armcontainerinstance.Volume{
					{
						Name:     to.Ptr("emptydir"),
						EmptyDir: &struct{}{},
					},
				},
			},
		}, nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	resp, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return server, err
	}

	server.Endpoint = *resp.Properties.IPAddress.Fqdn
	server.Status = *resp.Properties.ProvisioningState

	return server, nil
}

func (s *serverRepository) EnsureServerUp(server entity.Server) error {
	// Call the server endpoint to check if it is up
	serverEndpoint := "https://" + server.Endpoint + s.appConfig.ReadinessProbePath
	slog.Info("Checking if server is up: " + serverEndpoint)

	resp, err := http.Get(serverEndpoint)
	if err != nil {
		slog.Error("Failed to make HTTP request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Server is not up. Status code:", resp.StatusCode)
		return errors.New("server is not up")
	}

	return nil
}

func (s *serverRepository) DestroyAzureContainerGroup(server entity.Server) error {

	ctx := context.Background()

	clientFactory, err := armcontainerinstance.NewContainerGroupsClient(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return err
	}

	poller, err := clientFactory.BeginDelete(ctx, server.ResourceGroup, server.UserAlias+"-aci", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return err
	}

	return nil
}

// https://learn.microsoft.com/en-us/rest/api/managedidentity/user-assigned-identities/create-or-update?view=rest-managedidentity-2023-01-31&tabs=Go
func (s *serverRepository) CreateUserAssignedManagedIdentity(server entity.Server) (entity.Server, error) {
	ctx := context.Background()
	clientFactory, err := armmsi.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return server, err
	}

	res, err := clientFactory.NewUserAssignedIdentitiesClient().CreateOrUpdate(ctx, server.ResourceGroup, server.UserAlias+"-msi", armmsi.Identity{
		Location: to.Ptr(server.Region),
	}, nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return server, err
	}

	slog.Info("Managed Identity: " + *res.ID)
	slog.Info("Managed Identity Client ID: " + *res.Properties.ClientID)
	slog.Info("Managed Identity Principal ID: " + *res.Properties.PrincipalID)

	server.ManagedIdentityClientId = *res.Properties.ClientID
	server.ManagedIdentityPrincipalId = *res.Properties.PrincipalID
	server.ManagedIdentityResourceId = *res.ID

	return server, nil
}

func (s *serverRepository) DestroyServerEnv(server entity.Server) error {
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return err
	}

	ctx := context.Background()
	poller, err := clientFactory.NewManagedEnvironmentsClient().BeginDelete(ctx, server.ResourceGroup, server.UserAlias+"-env", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return err
	}

	return nil
}

func (s *serverRepository) DestroyServer(server entity.Server) error {
	ctx := context.Background()
	clientFactory, err := armappcontainers.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return err
	}

	poller, err := clientFactory.NewContainerAppsClient().BeginDelete(ctx, server.ResourceGroup, server.UserAlias+"-app", nil)
	if err != nil {
		slog.Error("failed to finish the request:", err)
		return err
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		slog.Error("failed to pull the result:", err)
		return err
	}

	return nil
}

// verify that user is the owner of the subscription
func (s *serverRepository) IsUserOwner(server entity.Server) (bool, error) {
	slog.Info("Checking if user " + server.UserAlias + " is owner of the subscription " + server.SubscriptionId)

	if server.UserAlias == "" {
		slog.Error("Error: userId is empty")
		return false, errors.New("userId is required")
	}

	if server.SubscriptionId == "" {
		slog.Error("Error: subscriptionId is empty")
		return false, errors.New("subscriptionId is required")
	}

	clientFactory, err := armauthorization.NewClientFactory(server.SubscriptionId, s.cred, nil)
	if err != nil {
		slog.Error("failed to create client:", err)
		return false, err
	}

	filter := "assignedTo('" + server.UserPrincipalId + "')"

	pager := clientFactory.NewRoleAssignmentsClient().NewListForSubscriptionPager(&armauthorization.RoleAssignmentsClientListForSubscriptionOptions{
		Filter:   &filter,
		TenantID: nil,
	})
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			slog.Error("failed to get the next page:", err)
			return false, err
		}
		for _, roleAssignment := range page.Value {
			slog.Debug("Role Assignment: " + *roleAssignment.Properties.PrincipalID + " " + *roleAssignment.Properties.Scope + " " + *roleAssignment.Properties.RoleDefinitionID)
			if *roleAssignment.Properties.PrincipalID == server.UserPrincipalId &&
				*roleAssignment.Properties.Scope == "/subscriptions/"+server.SubscriptionId &&
				*roleAssignment.Properties.RoleDefinitionID == entity.OwnerRoleDefinitionId {
				return true, nil
			}
		}
	}

	return false, nil
}
