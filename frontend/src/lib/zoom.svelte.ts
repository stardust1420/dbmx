const ZOOM_STEP = 0.1;
const MIN_ZOOM = 0.5;
const MAX_ZOOM = 2.0;
const DEFAULT_ZOOM = 1.0;
const STORAGE_KEY = 'dbmx-zoom-level';

function createZoomState() {
	let level = $state(loadZoom());

	function loadZoom(): number {
		try {
			const stored = localStorage.getItem(STORAGE_KEY);
			if (stored) {
				const parsed = parseFloat(stored);
				if (!isNaN(parsed) && parsed >= MIN_ZOOM && parsed <= MAX_ZOOM) {
					return parsed;
				}
			}
		} catch {
			// localStorage may not be available
		}
		return DEFAULT_ZOOM;
	}

	function saveZoom(value: number) {
		try {
			localStorage.setItem(STORAGE_KEY, value.toString());
		} catch {
			// localStorage may not be available
		}
	}

	function applyZoom(value: number) {
		const root = document.documentElement;
		if (!root) return;
		// Use CSS zoom on the root element. This is supported in both
		// WebView2 (Windows/Chromium) and WebKit (macOS). Applying to
		// <html> avoids layout conflicts with fixed-position elements
		// that can occur when zooming the <body>.
		root.style.zoom = value === DEFAULT_ZOOM ? '' : value.toString();
	}

	return {
		get level() {
			return level;
		},
		zoomIn() {
			level = Math.min(MAX_ZOOM, Math.round((level + ZOOM_STEP) * 10) / 10);
			saveZoom(level);
			applyZoom(level);
		},
		zoomOut() {
			level = Math.max(MIN_ZOOM, Math.round((level - ZOOM_STEP) * 10) / 10);
			saveZoom(level);
			applyZoom(level);
		},
		resetZoom() {
			level = DEFAULT_ZOOM;
			saveZoom(level);
			applyZoom(level);
		},
		initialize() {
			applyZoom(level);
		}
	};
}

export const zoom = createZoomState();
