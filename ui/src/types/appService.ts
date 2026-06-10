/**
 * App service types — mirror internal/domains/app_service/models JSON tags
 * and entity.AppService fields.
 */

/** Maps to constants.AppServiceStatus* (1=active, 2=inactive, 3=terminated · terminal). */
export type AppServiceStatus = 1 | 2 | 3;

/** Maps to constants.AllowedCtxInfos — fixed set. */
export type AppServiceCtxInfo = "authen" | "app_service";

/**
 * Maps to models.AppServiceListItem (GET /api/v1/app-service rows).
 * app_secret is intentionally absent: it is AES-encrypted at rest and never
 * re-displayed by any endpoint.
 */
export interface AppService {
	id: string;
	app_code: string;
	app_name: string;
	redirect_url: string;
	ctx_info: AppServiceCtxInfo;
	status: AppServiceStatus;
	/** Appearance icon key (shared allowlist); empty = neutral fallback. */
	icon: string;
	/** Appearance color palette key; empty = neutral fallback. */
	color: string;
	created_at: string;
	created_by: string;
	created_by_email: string;
	updated_at: string;
	updated_by: string;
}

/** Maps to models.ListRequest query params (GET /api/v1/app-service). */
export interface ListAppServicesRequest {
	page: number;
	page_size: number;
	/** Matches app_name or app_code. */
	search?: string;
	status?: AppServiceStatus;
	ctx_info?: AppServiceCtxInfo;
}

/** Maps to models.ListResponse. */
export interface ListAppServicesResponse {
	items: AppService[];
	total: number;
	page: number;
}

/** Maps to models.RegisterRequest (POST /api/v1/app-service/register). */
export interface RegisterAppServiceRequest {
	app_code: string;
	app_name: string;
	redirect_url: string;
	ctx_info: AppServiceCtxInfo;
	/** Optional appearance icon key; empty = neutral. */
	icon?: string;
	/** Optional appearance color palette key; empty = neutral. */
	color?: string;
}

/**
 * Maps to models.UpdateAppearanceRequest (PATCH /api/v1/app-service/:id).
 * Partial update — only the provided fields change.
 */
export interface UpdateAppServiceAppearanceRequest {
	app_name?: string;
	icon?: string;
	color?: string;
}

/** Maps to models.RegisterResponse — plaintext secret, returned exactly once. */
export interface RegisterAppServiceResponse {
	app_secret: string;
}

/** Maps to models.VerifyRequest (public POST /api/v1/app-service/verify). */
export interface VerifyAppServiceRequest {
	app_code: string;
	ctx_info: AppServiceCtxInfo;
	app_secret: string;
}

/** Maps to models.VerifyResponse. */
export interface VerifyAppServiceResponse {
	ok: boolean;
}

/** Maps to models.RefreshRequest — requires the CURRENT plaintext secret. */
export interface RefreshAppServiceRequest {
	app_code: string;
	app_secret: string;
	ctx_info: AppServiceCtxInfo;
}

/** Maps to models.RefreshResponse — the NEW plaintext secret, returned exactly once. */
export interface RefreshAppServiceResponse {
	app_secret: string;
}
