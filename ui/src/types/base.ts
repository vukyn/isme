export interface BasePaginationResponse<T> {
	code: number;
	message: string;
	data: {
		page: number;
		total: number;
		items: T[];
	};
}

export interface PageSizeOption {
	value: number;
	label: string;
}
