package auth

type AuthContext struct {
	Namespace          string
	ImagePullSecrets   []string
	ServiceAccountName string
	RegistryEndpoint string
}

