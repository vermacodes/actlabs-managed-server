package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

type Config struct {
	AuthTokenAud                             string
	AuthTokenIss                             string
	ProtectedLabSecret                       string
	UseMsi                                   bool
	ActlabsPort                              int32
	ActlabsAuthURL                           string
	ActlabsRootDir                           string
	ActlabsServerUPWaitTimeSeconds           string
	ActlabsCPU                               float64
	ActlabsMemory                            float64
	ActlabsReadinessProbeInitialDelaySeconds int32
	ActlabsReadinessProbeTimeoutSeconds      int32
	ActlabsReadinessProbePeriodSeconds       int32
	ActlabsReadinessProbeSuccessThreshold    int32
	ActlabsReadinessProbeFailureThreshold    int32
	CaddyCPU                                 float64
	CaddyMemory                              float64
	HttpPort                                 int32
	HttpsPort                                int32
	ReadinessProbePath                       string
	TenantID                                 string
	ServerManagerClientID                    string
	ActlabsSubscriptionID                    string
	ActlabsResourceGroup                     string
	ActlabsStorageAccount                    string
	ActlabsServerTableName                   string
	// Add other configuration fields as needed
}

func NewConfig() (*Config, error) {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	authTokenAud := getEnv("AUTH_TOKEN_AUD")
	if authTokenAud == "" {
		return nil, fmt.Errorf("AUTH_TOKEN_AUD not set")
	}

	authTokenIss := getEnv("AUTH_TOKEN_ISS")
	if authTokenIss == "" {
		return nil, fmt.Errorf("AUTH_TOKEN_ISS not set")
	}

	actlabsRootDir := getEnv("ACTLABS_ROOT_DIR")
	if actlabsRootDir == "" {
		return nil, fmt.Errorf("ROOT_DIR not set")
	}

	protectedLabSecret := getEnv("PROTECTED_LAB_SECRET")
	if protectedLabSecret == "" {
		return nil, fmt.Errorf("PROTECTED_LAB_SECRET not set")
	}

	useMsi, err := strconv.ParseBool(getEnvWithDefault("USE_MSI", "false"))
	if err != nil {
		return nil, err
	}

	actlabsServerUPWaitTimeSeconds := getEnvWithDefault("ACTLABS_SERVER_UP_WAIT_TIME_SECONDS", "180")
	if actlabsServerUPWaitTimeSeconds == "" {
		return nil, fmt.Errorf("ACTLABS_SERVER_UP_WAIT_TIME_SECONDS not set")
	}

	actlabsPort, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_PORT", "8881"), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("ACTLABS_PORT not set")
	}

	actlabsAuthURL := getEnv("ACTLABS_AUTH_URL")
	if actlabsAuthURL == "" {
		return nil, fmt.Errorf("ACTLABS_AUTH_URL not set")
	}

	httpPort, err := strconv.Atoi(getEnvWithDefault("HTTP_PORT", "80"))
	if err != nil {
		return nil, fmt.Errorf("HTTP_PORT not set")
	}

	httpsPort, err := strconv.Atoi(getEnvWithDefault("HTTPS_PORT", "443"))
	if err != nil {
		return nil, fmt.Errorf("HTTPS_PORT not set")
	}

	readinessProbePath := getEnvWithDefault("READINESS_PROBE_PATH", "/status")
	if readinessProbePath == "" {
		return nil, fmt.Errorf("READINESS_PROBE_PATH not set")
	}

	tenantID := getEnv("TENANT_ID")
	if tenantID == "" {
		return nil, fmt.Errorf("TENANT_ID not set")
	}

	actlabsCPUFloat, err := strconv.ParseFloat(getEnvWithDefault("ACTLABS_CPU", "0.5"), 32)
	if err != nil {
		return nil, err
	}

	actlabsMemoryFloat, err := strconv.ParseFloat(getEnvWithDefault("ACTLABS_MEMORY", "0.5"), 32)
	if err != nil {
		return nil, err
	}

	caddyCPUFloat, err := strconv.ParseFloat(getEnvWithDefault("CADDY_CPU", "0.5"), 32)
	if err != nil {
		return nil, err
	}

	caddyMemoryFloat, err := strconv.ParseFloat(getEnvWithDefault("CADDY_MEMORY", "0.5"), 32)
	if err != nil {
		return nil, err
	}

	actlabsReadinessProbeInitialDelaySecondsInt, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_READINESS_PROBE_INITIAL_DELAY_SECONDS", "10"), 10, 32)
	if err != nil {
		return nil, err
	}

	actlabsReadinessProbeTimeoutSecondsInt, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_READINESS_PROBE_TIMEOUT_SECONDS", "5"), 10, 32)
	if err != nil {
		return nil, err
	}

	actlabsReadinessProbePeriodSecondsInt, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_READINESS_PROBE_PERIOD_SECONDS", "10"), 10, 32)
	if err != nil {
		return nil, err
	}

	actlabsReadinessProbeSuccessThresholdInt, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_READINESS_PROBE_SUCCESS_THRESHOLD", "1"), 10, 32)
	if err != nil {
		return nil, err
	}

	actlabsReadinessProbeFailureThresholdInt, err := strconv.ParseInt(getEnvWithDefault("ACTLABS_READINESS_PROBE_FAILURE_THRESHOLD", "20"), 10, 32)
	if err != nil {
		return nil, err
	}

	serverManagerClientID := getEnv("SERVER_MANAGER_CLIENT_ID")
	if serverManagerClientID == "" {
		return nil, fmt.Errorf("SERVER_MANAGER_CLIENT_ID not set")
	}

	actlabsSubscriptionID := getEnv("ACTLABS_SUBSCRIPTION_ID")
	if actlabsSubscriptionID == "" {
		return nil, fmt.Errorf("ACTLABS_SUBSCRIPTION_ID not set")
	}

	actlabsResourceGroup := getEnv("ACTLABS_RESOURCE_GROUP")
	if actlabsResourceGroup == "" {
		return nil, fmt.Errorf("ACTLABS_RESOURCE_GROUP not set")
	}

	actlabsStorageAccount := getEnv("ACTLABS_STORAGE_ACCOUNT")
	if actlabsStorageAccount == "" {
		return nil, fmt.Errorf("ACTLABS_STORAGE_ACCOUNT not set")
	}

	actlabsServerTableName := getEnv("ACTLABS_SERVER_TABLE_NAME")
	if actlabsServerTableName == "" {
		return nil, fmt.Errorf("ACTLABS_SERVER_TABLE_NAME not set")
	}

	// Retrieve other environment variables and check them as needed

	return &Config{
		AuthTokenAud:                             authTokenAud,
		AuthTokenIss:                             authTokenIss,
		ActlabsRootDir:                           actlabsRootDir,
		ProtectedLabSecret:                       protectedLabSecret,
		UseMsi:                                   useMsi,
		ActlabsServerUPWaitTimeSeconds:           actlabsServerUPWaitTimeSeconds,
		ActlabsPort:                              int32(actlabsPort),
		ActlabsAuthURL:                           actlabsAuthURL,
		HttpPort:                                 int32(httpPort),
		HttpsPort:                                int32(httpsPort),
		ReadinessProbePath:                       readinessProbePath,
		TenantID:                                 tenantID,
		ActlabsCPU:                               actlabsCPUFloat,
		ActlabsMemory:                            actlabsMemoryFloat,
		CaddyCPU:                                 caddyCPUFloat,
		CaddyMemory:                              caddyMemoryFloat,
		ActlabsReadinessProbeInitialDelaySeconds: int32(actlabsReadinessProbeInitialDelaySecondsInt),
		ActlabsReadinessProbeTimeoutSeconds:      int32(actlabsReadinessProbeTimeoutSecondsInt),
		ActlabsReadinessProbePeriodSeconds:       int32(actlabsReadinessProbePeriodSecondsInt),
		ActlabsReadinessProbeSuccessThreshold:    int32(actlabsReadinessProbeSuccessThresholdInt),
		ActlabsReadinessProbeFailureThreshold:    int32(actlabsReadinessProbeFailureThresholdInt),
		ServerManagerClientID:                    serverManagerClientID,
		ActlabsSubscriptionID:                    actlabsSubscriptionID,
		ActlabsResourceGroup:                     actlabsResourceGroup,
		ActlabsStorageAccount:                    actlabsStorageAccount,
		ActlabsServerTableName:                   actlabsServerTableName,
		// Set other fields
	}, nil
}

// Helper function to retrieve the value and log it
func getEnv(env string) string {
	value := os.Getenv(env)
	slog.Info("environment variable", slog.String("name", env), slog.String("value", value))
	return value
}

// Helper function to retrieve the value, if none found, set default and log it
func getEnvWithDefault(env string, defaultValue string) string {
	value := os.Getenv(env)
	if value == "" {
		value = defaultValue
	}
	slog.Info("environment variable", slog.String("name", env), slog.String("value", value))
	return value
}
