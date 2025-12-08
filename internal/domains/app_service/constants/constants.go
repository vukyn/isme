package constants

const (
	AppServiceStatusActive     = 1
	AppServiceStatusInactive   = 2
	AppServiceStatusTerminated = 3

	CtxInfoAuthen     = "authen"
	CtxInfoAppService = "app_service"
)

var AllowedCtxInfos = map[string]struct{}{
	CtxInfoAuthen:     {},
	CtxInfoAppService: {},
}
