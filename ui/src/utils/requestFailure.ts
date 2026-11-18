/** Shape of a failed request, reduced to what the dialogs show. */
export interface RequestFailure {
	status: number | null;
	message: string;
}

/** Reads an axios-style error into a status + the server's message (if any). */
export const toRequestFailure = (error: unknown): RequestFailure => {
	const err = error as { response?: { status?: number; data?: { message?: string } } };
	return {
		status: err?.response?.status ?? null,
		message: err?.response?.data?.message ?? "",
	};
};
