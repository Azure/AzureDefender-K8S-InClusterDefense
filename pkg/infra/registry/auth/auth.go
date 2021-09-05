package auth

type AuthType string
const(
	K8SAuth = "K8SAuth"
	ACRAuth = "ACRAuth"
)

type AuthContext struct {
	Namespace          string
	ImagePullSecrets   []string
	ServiceAccountName string
	RegistryEndpoint string
}

type AuthConfig struct {
	Context *AuthContext
	AuthType AuthType
}

