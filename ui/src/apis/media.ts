import type { MediaUploadResponse } from "@/types";
import { API_ENDPOINTS } from "@/consts";
import { apiClient } from "@/utils/axios";

/**
 * Upload an avatar image through isme's backend proxy. The proxy forwards to
 * medioa2's public upload API with isme's server-side API key — the browser
 * never holds the key. Returns the stable, servable URL plus the medioa file id.
 */
export const uploadMedia = async (file: File): Promise<MediaUploadResponse["data"]> => {
	const formData = new FormData();
	formData.append("file", file);

	const response = await apiClient.post<MediaUploadResponse>(API_ENDPOINTS.MEDIA_UPLOAD, formData, {
		headers: { "Content-Type": "multipart/form-data" },
	});
	return response.data.data;
};
