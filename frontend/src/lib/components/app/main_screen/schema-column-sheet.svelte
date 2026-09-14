<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { toast } from 'svelte-sonner';
	import { cn } from '$lib/utils.js';
	import type { model } from '$lib/wailsjs/go/models';
	import {
		AddColumn,
		GetColumnDefinition,
		GetSchemaEditorOptions,
		UpdateColumn
	} from '$lib/wailsjs/go/app/Connections';
	import SchemaCombobox, { type ComboboxOption } from './schema-combobox.svelte';

	let {
		tabID,
		tableName,
		open = $bindable(false),
		/** The column being edited; empty opens the sheet as "add column". */
		columnName = '',
		onSaved
	}: {
		tabID: number;
		tableName: string;
		open?: boolean;
		columnName?: string;
		onSaved?: () => void;
	} = $props();

	const isEdit = $derived(columnName !== '');

	function blankColumn(): model.ColumnDefinition {
		return {
			name: '',
			dataType: '',
			isNullable: true,
			defaultValue: '',
			identity: '',
			generatedExpression: '',
			collation: '',
			comment: '',
			usingExpression: '',
			isPrimaryKey: false
		};
	}

	let form = $state<model.ColumnDefinition>(blankColumn());
	let original = $state<model.ColumnDefinition | null>(null);

	// The type is edited as three pieces, because a modifier cannot be typed into the
	// picker: its trigger is a button, so once the list closes there is nowhere to put
	// the length. form.dataType is assembled from them.
	let baseType = $state('');
	let typeModifier = $state('');
	let arraySuffix = $state('');
	let types = $state<model.TypeOption[]>([]);
	let collationOptions = $state<string[]>([]);
	let loading = $state(false);
	let saving = $state(false);
	let showErrors = $state(false);

	$effect(() => {
		if (open) {
			load();
		}
	});

	async function load() {
		loading = true;
		saving = false;
		showErrors = false;
		form = blankColumn();
		original = null;
		baseType = '';
		typeModifier = '';
		arraySuffix = '';

		try {
			const [options, existing] = await Promise.all([
				GetSchemaEditorOptions(tabID),
				columnName ? GetColumnDefinition(tabID, tableName, columnName) : Promise.resolve(null)
			]);

			types = options.types ?? [];
			collationOptions = options.collations ?? [];

			if (existing) {
				form = { ...blankColumn(), ...existing };
				original = { ...blankColumn(), ...existing };
				const parts = splitType(existing.dataType);
				baseType = parts.base;
				typeModifier = parts.modifier;
				arraySuffix = parts.suffix;
			}
		} catch (error) {
			toast.error('Failed to open the column editor', { description: String(error) });
			open = false;
		} finally {
			loading = false;
		}
	}

	// format_type does not always put the modifier last: a varchar reads
	// "character varying(255)", but a timestamp reads "timestamp(3) without time zone",
	// with the modifier wedged after the first word. Both are split back into a base
	// name the picker holds and a modifier of its own, and reassembled the same way.
	function splitType(dataType: string): { base: string; modifier: string; suffix: string } {
		let text = dataType.trim();
		let suffix = '';
		while (text.endsWith('[]')) {
			suffix = `[]${suffix}`;
			text = text.slice(0, -2).trim();
		}

		const parts = text.match(/^([^()]*?)\s*\(([^()]*)\)\s*([^()]*?)$/);
		if (!parts) return { base: text, modifier: '', suffix };

		return {
			base: [parts[1], parts[3]].filter(Boolean).join(' ').trim(),
			modifier: parts[2].trim(),
			suffix
		};
	}

	function composeType(base: string, modifier: string, suffix: string): string {
		const trimmedBase = base.trim();
		if (trimmedBase === '') return '';

		let text = trimmedBase;
		const trimmedModifier = modifier.trim();
		if (trimmedModifier !== '') {
			// "timestamp(3) without time zone" -- postgres rejects the modifier on the end.
			const datetime = trimmedBase.match(/^(timestamp|time)( with(?:out)? time zone)$/i);
			text = datetime
				? `${datetime[1]}(${trimmedModifier})${datetime[2]}`
				: `${trimmedBase}(${trimmedModifier})`;
		}

		return text + suffix;
	}

	// A value typed into the picker rather than chosen from it may already carry its
	// own modifier and array suffix, so it is split the same way a stored type is.
	function setBaseType(next: string) {
		const parts = splitType(next);
		baseType = parts.base;
		if (parts.modifier !== '') typeModifier = parts.modifier;
		if (parts.suffix !== '') arraySuffix = parts.suffix;
	}

	const typeOptions = $derived.by(() => {
		const options: ComboboxOption[] = types.map((type) => ({
			value: type.name,
			group: type.isCommon ? 'Common' : type.kind === 'base' ? 'All types' : 'This database',
			detail:
				type.kind === 'enum' && type.enumValues?.length
					? type.enumValues.join(', ')
					: type.kind === 'base'
						? undefined
						: type.kind
		}));
		return options.sort((left, right) => groupRank(left.group) - groupRank(right.group));
	});

	const selectedType = $derived(types.find((type) => type.name === baseType));

	// A type the picker does not know is one the user typed, and it may well take a
	// modifier, so the field stays available rather than disappearing on them.
	const acceptsModifier = $derived(
		baseType !== '' && (selectedType === undefined || selectedType.acceptsModifier)
	);

	const modifierHint = $derived.by(() => {
		const lower = baseType.toLowerCase();
		if (lower === 'numeric' || lower === 'decimal') {
			return { label: 'Precision and scale', placeholder: '10,2', description: 'Total digits, then digits after the point.' };
		}
		if (/^(time|timestamp|interval)/.test(lower)) {
			return { label: 'Fractional seconds', placeholder: '6', description: 'Digits kept after the decimal point, 0 to 6.' };
		}
		return { label: 'Length', placeholder: '255', description: 'Maximum number of characters. Leave empty for no limit.' };
	});

	function groupRank(group: string | undefined): number {
		if (group === 'Common') return 0;
		if (group === 'This database') return 1;
		return 2;
	}

	// Postgres allows exactly one of these three to define a column's value.
	const valueSource = $derived(
		form.generatedExpression.trim() !== ''
			? 'generated'
			: form.identity !== ''
				? 'identity'
				: form.defaultValue.trim() !== ''
					? 'default'
					: 'none'
	);

	const dataType = $derived(composeType(baseType, typeModifier, arraySuffix));
	const typeChanged = $derived(original !== null && dataType !== original.dataType);

	const nameError = $derived(form.name.trim() === '' ? 'A column name is required.' : '');
	const typeError = $derived(dataType === '' ? 'A data type is required.' : '');
	const generatedError = $derived(
		isEdit &&
			original !== null &&
			form.generatedExpression.trim() !== '' &&
			form.generatedExpression.trim() !== original.generatedExpression
			? "A generated column's expression cannot be changed in place. Drop the column and add it again."
			: ''
	);
	const hasErrors = $derived(nameError !== '' || typeError !== '' || generatedError !== '');

	function save() {
		showErrors = true;
		if (hasErrors) return;

		const payload: model.ColumnDefinition = {
			...form,
			name: form.name.trim(),
			dataType,
			defaultValue: form.defaultValue.trim(),
			generatedExpression: form.generatedExpression.trim(),
			collation: form.collation.trim(),
			comment: form.comment.trim(),
			usingExpression: typeChanged ? form.usingExpression.trim() : ''
		};

		saving = true;
		const request = isEdit
			? UpdateColumn(tabID, tableName, columnName, payload)
			: AddColumn(tabID, tableName, payload);

		request
			.then(() => {
				toast.success(isEdit ? 'Column updated' : 'Column added', {
					description: `${payload.name} on ${tableName}.`
				});
				open = false;
				onSaved?.();
			})
			.catch((error) => {
				toast.error(isEdit ? 'Failed to update the column' : 'Failed to add the column', {
					description: String(error)
				});
				saving = false;
			});
	}
</script>

<Sheet.Root bind:open>
	<Sheet.Content side="right" class="flex w-full flex-col gap-0 p-0 sm:max-w-lg">
		<Sheet.Header class="border-b px-6 py-4 pr-12 text-left">
			<Sheet.Title>
				{isEdit ? `Edit ${columnName}` : `Add a column to ${tableName}`}
			</Sheet.Title>
			<Sheet.Description>
				{isEdit
					? 'Only what you change is altered, and the whole edit is applied in one transaction.'
					: 'Constraints and indexes for this column are managed under Rules and Indexes.'}
			</Sheet.Description>
		</Sheet.Header>

		<div class="flex-1 overflow-y-auto px-6 py-5">
			{#if loading}
				<div class="flex flex-col gap-6">
					{#each [1, 2, 3, 4] as placeholder (placeholder)}
						<div class="flex flex-col gap-2">
							<Skeleton class="h-4 w-32" />
							<Skeleton class="h-9 w-full" />
						</div>
					{/each}
				</div>
			{:else}
				<Field.FieldGroup>
					<Field.Field data-invalid={showErrors && nameError !== ''}>
						<Field.FieldLabel for="column-name" class="flex items-center gap-2">
							Name
							{#if original?.isPrimaryKey}
								<Badge variant="secondary" class="px-1.5 py-0 text-[10px]">PK</Badge>
							{/if}
						</Field.FieldLabel>
						<Input
							id="column-name"
							bind:value={form.name}
							placeholder="created_at"
							aria-invalid={showErrors && nameError !== ''}
						/>
						{#if showErrors && nameError}
							<Field.FieldError>{nameError}</Field.FieldError>
						{/if}
					</Field.Field>

					<Field.Field data-invalid={showErrors && typeError !== ''}>
						<Field.FieldLabel for="column-type">Data type</Field.FieldLabel>
						<SchemaCombobox
							id="column-type"
							bind:value={() => baseType, setBaseType}
							options={typeOptions}
							placeholder="Select a type"
							searchPlaceholder="Search types, or type one"
							emptyText="No type matches. Press Enter to use what you typed."
							invalid={showErrors && typeError !== ''}
						/>
						{#if showErrors && typeError}
							<Field.FieldError>{typeError}</Field.FieldError>
						{/if}
					</Field.Field>

					{#if acceptsModifier}
						<Field.Field>
							<Field.FieldLabel for="column-type-modifier">
								{modifierHint.label}
							</Field.FieldLabel>
							<Input
								id="column-type-modifier"
								inputmode="numeric"
								bind:value={typeModifier}
								placeholder={modifierHint.placeholder}
							/>
							<Field.FieldDescription>{modifierHint.description}</Field.FieldDescription>
						</Field.Field>
					{/if}

					<Field.Field orientation="horizontal">
						<Field.FieldContent>
							<Field.FieldLabel for="column-type-array">Array</Field.FieldLabel>
							<Field.FieldDescription>
								Hold many values of this type in each row.
							</Field.FieldDescription>
						</Field.FieldContent>
						<Switch
							id="column-type-array"
							checked={arraySuffix !== ''}
							onCheckedChange={(checked: boolean) => (arraySuffix = checked ? '[]' : '')}
						/>
					</Field.Field>

					{#if dataType !== ''}
						<Field.Field>
							<Field.FieldLabel>Column will be declared as</Field.FieldLabel>
							<pre class="bg-muted/50 text-muted-foreground overflow-x-auto rounded-md border px-3 py-2 font-mono text-xs">{form.name.trim() ||
									'column'} {dataType}</pre>
						</Field.Field>
					{/if}

					{#if typeChanged}
						<Field.Field>
							<Field.FieldLabel for="column-using">Conversion expression</Field.FieldLabel>
							<Input
								id="column-using"
								class="font-mono text-xs"
								bind:value={form.usingExpression}
								placeholder={`${form.name || 'column'}::${dataType || 'type'}`}
							/>
							<Field.FieldDescription>
								Only needed when postgres has no automatic cast from
								<code>{original?.dataType}</code> to the new type.
							</Field.FieldDescription>
						</Field.Field>
					{/if}

					<Field.Field orientation="horizontal">
						<Field.FieldContent>
							<Field.FieldLabel for="column-nullable">Allow NULL</Field.FieldLabel>
							<Field.FieldDescription>
								Turning this off on a table with rows needs every existing row to already
								have a value.
							</Field.FieldDescription>
						</Field.FieldContent>
						<Switch id="column-nullable" bind:checked={form.isNullable} />
					</Field.Field>

					<Field.Field>
						<Field.FieldLabel for="column-default">Default</Field.FieldLabel>
						<Input
							id="column-default"
							class="font-mono text-xs"
							bind:value={form.defaultValue}
							disabled={valueSource === 'identity' || valueSource === 'generated'}
							placeholder="now()"
						/>
						<Field.FieldDescription>
							A SQL expression. Quote string literals: <code>'draft'</code>.
						</Field.FieldDescription>
					</Field.Field>

					<Field.Field>
						<Field.FieldLabel for="column-identity">Identity</Field.FieldLabel>
						<Select.Root
							type="single"
							value={form.identity}
							onValueChange={(next) => (form.identity = next)}
						>
							<Select.Trigger
								id="column-identity"
								class="w-full"
								disabled={valueSource === 'default' || valueSource === 'generated'}
							>
								{form.identity === '' ? 'None' : `GENERATED ${form.identity} AS IDENTITY`}
							</Select.Trigger>
							<Select.Content>
								<Select.Item value="">None</Select.Item>
								<Select.Item value="BY DEFAULT">GENERATED BY DEFAULT AS IDENTITY</Select.Item>
								<Select.Item value="ALWAYS">GENERATED ALWAYS AS IDENTITY</Select.Item>
							</Select.Content>
						</Select.Root>
						<Field.FieldDescription>
							Postgres supplies the value from a sequence it owns.
						</Field.FieldDescription>
					</Field.Field>

					<Field.Field data-invalid={showErrors && generatedError !== ''}>
						<Field.FieldLabel for="column-generated">Generated from</Field.FieldLabel>
						<Input
							id="column-generated"
							class="font-mono text-xs"
							bind:value={form.generatedExpression}
							disabled={valueSource === 'default' || valueSource === 'identity'}
							aria-invalid={showErrors && generatedError !== ''}
							placeholder="price * quantity"
						/>
						{#if showErrors && generatedError}
							<Field.FieldError>{generatedError}</Field.FieldError>
						{:else}
							<Field.FieldDescription>
								Stored and recomputed by postgres.
								{#if isEdit && original?.generatedExpression}
									Clearing this drops the expression and keeps the values.
								{/if}
							</Field.FieldDescription>
						{/if}
					</Field.Field>

					<Field.Field>
						<Field.FieldLabel for="column-collation">Collation</Field.FieldLabel>
						<SchemaCombobox
							id="column-collation"
							bind:value={form.collation}
							options={collationOptions.map((collation) => ({ value: collation }))}
							placeholder="Database default"
							searchPlaceholder="Search collations"
							emptyText="No collation matches."
						/>
						<Field.FieldDescription>Sort order for text columns.</Field.FieldDescription>
					</Field.Field>

					<Field.Field>
						<Field.FieldLabel for="column-comment">Comment</Field.FieldLabel>
						<Textarea id="column-comment" rows={2} bind:value={form.comment} />
					</Field.Field>
				</Field.FieldGroup>
			{/if}
		</div>

		<Sheet.Footer class="flex-row justify-end gap-2 border-t px-6 py-4 sm:space-x-0">
			<Button variant="outline" onclick={() => (open = false)} disabled={saving}>Cancel</Button>
			<Button onclick={save} disabled={loading || saving} class="relative">
				<span class={cn(saving && 'invisible')}>{isEdit ? 'Save changes' : 'Add column'}</span>
				{#if saving}
					<span class="absolute inset-0 flex items-center justify-center">
						<Spinner />
					</span>
				{/if}
			</Button>
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
