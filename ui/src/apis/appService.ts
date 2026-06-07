import type {
	AppServiceStatus,
	ListAppServicesRequest,
	ListAppServicesResponse,
	RefreshAppServiceRequest,
	RefreshAppServiceResponse,
	RegisterAppServiceRequest,
	RegisterAppServiceResponse,
	VerifyAppServiceRequest,
	VerifyAppServiceResponse,
} from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/** kuery http/base.Response envelope — every endpoint wraps its payload in `data`. */
interface Envelope<T> {
	code: number;
	message: string;
	data: T;
}

/**
 * POST /api/v1/app-service/register (AuthMiddleware) — creates the app with
 * status forced active and returns the plaintext app_secret EXACTLY ONCE.
 */
export const registerAppService = async (request: RegisterAppServiceRequest): Promise<RegisterAppServiceResponse> => {
	const response = await apiClient.post<Envelope<RegisterAppServiceResponse>>(API_ENDPOINTS.APP_SERVICE_REGISTER, request);
	return response.data.data;
};

/** POST /api/v1/app-service/verify (public) — machine-to-machine credential check. */
export const verifyAppService = async (request: VerifyAppServiceRequest): Promise<VerifyAppServiceResponse> => {
	const response = await apiClient.post<Envelope<VerifyAppServiceResponse>>(API_ENDPOINTS.APP_SERVICE_VERIFY, request);
	return response.data.data;
};

/**
 * POST /api/v1/app-service/refresh (AuthMiddleware) — rotates the secret.
 * Requires the CURRENT plaintext secret; caller must be admin OR the creator.
 * Returns the NEW plaintext secret EXACTLY ONCE.
 */
export const refreshAppServiceSecret = async (request: RefreshAppServiceRequest): Promise<RefreshAppServiceResponse> => {
	const response = await apiClient.post<Envelope<RefreshAppServiceResponse>>(API_ENDPOINTS.APP_SERVICE_REFRESH, request);
	return response.data.data;
};

/**
 * GET /api/v1/app-service (AuthMiddleware + app_service:read) — paginated list.
 * Rows never include app_secret (encrypted at rest, irrecoverable by design).
 */
export const listAppServices = async (request: ListAppServicesRequest): Promise<ListAppServicesResponse> => {
	const response = await apiClient.get<Envelope<ListAppServicesResponse>>(API_ENDPOINTS.APP_SERVICES, {
		params: {
			page: request.page,
			page_size: request.page_size,
			search: request.search || undefined,
			status: request.status || undefined,
			ctx_info: request.ctx_info || undefined,
		},
	});
	const data = response.data.data;
	return { items: data.items ?? [], total: data.total, page: data.page };
};

/**
 * PATCH /api/v1/app-service/:appServiceID/status (AuthMiddleware + app_service:update).
 * Terminated (3) is terminal server-side; same-status updates are a no-op.
 */
export const updateAppServiceStatus = async (appServiceId: string, status: AppServiceStatus): Promise<void> => {
	await apiClient.patch(API_ENDPOINTS.APP_SERVICE_STATUS(appServiceId), { status });
};
