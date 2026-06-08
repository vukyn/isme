// navigator.clipboard only exists in secure contexts (https / localhost) —
// *.local dev hosts fall back to a hidden textarea + execCommand copy, which
// browsers may still reject outside a user gesture (caller surfaces the error).
export const copyToClipboard = async (text: string): Promise<void> => {
	if (navigator.clipboard) {
		await navigator.clipboard.writeText(text);
		return;
	}

	const textarea = document.createElement("textarea");
	textarea.value = text;
	textarea.style.position = "fixed";
	textarea.style.opacity = "0";
	// Mount inside the active dialog (if any) so its focus trap doesn't steal
	// focus from the textarea and silently break the selection being copied.
	const host = (document.activeElement?.closest("[role='dialog']") as HTMLElement | null) ?? document.body;
	host.appendChild(textarea);
	textarea.focus();
	textarea.select();
	try {
		if (!document.execCommand("copy")) {
			throw new Error("Copy command was rejected");
		}
	} finally {
		textarea.remove();
	}
};
