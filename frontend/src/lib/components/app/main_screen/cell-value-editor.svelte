<script lang="ts">
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { toast } from 'svelte-sonner';
	import { Copy, WrapText } from 'lucide-svelte';

	let {
		open = $bindable(false),
		anchor = null,
		cellId = '',
		columnName = '',
		dataType = '',
		value = '',
		readOnly = false,
		readOnlyReason = '',
		onApply
	}: {
		open?: boolean;
		anchor?: HTMLElement | null;
		/** Identifies the cell being edited, so moving to another one reseeds. */
		cellId?: string;
		columnName?: string;
		dataType?: string;
		value?: string;
		readOnly?: boolean;
		readOnlyReason?: string;
		onApply?: (value: string) => void;
	} = $props();

	/**
	 * The grid prints an absent value as NULL and an empty string as EMPTY, so
	 * neither reaches the editor as JSON. Both open an empty editor instead of a
	 * document that cannot parse.
	 */
	const SENTINELS = new Set(['NULL', 'EMPTY']);

	function reformat(text: string, indent: number): string | null {
		try {
			return JSON.stringify(JSON.parse(text), null, indent);
		} catch {
			return null;
		}
	}

	let draft = $state('');

	/**
	 * The editor opens on the formatted value because that is the readable form,
	 * and reading is most of what this is for. Applying re-encodes compactly, so
	 * formatting only ever changes what is on screen, never what is stored.
	 */
	/**
	 * Keyed on the cell as well as its value, so moving the editor to another
	 * cell always reseeds the draft. Opening on a second cell that happens to
	 * hold an identical value would otherwise leave the first cell's unapplied
	 * edits on screen, and clicking one cell while another is open collapses the
	 * close and the reopen into a single update that `open` alone cannot see.
	 */
	const seed = $derived({ cellId, value });

	$effect(() => {
		if (!open) return;
		const source = SENTINELS.has(seed.value) ? '' : seed.value;
		draft = reformat(source, 2) ?? source;
	});

	const parseError = $derived.by(() => {
		if (draft.trim() === '') return null;
		try {
			JSON.parse(draft);
			return null;
		} catch (error) {
			return error instanceof Error ? error.message : String(error);
		}
	});

	const compacted = $derived(draft.trim() === '' ? null : reformat(draft, 0));
	const canApply = $derived(!readOnly && compacted !== null);

	function apply() {
		if (compacted === null) return;
		onApply?.(compacted);
		open = false;
	}

	async function copyValue() {
		try {
			await navigator.clipboard.writeText(draft);
			toast.success('Copied', { description: `${columnName} copied to the clipboard.` });
		} catch (error) {
			toast.error('Could not copy the value.', { description: String(error) });
		}
	}
</script>

<Popover.Root bind:open>
	<Popover.Content
		customAnchor={anchor}
		align="start"
		sideOffset={6}
		collisionPadding={12}
		class="motion-reduce:animate-none flex w-[min(38rem,calc(100vw-2rem))] flex-col gap-2 p-3"
	>
		<div class="flex items-baseline justify-between gap-2">
			<span class="truncate font-mono text-sm font-medium">{columnName}</span>
			{#if dataType}
				<span class="text-muted-foreground shrink-0 font-mono text-xs">{dataType}</span>
			{/if}
		</div>

		<Textarea
			bind:value={draft}
			readonly={readOnly}
			spellcheck={false}
			autocomplete="off"
			aria-label={`${columnName} value`}
			aria-invalid={parseError !== null}
			class="h-[17rem] resize-none font-mono text-xs leading-relaxed"
		/>

		<p class="min-h-4 text-xs" class:text-destructive={parseError !== null}>
			{#if parseError}
				Not valid JSON — {parseError}
			{:else if readOnly}
				<span class="text-muted-foreground">
					{readOnlyReason || 'This value can be read and copied, but not edited.'}
				</span>
			{:else if draft.trim() === ''}
				<span class="text-muted-foreground">Empty. Type JSON to give this cell a value.</span>
			{:else}
				<span class="text-muted-foreground">Apply queues the change; ⌘S saves the grid.</span>
			{/if}
		</p>

		<div class="flex items-center justify-between gap-2">
			<div class="flex gap-2">
				<Button variant="outline" size="sm" onclick={copyValue}>
					<Copy class="size-4" />
					Copy
				</Button>
				<Button
					variant="outline"
					size="sm"
					disabled={parseError !== null || draft.trim() === ''}
					onclick={() => (draft = reformat(draft, 2) ?? draft)}
				>
					<WrapText class="size-4" />
					Format
				</Button>
			</div>
			<div class="flex gap-2">
				<Button variant="outline" size="sm" onclick={() => (open = false)}>
					{readOnly ? 'Close' : 'Cancel'}
				</Button>
				{#if !readOnly}
					<Button size="sm" disabled={!canApply} onclick={apply}>Apply</Button>
				{/if}
			</div>
		</div>
	</Popover.Content>
</Popover.Root>
