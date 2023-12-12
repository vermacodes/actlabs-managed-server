package entity

const OwnerRoleDefinitionId string = "/subscriptions/da846304-0089-48e0-bfa7-65f68a3eb74f/providers/Microsoft.Authorization/roleDefinitions/8e3af657-a8ff-443c-a75c-2fe8c4bcb635"

type Server struct {
	Endpoint                   string `json:"endpoint"`
	Status                     string `json:"status"`
	Region                     string `json:"region"`
	UserPrincipalId            string `json:"userPrincipalId"`
	UserPrincipalName          string `json:"userPrincipalName"`
	UserAlias                  string `json:"userAlias"`
	ManagedIdentityResourceId  string `json:"managedIdentityResourceId"`
	ManagedIdentityClientId    string `json:"managedIdentityClientId"`
	ManagedIdentityPrincipalId string `json:"managedIdentityPrincipalId"`
	ClientSecret               string `json:"clientSecret"`
	SubscriptionId             string `json:"subscriptionId"`
	ResourceGroup              string `json:"resourceGroup"`
	ServerEnvId                string `json:"serverEnvId"`
	LogLevel                   string `json:"logLevel"`
	LastActivityTime           string `json:"lastActivityTime"`
}

type ServerService interface {
	DeployServer(server Server) (Server, error)
	DestroyServer(server Server) error
	GetServer(server Server) (Server, error)
}

type ServerRepository interface {
	DeployServer(server Server) (Server, error)
	DeployServerEnv(server Server) (string, error)
	CreateUserAssignedManagedIdentity(server Server) (Server, error)

	DestroyServer(server Server) error
	DestroyServerEnv(server Server) error

	GetServer(server Server) (Server, error)
	GetServerEnv(server Server) (string, error)
	GetUserAssignedManagedIdentity(server Server) (Server, error)

	GetSecret(secretName string) (string, error)
	IsUserOwner(server Server) (bool, error)
}
