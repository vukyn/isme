package constants

const (
	// Singleton
	CONTAINER_NAME_CONFIG     = "config"
	CONTAINER_NAME_LOGGER     = "logger"
	CONTAINER_NAME_DB         = "db"
	CONTAINER_NAME_CACHE      = "cache"
	CONTAINER_NAME_MIDDLEWARE = "middleware"

	// Repositories
	CONTAINER_NAME_USER_REPOSITORY         = "user_repository"
	CONTAINER_NAME_USER_SESSION_REPOSITORY = "user_session_repository"
	CONTAINER_NAME_APP_SERVICE_REPOSITORY  = "app_service_repository"

	// Usecases
	CONTAINER_NAME_AUTH_USECASE        = "auth_usecase"
	CONTAINER_NAME_APP_SERVICE_USECASE = "app_service_usecase"
)
