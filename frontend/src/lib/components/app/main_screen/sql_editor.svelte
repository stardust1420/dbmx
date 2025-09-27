<script lang="ts">
	import loader from '@monaco-editor/loader';
	import { onMount, onDestroy } from 'svelte';

	// ----- Types (strict & runtime-safe) -----
	import type * as MonacoNS from 'monaco-editor';

	let editorContainer: HTMLElement;
	let editor: MonacoNS.editor.IStandaloneCodeEditor | null = null;
	let model: MonacoNS.editor.ITextModel | null = null;

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
		suggestions = $bindable<string[]>([])
	} = $props();

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

	// External → editor value sync (skip no-ops; keep undo)
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

		monaco.languages.register({ id: 'sql' });
		monaco.editor.defineTheme('aurora-sql', AuroraSQLTheme);

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
			theme: 'aurora-sql',
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
	style="height: {height}; width: {width};"
></div>
