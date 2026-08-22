<script lang="ts">
	import loader from '@monaco-editor/loader';
	import { onMount, onDestroy, tick } from 'svelte';
	import { GetQueryHistory } from '$lib/wailsjs/go/app/QueryHistory.js';
	import { SaveQuery, GetSavedQueries } from '$lib/wailsjs/go/app/SavedQueries.js';
	import { format as formatSQL } from 'sql-formatter';
	import { mode } from 'mode-watcher';

	// ----- Types (strict & runtime-safe) -----
	import type * as MonacoNS from 'monaco-editor';

	let editorContainer: HTMLElement;
	let editor: MonacoNS.editor.IStandaloneCodeEditor | null = null;
	let model: MonacoNS.editor.ITextModel | null = null;
	let monacoInstance: typeof import('monaco-editor') | null = null;

	let completionProviderDisposable: MonacoNS.IDisposable | null = null;
	let cursorSub: MonacoNS.IDisposable | null = null;
	let mouseSub: MonacoNS.IDisposable | null = null;

	let isInitialized = false;

	// Props (Svelte runes)
	let {
		value = $bindable(''),
		selectedQuery = $bindable(''),
		height = '100%',
		width = '100%',
		suggestions = $bindable<string[]>([]),
		executeQuery
	} = $props();

	// Query history picker state
	interface HistoryItem {
		label: string;
		detail: string;
	}
	let showHistoryPicker = $state(false);
	let historyItems = $state<HistoryItem[]>([]);
	let filteredItems = $state<HistoryItem[]>([]);
	let historySearchTerm = $state('');
	let historySelectedIndex = $state(0);
	let historySearchInput: HTMLInputElement | null = $state(null);
	let historyCursorPosition: import('monaco-editor').IRange | null = $state(null);
	let historyListContainer: HTMLElement | null = $state(null);

	// Save query dialog state
	let showSaveDialog = $state(false);
	let saveQueryTitle = $state('');
	let saveQueryText = $state('');
	let saveQueryTitleInput: HTMLInputElement | null = $state(null);

	// Saved queries picker state
	interface SavedItem {
		id: number;
		label: string;
		title: string;
		detail: string;
	}
	let showSavedQueriesPicker = $state(false);
	let savedItems = $state<SavedItem[]>([]);
	let filteredSavedItems = $state<SavedItem[]>([]);
	let savedSearchTerm = $state('');
	let savedSelectedIndex = $state(0);
	let savedSearchInput: HTMLInputElement | null = $state(null);
	let savedCursorPosition: import('monaco-editor').IRange | null = $state(null);
	let savedListContainer: HTMLElement | null = $state(null);

	async function scrollSelectedIntoView() {
		await tick();
		if (!historyListContainer) return;
		const selected = historyListContainer.children[historySelectedIndex] as HTMLElement | undefined;
		selected?.scrollIntoView({ block: 'nearest' });
	}

	function formatTimestamp(ts: string): string {
		const date = new Date(ts);
		if (isNaN(date.getTime())) return ts;
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMin = Math.floor(diffMs / 60000);
		const diffHr = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		let relative: string;
		if (diffMin < 1) relative = 'just now';
		else if (diffMin < 60) relative = `${diffMin}m ago`;
		else if (diffHr < 24) relative = `${diffHr}h ago`;
		else if (diffDays < 7) relative = `${diffDays}d ago`;
		else relative = date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });

		const time = date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
		return `${relative} · ${time}`;
	}

	function filterHistory() {
		const term = historySearchTerm.toLowerCase();
		if (!term) {
			filteredItems = historyItems;
		} else {
			filteredItems = historyItems.filter((item) => item.label.toLowerCase().includes(term));
		}
		historySelectedIndex = 0;
	}

	function selectHistoryItem(item: HistoryItem) {
		if (editor && model) {
			const range = historyCursorPosition
				? historyCursorPosition
				: model.getFullModelRange();
			editor.pushUndoStop();
			editor.executeEdits('historyPick', [{ range, text: item.label }]);
			editor.pushUndoStop();
			value = model.getValue();
		}
		showHistoryPicker = false;
	}

	function handleHistoryKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			showHistoryPicker = false;
			editor?.focus();
		} else if (e.key === 'ArrowDown') {
			e.preventDefault();
			historySelectedIndex = Math.min(historySelectedIndex + 1, filteredItems.length - 1);
			scrollSelectedIntoView();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			historySelectedIndex = Math.max(historySelectedIndex - 1, 0);
			scrollSelectedIntoView();
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (filteredItems.length > 0) {
				selectHistoryItem(filteredItems[historySelectedIndex]);
				editor?.focus();
			}
		}
	}

	// Save query dialog functions
	async function handleSaveQuery() {
		if (!saveQueryTitle.trim()) return;
		try {
			await SaveQuery(saveQueryTitle.trim(), saveQueryText);
			showSaveDialog = false;
			saveQueryTitle = '';
			saveQueryText = '';
			editor?.focus();
		} catch (e) {
			console.error('Failed to save query:', e);
		}
	}

	function handleSaveDialogKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			showSaveDialog = false;
			editor?.focus();
		} else if (e.key === 'Enter') {
			e.preventDefault();
			handleSaveQuery();
		}
	}

	// Saved queries picker functions
	async function scrollSavedSelectedIntoView() {
		await tick();
		if (!savedListContainer) return;
		const selected = savedListContainer.children[savedSelectedIndex] as HTMLElement | undefined;
		selected?.scrollIntoView({ block: 'nearest' });
	}

	function filterSavedQueries() {
		const term = savedSearchTerm.toLowerCase();
		if (!term) {
			filteredSavedItems = savedItems;
		} else {
			filteredSavedItems = savedItems.filter(
				(item) => item.title.toLowerCase().includes(term) || item.label.toLowerCase().includes(term)
			);
		}
		savedSelectedIndex = 0;
	}

	function selectSavedItem(item: SavedItem) {
		if (editor && model) {
			const range = savedCursorPosition
				? savedCursorPosition
				: model.getFullModelRange();
			editor.pushUndoStop();
			editor.executeEdits('savedQueryPick', [{ range, text: item.label }]);
			editor.pushUndoStop();
			value = model.getValue();
		}
		showSavedQueriesPicker = false;
	}

	function handleSavedQueriesKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') {
			showSavedQueriesPicker = false;
			editor?.focus();
		} else if (e.key === 'ArrowDown') {
			e.preventDefault();
			savedSelectedIndex = Math.min(savedSelectedIndex + 1, filteredSavedItems.length - 1);
			scrollSavedSelectedIntoView();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			savedSelectedIndex = Math.max(savedSelectedIndex - 1, 0);
			scrollSavedSelectedIntoView();
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (filteredSavedItems.length > 0) {
				selectSavedItem(filteredSavedItems[savedSelectedIndex]);
				editor?.focus();
			}
		}
	}

	function refreshSuggestWidget() {
		if (!editor) return;
		const sc = (editor as any).getContribution?.('editor.contrib.suggestController');
		if (sc?.cancelSuggestWidget) sc.cancelSuggestWidget();
		if (sc?.triggerSuggest) sc.triggerSuggest();
		else editor.trigger('keyboard', 'editor.action.triggerSuggest', {});
	}

	// Local caches for fast, case-insensitive matching
	let localSuggestions: string[] = [];
	let localSuggestionsLC: string[] = [];
	$effect(() => {
		localSuggestions = Array.isArray(suggestions) ? suggestions.slice() : [];
		localSuggestionsLC = localSuggestions.map((s) => s.toLowerCase());

		if (editor && isInitialized) schedule(refreshSuggestWidget);
	});

	const AuroraSQLTheme: MonacoNS.editor.IStandaloneThemeData = {
		base: 'vs-dark',
		inherit: true,
		rules: [
			// base / comments
			{ token: '', foreground: 'E6E9EF', background: '000000' },
			{ token: 'comment', foreground: '7D8696', fontStyle: 'italic' },
			{ token: 'comment.sql', foreground: '7D8696', fontStyle: 'italic' },

			// keywords (SELECT, WHERE, JOIN, LIMIT…)
			{ token: 'keyword', foreground: '8CAAEE' }, // soft azure
			{ token: 'keyword.sql', foreground: '8CAAEE' },

			// types (INT, VARCHAR…), NULL/TRUE/FALSE
			{ token: 'type', foreground: '8BD5CA' }, // teal
			{ token: 'type.sql', foreground: '8BD5CA' },
			{ token: 'predefined', foreground: 'F2CDCD' }, // NULL/TRUE/FALSE (rose)
			{ token: 'predefined.sql', foreground: 'F2CDCD' },

			// strings & numbers
			{ token: 'string', foreground: 'A6E3A1' }, // mint
			{ token: 'string.sql', foreground: 'A6E3A1' },
			{ token: 'string.escape', foreground: 'B5E8E0' },
			{ token: 'number', foreground: 'F5A97F' }, // coral/peach
			{ token: 'number.sql', foreground: 'F5A97F' },

			// functions (COUNT, SUM, MAX…)
			{ token: 'entity.name.function', foreground: 'EED49F' }, // amber
			{ token: 'support.function', foreground: 'EED49F' },

			// identifiers (tables/columns), quoted names
			{ token: 'identifier', foreground: 'DCE1EA' },
			{ token: 'identifier.sql', foreground: 'DCE1EA' },
			{ token: 'identifier.quote', foreground: 'AAB4CF' },

			// operators / punctuation
			{ token: 'operator', foreground: 'BAC2DE' },
			{ token: 'operator.sql', foreground: 'BAC2DE' },
			{ token: 'delimiter', foreground: '6C7086' },
			{ token: 'delimiter.sql', foreground: '6C7086' },
			{ token: 'delimiter.bracket', foreground: 'B4BEFE' },

			// errors (muted)
			{ token: 'invalid', foreground: 'F38BA8' },
			{ token: 'invalid.deprecated', foreground: 'F9E2AF' }
		],
		colors: {
			// canvas & text
			'editor.background': '#000000',
			'editor.foreground': '#E6E9EF',

			// gutter / guides
			'editorLineNumber.foreground': '#2A3340',
			'editorLineNumber.activeForeground': '#9DB2CE',
			'editorIndentGuide.background': '#121212',
			'editorIndentGuide.activeBackground': '#1E1E1E',
			'editorWhitespace.foreground': '#202020',
			editorLineHighlightBackground: '#0A0C10',

			// caret / selection / word highlights
			'editorCursor.foreground': '#E6E9EF',
			'editor.selectionBackground': '#1E2A3A',
			'editor.inactiveSelectionBackground': '#10151C',
			'editor.wordHighlightBackground': '#0B122066',
			'editor.wordHighlightStrongBackground': '#0B122099',

			// matches / search
			'editor.findMatchBackground': '#22D3EE66',
			'editor.findMatchHighlightBackground': '#6EE7B780',
			'editor.findRangeHighlightBackground': '#37415166',

			// brackets
			'editorBracketMatch.background': '#101318',
			'editorBracketMatch.border': '#2A3646',

			// autocomplete
			'editorSuggestWidget.background': '#0B0D11',
			'editorSuggestWidget.border': '#1F2633',
			'editorSuggestWidget.foreground': '#D7DBE2',
			'editorSuggestWidget.selectedBackground': '#151B24',
			'editorSuggestWidget.highlightForeground': '#8CAAEE',

			// hover / peek
			'editorHoverWidget.background': '#0B0D11',
			'editorHoverWidget.border': '#1F2633',

			// scrollbar
			'scrollbarSlider.background': '#2A2A2AAA',
			'scrollbarSlider.hoverBackground': '#3A3A3AAA',
			'scrollbarSlider.activeBackground': '#4A4A4AAA',

			// markers
			'editorError.foreground': '#F38BA8',
			'editorWarning.foreground': '#EED49F',
			'editorInfo.foreground': '#8CAAEE'
		}
	};

	const AuroraSQLLightTheme: MonacoNS.editor.IStandaloneThemeData = {
		base: 'vs',
		inherit: true,
		rules: [
			{ token: '', foreground: '1E293B', background: 'FFFFFF' },
			{ token: 'comment', foreground: '94A3B8', fontStyle: 'italic' },
			{ token: 'comment.sql', foreground: '94A3B8', fontStyle: 'italic' },
			{ token: 'keyword', foreground: '2563EB' },
			{ token: 'keyword.sql', foreground: '2563EB' },
			{ token: 'type', foreground: '0D9488' },
			{ token: 'type.sql', foreground: '0D9488' },
			{ token: 'predefined', foreground: 'BE185D' },
			{ token: 'predefined.sql', foreground: 'BE185D' },
			{ token: 'string', foreground: '16A34A' },
			{ token: 'string.sql', foreground: '16A34A' },
			{ token: 'string.escape', foreground: '0D9488' },
			{ token: 'number', foreground: 'EA580C' },
			{ token: 'number.sql', foreground: 'EA580C' },
			{ token: 'entity.name.function', foreground: 'D97706' },
			{ token: 'support.function', foreground: 'D97706' },
			{ token: 'identifier', foreground: '334155' },
			{ token: 'identifier.sql', foreground: '334155' },
			{ token: 'identifier.quote', foreground: '64748B' },
			{ token: 'operator', foreground: '475569' },
			{ token: 'operator.sql', foreground: '475569' },
			{ token: 'delimiter', foreground: '94A3B8' },
			{ token: 'delimiter.sql', foreground: '94A3B8' },
			{ token: 'delimiter.bracket', foreground: '7C3AED' },
			{ token: 'invalid', foreground: 'DC2626' },
			{ token: 'invalid.deprecated', foreground: 'D97706' }
		],
		colors: {
			'editor.background': '#FFFFFF',
			'editor.foreground': '#1E293B',
			'editorLineNumber.foreground': '#CBD5E1',
			'editorLineNumber.activeForeground': '#475569',
			'editorIndentGuide.background': '#F1F5F9',
			'editorIndentGuide.activeBackground': '#E2E8F0',
			'editorWhitespace.foreground': '#F1F5F9',
			editorLineHighlightBackground: '#F8FAFC',
			'editorCursor.foreground': '#1E293B',
			'editor.selectionBackground': '#DBEAFE',
			'editor.inactiveSelectionBackground': '#EFF6FF',
			'editor.wordHighlightBackground': '#DBEAFE66',
			'editor.wordHighlightStrongBackground': '#DBEAFE99',
			'editor.findMatchBackground': '#FDE68A',
			'editor.findMatchHighlightBackground': '#BBF7D080',
			'editor.findRangeHighlightBackground': '#F1F5F966',
			'editorBracketMatch.background': '#EDE9FE',
			'editorBracketMatch.border': '#C4B5FD',
			'editorSuggestWidget.background': '#FFFFFF',
			'editorSuggestWidget.border': '#E2E8F0',
			'editorSuggestWidget.foreground': '#334155',
			'editorSuggestWidget.selectedBackground': '#F1F5F9',
			'editorSuggestWidget.highlightForeground': '#2563EB',
			'editorHoverWidget.background': '#FFFFFF',
			'editorHoverWidget.border': '#E2E8F0',
			'scrollbarSlider.background': '#CBD5E1AA',
			'scrollbarSlider.hoverBackground': '#94A3B8AA',
			'scrollbarSlider.activeBackground': '#64748BAA',
			'editorError.foreground': '#DC2626',
			'editorWarning.foreground': '#D97706',
			'editorInfo.foreground': '#2563EB'
		}
	};
	$effect(() => {
		const incoming = value;
		if (!editor || !isInitialized || !model) return;

		const current = model.getValue();
		if (current === incoming) return;

		const fullRange = model.getFullModelRange();
		editor.pushUndoStop();
		editor.executeEdits('propSync', [{ range: fullRange, text: incoming }]);
		editor.pushUndoStop();
	});

	// Switch Monaco theme when app mode changes
	$effect(() => {
		const currentMode = mode.current;
		if (!monacoInstance || !editor) return;
		const themeName = currentMode === 'light' ? 'aurora-sql-light' : 'aurora-sql';
		monacoInstance.editor.setTheme(themeName);
	});

	// Helper: compute blank-line-delimited block range
	function calcBlockRange(
		monaco: typeof import('monaco-editor'),
		m: MonacoNS.editor.ITextModel,
		lineNumber: number
	): MonacoNS.Range {
		let start = lineNumber;
		let end = lineNumber;
		const last = m.getLineCount();

		while (start > 1) {
			if (m.getLineContent(start - 1).trim() === '') break;
			start--;
		}
		while (end < last) {
			if (m.getLineContent(end + 1).trim() === '') break;
			end++;
		}

		return new monaco.Range(start, 1, end, m.getLineLength(end) + 1);
	}

	// rAF throttle
	let rafToken: number | null = null;
	function schedule(fn: () => void) {
		if (rafToken != null) cancelAnimationFrame(rafToken);
		rafToken = requestAnimationFrame(() => {
			rafToken = null;
			fn();
		});
	}

	onMount(async () => {
		const monaco = await loader.init();
		monacoInstance = monaco;

		monaco.languages.register({ id: 'sql' });
		monaco.editor.defineTheme('aurora-sql', AuroraSQLTheme);
		monaco.editor.defineTheme('aurora-sql-light', AuroraSQLLightTheme);

		// Completion provider (case-insensitive startsWith)
		completionProviderDisposable = monaco.languages.registerCompletionItemProvider('sql', {
			triggerCharacters: [' ', '.', '(', '_'],
			provideCompletionItems: (m, position) => {
				const word = m.getWordUntilPosition(position);
				const prefix = (word.word || '').toLowerCase();

				const makeItem = (keyword: string, sortIdx: number) => ({
					label: keyword,
					kind: monaco.languages.CompletionItemKind.Keyword,
					insertText: keyword,
					sortText: String(sortIdx).padStart(4, '0'),
					range: {
						startLineNumber: position.lineNumber,
						startColumn: word.startColumn,
						endLineNumber: position.lineNumber,
						endColumn: word.endColumn
					}
				});

				if (!prefix) {
					return { suggestions: localSuggestions.map((kw, i) => makeItem(kw, i)) };
				}

				const out: ReturnType<typeof makeItem>[] = [];
				for (let i = 0; i < localSuggestionsLC.length; i++) {
					if (localSuggestionsLC[i].startsWith(prefix)) {
						out.push(makeItem(localSuggestions[i], out.length));
					}
				}
				return { suggestions: out };
			}
		});

		editor = monaco.editor.create(editorContainer, {
			value,
			language: 'sql', // ✅ matches provider
			theme: mode.current === 'light' ? 'aurora-sql-light' : 'aurora-sql',
			automaticLayout: true,
			minimap: { enabled: false },
			fontSize: 14,
			wordWrap: 'off',
			scrollBeyondLastLine: false,
			renderWhitespace: 'none',
			renderLineHighlight: 'line',
			glyphMargin: true,
			quickSuggestions: { other: true, comments: false, strings: true },
			suggestOnTriggerCharacters: true,
			fixedOverflowWidgets: true,
			smoothScrolling: true, // smooth viewport scrolling
			cursorSmoothCaretAnimation: 'on' as any, // animate left/right caret moves
			cursorBlinking: 'smooth', // nicer blink animation (optional)
			stickyTabStops: true, // nicer left/right in leading spaces
			scrollBeyondLastColumn: 3 // avoids hugging the right edge (optional)
		});

		model = editor.getModel();
		if (!model) {
			// extremely unlikely, but guard for TS + runtime
			isInitialized = true;
			return;
		}

		// Reusable decorations
		const deco = editor.createDecorationsCollection([]);

		function updateQuerySelection() {
			if (!editor || !model) return;

			const pos = editor.getPosition();
			if (!pos) return;

			const range = calcBlockRange(monaco, model!, pos.lineNumber);

			const existing = deco.getRanges();
			if (existing.length === 1) {
				const r = existing[0];
				if (
					r.startLineNumber === range.startLineNumber &&
					r.endLineNumber === range.endLineNumber
				) {
					selectedQuery = model!.getValueInRange(range).trim();
					return;
				}
			}

			deco.set([
				{
					range,
					options: {
						isWholeLine: true,
						className: 'bg-green-100 bg-opacity-5',
						glyphMarginClassName: 'bg-green-500 bg-opacity-20'
					}
				}
			]);

			selectedQuery = model!.getValueInRange(range).trim();
		}

		// Mouse + cursor events (throttled)
		mouseSub = editor.onMouseDown(() => schedule(updateQuerySelection));
		cursorSub = editor.onDidChangeCursorPosition(() => schedule(updateQuerySelection));

		// Keep external value in sync
		editor.onDidChangeModelContent(() => {
			value = editor!.getValue();
		});

		// Nice UX: open suggestions on focus
		editor.onDidFocusEditorText(() => {
			editor!.trigger('focus', 'editor.action.triggerSuggest', {});
		});

		editor.addAction({
			id: 'run-query',
			label: 'Run Query',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 1,
			keybindings: [monaco.KeyMod.Alt | monaco.KeyCode.Enter],
			run: (ed) => {
				try {
					console.log('Executing query from editor action');
					executeQuery();
				} catch (e) {
					console.error('Failed to run SQL query:', e);
				}
			}
		});

		editor.addAction({
			id: 'explain-query',
			label: 'Explain Query',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 2,
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyE],
			run: (ed) => {
				try {
					console.log('Executing query from editor action');
					executeQuery(true);
				} catch (e) {
					console.error('Failed to run SQL query:', e);
				}
			}
		});

		// Register "Pretty Format Query" in the right-click context menu
		editor.addAction({
			id: 'pretty-format-query',
			label: 'Pretty Format Query',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 3,
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF],
			run: (ed) => {
				if (!model) return;
				const pos = ed.getPosition();
				if (!pos) return;

				const range = calcBlockRange(monaco, model, pos.lineNumber);
				const queryText = model.getValueInRange(range);

				try {
					const formatted = formatSQL(queryText, {
						language: 'sql',
						tabWidth: 4,
						keywordCase: 'upper'
					});

					ed.pushUndoStop();
					ed.executeEdits('prettyFormat', [{ range, text: formatted }]);
					ed.pushUndoStop();
					value = model!.getValue();
				} catch (e) {
					console.error('Failed to format SQL query:', e);
				}
			}
		});

		// Register Cmd+H / Ctrl+H to show query history quick pick
		editor.addAction({
			id: 'show-query-history',
			label: 'Show Query History',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 4,
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyH],
			run: async () => {
				try {
					const history = await GetQueryHistory();
					if (!history || history.length === 0) {
						return;
					}
					historyItems = history.map((h) => ({
						label: h.query,
						detail: formatTimestamp(h.executedAt)
					}));
					filteredItems = historyItems;
					historySearchTerm = '';
					historyCursorPosition = editor?.getSelection() ?? null;
					showHistoryPicker = true;
					// Focus the search input after DOM update
					requestAnimationFrame(() => {
						historySearchInput?.focus();
					});
				} catch (e) {
					console.error('Failed to load query history:', e);
				}
			}
		});

		// Register "Save Query" in right-click context menu
		editor.addAction({
			id: 'save-query',
			label: 'Save Query',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 5,
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS],
			run: () => {
				if (!model) return;
				const queryText = selectedQuery || '';
				if (!queryText.trim()) return;
				saveQueryText = queryText;
				saveQueryTitle = '';
				showSaveDialog = true;
				requestAnimationFrame(() => {
					saveQueryTitleInput?.focus();
				});
			}
		});

		// Register "Show Saved Queries" in right-click context menu
		editor.addAction({
			id: 'show-saved-queries',
			label: 'Show Saved Queries',
			contextMenuGroupId: 'modification',
			contextMenuOrder: 6,
			keybindings: [monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyS],
			run: async () => {
				try {
					const queries = await GetSavedQueries();
					if (!queries || queries.length === 0) {
						return;
					}
					savedItems = queries.map((q) => ({
						id: q.id,
						label: q.query,
						title: q.title,
						detail: `${q.title} · ${formatTimestamp(q.savedAt)}`
					}));
					filteredSavedItems = savedItems;
					savedSearchTerm = '';
					savedCursorPosition = editor?.getSelection() ?? null;
					showSavedQueriesPicker = true;
					requestAnimationFrame(() => {
						savedSearchInput?.focus();
					});
				} catch (e) {
					console.error('Failed to load saved queries:', e);
				}
			}
		});

		isInitialized = true;
	});

	onDestroy(() => {
		cursorSub?.dispose();
		mouseSub?.dispose();
		completionProviderDisposable?.dispose();
		editor?.dispose();
		editor = null;
		model = null;
	});

	// Optional: selecting text outside Monaco (kept from your version)
	function handleSelection() {
		const selection = window.getSelection();
		if (selection && selection.toString().trim() !== '') {
			selectedQuery = selection.toString().trim();
		}
	}
</script>

<div
	onselect={handleSelection}
	bind:this={editorContainer}
	class="sql-editor"
	style="height: {height}; width: {width}; position: relative;"
></div>

{#if showHistoryPicker}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-start justify-center pt-[10%]"
		onmousedown={(e) => { if (e.target === e.currentTarget) { showHistoryPicker = false; editor?.focus(); } }}
	>
		<div class="w-[600px] max-h-[400px] bg-[#1e1e1e] border border-[#3c3c3c] rounded-md shadow-2xl flex flex-col overflow-hidden">
			<input
				bind:this={historySearchInput}
				bind:value={historySearchTerm}
				oninput={filterHistory}
				onkeydown={handleHistoryKeydown}
				type="text"
				placeholder="Search query history..."
				class="w-full px-3 py-2 bg-[#252526] text-[#cccccc] text-sm border-b border-[#3c3c3c] outline-none placeholder-[#6c6c6c]"
			/>
			<div class="overflow-y-auto flex-1" bind:this={historyListContainer}>
				{#each filteredItems as item, i}
					<button
						class="w-full text-left px-3 py-2 text-sm cursor-pointer hover:bg-[#2a2d2e] {i === historySelectedIndex ? 'bg-[#04395e]' : ''}"
						onmousedown={() => { selectHistoryItem(item); editor?.focus(); }}
					>
						<div class="text-[#cccccc] truncate font-mono text-xs">{item.label}</div>
						<div class="text-[#6c6c6c] text-xs mt-0.5">{item.detail}</div>
					</button>
				{/each}
				{#if filteredItems.length === 0}
					<div class="px-3 py-4 text-[#6c6c6c] text-sm text-center">No matching queries</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

{#if showSaveDialog}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-start justify-center pt-[10%]"
		onmousedown={(e) => { if (e.target === e.currentTarget) { showSaveDialog = false; editor?.focus(); } }}
	>
		<div class="w-[400px] bg-[#1e1e1e] border border-[#3c3c3c] rounded-md shadow-2xl flex flex-col overflow-hidden">
			<div class="px-3 py-2 border-b border-[#3c3c3c] text-[#cccccc] text-sm font-medium">Save Query</div>
			<div class="px-3 py-2">
				<label class="text-[#9c9c9c] text-xs block mb-1">Title</label>
				<input
					bind:this={saveQueryTitleInput}
					bind:value={saveQueryTitle}
					onkeydown={handleSaveDialogKeydown}
					type="text"
					placeholder="Enter a title for this query..."
					class="w-full px-3 py-2 bg-[#252526] text-[#cccccc] text-sm border border-[#3c3c3c] rounded outline-none placeholder-[#6c6c6c] focus:border-[#007acc]"
				/>
				<div class="mt-2 text-[#6c6c6c] text-xs font-mono truncate">{saveQueryText}</div>
			</div>
			<div class="px-3 py-2 flex justify-end gap-2 border-t border-[#3c3c3c]">
				<button
					class="px-3 py-1 text-sm text-[#cccccc] bg-[#3c3c3c] rounded hover:bg-[#4c4c4c]"
					onmousedown={() => { showSaveDialog = false; editor?.focus(); }}
				>Cancel</button>
				<button
					class="px-3 py-1 text-sm text-primary-foreground bg-primary rounded hover:bg-primary/90 disabled:opacity-50"
					disabled={!saveQueryTitle.trim()}
					onmousedown={handleSaveQuery}
				>Save</button>
			</div>
		</div>
	</div>
{/if}

{#if showSavedQueriesPicker}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-50 flex items-start justify-center pt-[10%]"
		onmousedown={(e) => { if (e.target === e.currentTarget) { showSavedQueriesPicker = false; editor?.focus(); } }}
	>
		<div class="w-[600px] max-h-[400px] bg-[#1e1e1e] border border-[#3c3c3c] rounded-md shadow-2xl flex flex-col overflow-hidden">
			<input
				bind:this={savedSearchInput}
				bind:value={savedSearchTerm}
				oninput={filterSavedQueries}
				onkeydown={handleSavedQueriesKeydown}
				type="text"
				placeholder="Search saved queries..."
				class="w-full px-3 py-2 bg-[#252526] text-[#cccccc] text-sm border-b border-[#3c3c3c] outline-none placeholder-[#6c6c6c]"
			/>
			<div class="overflow-y-auto flex-1" bind:this={savedListContainer}>
				{#each filteredSavedItems as item, i}
					<button
						class="w-full text-left px-3 py-2 text-sm cursor-pointer hover:bg-[#2a2d2e] {i === savedSelectedIndex ? 'bg-[#04395e]' : ''}"
						onmousedown={() => { selectSavedItem(item); editor?.focus(); }}
					>
						<div class="text-[#e0e0e0] text-xs font-medium">{item.title}</div>
						<div class="text-[#cccccc] truncate font-mono text-xs mt-0.5">{item.label}</div>
						<div class="text-[#6c6c6c] text-xs mt-0.5">{item.detail}</div>
					</button>
				{/each}
				{#if filteredSavedItems.length === 0}
					<div class="px-3 py-4 text-[#6c6c6c] text-sm text-center">No saved queries</div>
				{/if}
			</div>
		</div>
	</div>
{/if}