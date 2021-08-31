package auth

type AuthContext struct {
	namespace          string
	imagePullSecrets   []string
	serviceAccountName string
}

func (ctx *AuthContext) Namespace() string {
	return ctx.namespace
}

func (ctx *AuthContext) ImagePullSecrets() []string {
	return ctx.imagePullSecrets
}

func (ctx *AuthContext) ServiceAccountName() string {
	return ctx.serviceAccountName
}

func NewAuthContext(namespace string, imagePullSecrets []string, serviceAccountName string) *AuthContext {
	return &AuthContext{
		namespace:          namespace,
		imagePullSecrets:   imagePullSecrets,
		serviceAccountName: serviceAccountName,
	}
}
