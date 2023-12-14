package service

import (
	"actlabs-managed-server/internal/entity"
	"errors"
	"strings"

	"golang.org/x/exp/slog"
)

type serverService struct {
	serverRepository entity.ServerRepository
}

func NewServerService(serverRepository entity.ServerRepository) entity.ServerService {
	return &serverService{serverRepository: serverRepository}
}

func (s *serverService) DeployServer(server entity.Server) (entity.Server, error) {

	// Validate input.
	if err := s.Validate(server); err != nil {
		slog.Error("Error:", err)
		return server, err
	}

	s.ServerDefaults(&server)          // Set defaults.
	s.ContainerAppEnvironment(&server) // Create container app environment if it doesn't exist.
	s.UserAssignedIdentity(&server)    // Managed Identity

	server, err := s.serverRepository.DeployServer(server)
	if err != nil {
		slog.Error("Error:", err)
		return server, err
	}

	//redact the secrets
	//server.ClientSecret = "REDACTED"

	return server, nil
}

func (s *serverService) DestroyServer(server entity.Server) error {

	if err := s.Validate(server); err != nil {
		slog.Error("Error:", err)
		return err
	}

	s.ServerDefaults(&server)

	if err := s.serverRepository.DestroyServer(server); err != nil {
		slog.Error("Error:", err)
		return err
	}

	if server.DeleteServerEnv {
		if err := s.serverRepository.DestroyServerEnv(server); err != nil {
			slog.Error("Error:", err)
			return err
		}
	}

	return nil
}

func (s *serverService) GetServer(server entity.Server) (entity.Server, error) {
	// Validate input.
	if err := s.Validate(server); err != nil {
		slog.Error("Error:", err)
		return server, err
	}

	s.ServerDefaults(&server)          // Set defaults.
	s.ContainerAppEnvironment(&server) // Create container app environment if it doesn't exist.

	return s.serverRepository.GetServer(server)
}

func (s *serverService) Validate(server entity.Server) error {
	if server.UserPrincipalName == "" || server.UserPrincipalId == "" || server.SubscriptionId == "" {
		slog.Error("Error: userPrincipalName, userPrincipalId, and subscriptionId are required")
		return errors.New("missing required information")
	}

	if server.UserAlias == "" {
		server.UserAlias = strings.Split(server.UserPrincipalName, "@")[0]
	}

	ok, err := s.serverRepository.IsUserOwner(server)
	if err != nil {
		slog.Error("Error:", err)
		return err
	}
	if !ok {
		slog.Error("Error: user is not the owner of the subscription")
		return errors.New("insufficient permissions")
	}

	return nil
}

func (s *serverService) ServerDefaults(server *entity.Server) {
	if server.UserAlias == "" {
		server.UserAlias = strings.Split(server.UserPrincipalName, "@")[0]
	}

	if server.LogLevel == "" {
		server.LogLevel = "0"
	}

	if server.Region == "" {
		server.Region = "East US"
	}

	if server.ResourceGroup == "" {
		server.ResourceGroup = "repro-project"
	}
}

func (s *serverService) ContainerAppEnvironment(server *entity.Server) error {
	envId, err := s.serverRepository.GetServerEnv(*server)
	if err != nil {
		slog.Info("Container apps environment not found, creating...")
	}
	if envId == "" {
		envId, err = s.serverRepository.DeployServerEnv(*server)
		if err != nil {
			slog.Error("Error:", err)
			return err
		}
	}

	server.ServerEnvId = envId

	return nil
}

func (s *serverService) UserAssignedIdentity(server *entity.Server) error {

	var err error
	*server, err = s.serverRepository.GetUserAssignedManagedIdentity(*server)
	if err != nil {
		slog.Info("Managed Identity not found, creating...")
	}

	*server, err = s.serverRepository.CreateUserAssignedManagedIdentity(*server)
	if err != nil {
		slog.Error("Error:", err)
		return err
	}

	return nil
}
