package constants

const (
	// Singleton
	CONTAINER_NAME_CONFIG     = "config"
	CONTAINER_NAME_LOGGER     = "logger"
	CONTAINER_NAME_DB         = "db"
	CONTAINER_NAME_CACHE      = "cache"
	CONTAINER_NAME_MIDDLEWARE = "middleware"
	CONTAINER_NAME_SCHEDULER  = "scheduler"

	// Repositories
	CONTAINER_NAME_USER_REPOSITORY            = "user_repository"
	CONTAINER_NAME_USER_SESSION_REPOSITORY    = "user_session_repository"
	CONTAINER_NAME_APP_SERVICE_REPOSITORY     = "app_service_repository"
	CONTAINER_NAME_ROLE_REPOSITORY            = "role_repository"
	CONTAINER_NAME_USER_INVITATION_REPOSITORY = "user_invitation_repository"
	CONTAINER_NAME_SETTINGS_REPOSITORY        = "settings_repository"

	// Usecases
	CONTAINER_NAME_AUTH_USECASE            = "auth_usecase"
	CONTAINER_NAME_APP_SERVICE_USECASE     = "app_service_usecase"
	CONTAINER_NAME_USER_USECASE            = "user_usecase"
	CONTAINER_NAME_ROLE_USECASE            = "role_usecase"
	CONTAINER_NAME_USER_INVITATION_USECASE = "user_invitation_usecase"
	CONTAINER_NAME_SETTINGS_USECASE        = "settings_usecase"
)
