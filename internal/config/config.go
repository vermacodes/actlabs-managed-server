package config

import (
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
	// Add other configuration fields as needed
}

func NewConfig() *Config {

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file")
	}

	authTokenAud := os.Getenv("AUTH_TOKEN_AUD")
	if authTokenAud == "" {
		slog.Error("AUTH_TOKEN_AUD not set")
		os.Exit(1)
	}
	slog.Info("AUTH_TOKEN_AUD: " + authTokenAud)

	authTokenIss := os.Getenv("AUTH_TOKEN_ISS")
	if authTokenIss == "" {
		slog.Error("AUTH_TOKEN_ISS not set")
		os.Exit(1)
	}
	slog.Info("AUTH_TOKEN_ISS: " + authTokenIss)

	actlabsRootDir := os.Getenv("ACTLABS_ROOT_DIR")
	if actlabsRootDir == "" {
		slog.Error("ROOT_DIR not set")
		os.Exit(1)
	}
	slog.Info("ROOT_DIR: " + actlabsRootDir)

	protectedLabSecret := os.Getenv("PROTECTED_LAB_SECRET")
	if protectedLabSecret == "" {
		slog.Error("PROTECTED_LAB_SECRET not set")
		os.Exit(1)
	}

	useMsiString := os.Getenv("USE_MSI")
	if useMsiString == "" {
		slog.Error("USE_MSI not set")
		os.Exit(1)
	}
	useMsi := false
	if useMsiString == "true" {
		slog.Info("USE_MSI: true")
		useMsi = true
	} else {
		slog.Info("USE_MSI: false")
	}

	actlabsServerUPWaitTimeSeconds := os.Getenv("ACTLABS_SERVER_UP_WAIT_TIME_SECONDS")
	if actlabsServerUPWaitTimeSeconds == "" {
		slog.Error("ACTLABS_SERVER_UP_WAIT_TIME_SECONDS not set defaulting to 180 seconds")
		actlabsServerUPWaitTimeSeconds = "180"
	}

	actlabsPortString := os.Getenv("ACTLABS_PORT")
	if actlabsPortString == "" {
		slog.Error("ACTLABS_PORT not set defaulting to 8881")
		actlabsPortString = "8881"
	}
	actlabsPort, err := strconv.ParseInt(actlabsPortString, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_PORT not set defaulting to 8881")
		actlabsPort = 8881
	}

	actlabsAuthURL := os.Getenv("ACTLABS_AUTH_URL")
	if actlabsAuthURL == "" {
		slog.Error("ACTLABS_AUTH_URL not set defaulting to http://localhost:8880")
		os.Exit(1)
	}

	httpPortString := os.Getenv("HTTP_PORT")
	if httpPortString == "" {
		slog.Error("HTTP_PORT not set defaulting to 80")
		httpPortString = "80"
	}
	httpPort, err := strconv.Atoi(httpPortString)
	if err != nil {
		slog.Error("HTTP_PORT not set defaulting to 80")
		httpPort = 80
	}

	httpsPortString := os.Getenv("HTTPS_PORT")
	if httpsPortString == "" {
		slog.Error("HTTPS_PORT not set defaulting to 443")
		httpsPortString = "443"
	}
	httpsPort, err := strconv.Atoi(httpsPortString)
	if err != nil {
		slog.Error("HTTPS_PORT not set defaulting to 443")
		httpsPort = 443
	}

	readinessProbePath := os.Getenv("READINESS_PROBE_PATH")
	if readinessProbePath == "" {
		slog.Error("READINESS_PROBE_PATH not set defaulting to /status")
		readinessProbePath = "/status"
	}

	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		slog.Error("TENANT_ID not set")
		os.Exit(1)
	}

	actlabsCPU := os.Getenv("ACTLABS_CPU")
	if actlabsCPU == "" {
		slog.Error("ACTLABS_CPU not set defaulting to 0.5")
		actlabsCPU = "0.5"
	}
	actlabsCPUFloat, err := strconv.ParseFloat(actlabsCPU, 32)
	if err != nil {
		slog.Error("ACTLABS_CPU not set defaulting to 0.5")
		actlabsCPUFloat = 0.5
	}

	actlabsMemory := os.Getenv("ACTLABS_MEMORY")
	if actlabsMemory == "" {
		slog.Error("ACTLABS_MEMORY not set defaulting to 0.5")
		actlabsMemory = "0.5"
	}
	actlabsMemoryFloat, err := strconv.ParseFloat(actlabsMemory, 32)
	if err != nil {
		slog.Error("ACTLABS_MEMORY not set defaulting to 0.5")
		actlabsMemoryFloat = 0.5
	}

	caddyCPU := os.Getenv("CADDY_CPU")
	if caddyCPU == "" {
		slog.Error("CADDY_CPU not set defaulting to 0.5")
		caddyCPU = "0.5"
	}
	caddyCPUFloat, err := strconv.ParseFloat(caddyCPU, 32)
	if err != nil {
		slog.Error("CADDY_CPU not set defaulting to 0.5")
		caddyCPUFloat = 0.5
	}

	caddyMemory := os.Getenv("CADDY_MEMORY")
	if caddyMemory == "" {
		slog.Error("CADDY_MEMORY not set defaulting to 0.5")
		caddyMemory = "0.5"
	}
	caddyMemoryFloat, err := strconv.ParseFloat(caddyMemory, 32)
	if err != nil {
		slog.Error("CADDY_MEMORY not set defaulting to 0.5")
		caddyMemoryFloat = 0.5
	}

	actlabsReadinessProbeInitialDelaySeconds := os.Getenv("ACTLABS_READINESS_PROBE_INITIAL_DELAY_SECONDS")
	if actlabsReadinessProbeInitialDelaySeconds == "" {
		slog.Error("ACTLABS_READINESS_PROBE_INITIAL_DELAY_SECONDS not set defaulting to 5")
		actlabsReadinessProbeInitialDelaySeconds = "10"
	}
	actlabsReadinessProbeInitialDelaySecondsInt, err := strconv.ParseInt(actlabsReadinessProbeInitialDelaySeconds, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_READINESS_PROBE_INITIAL_DELAY_SECONDS not set defaulting to 5")
		actlabsReadinessProbeInitialDelaySecondsInt = 10
	}

	actlabsReadinessProbeTimeoutSeconds := os.Getenv("ACTLABS_READINESS_PROBE_TIMEOUT_SECONDS")
	if actlabsReadinessProbeTimeoutSeconds == "" {
		slog.Error("ACTLABS_READINESS_PROBE_TIMEOUT_SECONDS not set defaulting to 5")
		actlabsReadinessProbeTimeoutSeconds = "5"
	}
	actlabsReadinessProbeTimeoutSecondsInt, err := strconv.ParseInt(actlabsReadinessProbeTimeoutSeconds, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_READINESS_PROBE_TIMEOUT_SECONDS not set defaulting to 5")
		actlabsReadinessProbeTimeoutSecondsInt = 5
	}

	actlabsReadinessProbePeriodSeconds := os.Getenv("ACTLABS_READINESS_PROBE_PERIOD_SECONDS")
	if actlabsReadinessProbePeriodSeconds == "" {
		slog.Error("ACTLABS_READINESS_PROBE_PERIOD_SECONDS not set defaulting to 5")
		actlabsReadinessProbePeriodSeconds = "10"
	}
	actlabsReadinessProbePeriodSecondsInt, err := strconv.ParseInt(actlabsReadinessProbePeriodSeconds, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_READINESS_PROBE_PERIOD_SECONDS not set defaulting to 10")
		actlabsReadinessProbePeriodSecondsInt = 10
	}

	actlabsReadinessProbeSuccessThreshold := os.Getenv("ACTLABS_READINESS_PROBE_SUCCESS_THRESHOLD")
	if actlabsReadinessProbeSuccessThreshold == "" {
		slog.Error("ACTLABS_READINESS_PROBE_SUCCESS_THRESHOLD not set defaulting to 1")
		actlabsReadinessProbeSuccessThreshold = "1"
	}
	actlabsReadinessProbeSuccessThresholdInt, err := strconv.ParseInt(actlabsReadinessProbeSuccessThreshold, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_READINESS_PROBE_SUCCESS_THRESHOLD not set defaulting to 1")
		actlabsReadinessProbeSuccessThresholdInt = 1
	}

	actlabsReadinessProbeFailureThreshold := os.Getenv("ACTLABS_READINESS_PROBE_FAILURE_THRESHOLD")
	if actlabsReadinessProbeFailureThreshold == "" {
		slog.Error("ACTLABS_READINESS_PROBE_FAILURE_THRESHOLD not set defaulting to 20")
		actlabsReadinessProbeFailureThreshold = "20"
	}
	actlabsReadinessProbeFailureThresholdInt, err := strconv.ParseInt(actlabsReadinessProbeFailureThreshold, 10, 32)
	if err != nil {
		slog.Error("ACTLABS_READINESS_PROBE_FAILURE_THRESHOLD not set defaulting to 20")
		actlabsReadinessProbeFailureThresholdInt = 20
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

		// Set other fields
	}
}
