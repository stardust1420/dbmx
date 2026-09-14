<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as Field from '$lib/components/ui/field/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import { toast } from 'svelte-sonner';
	import { cn } from '$lib/utils.js';
	import type { model } from '$lib/wailsjs/go/models';
	import {
		AddConstraint,
		GetConstraintDefinition,
		GetSchemaEditorOptions,
		GetTableColumnsMeta,
		UpdateConstraint
	} from '$lib/wailsjs/go/app/Connections';
	import SchemaCombobox from './schema-combobox.svelte';
	import SchemaExpressionList from './schema-expression-list.svelte';

	let {
		tabID,
		tableName,
		open = $bindable(false),
		/** The constraint being edited; empty opens the sheet as "add constraint". */
		constraintName = '',
		onSaved
	}: {
		tabID: number;
		tableName: string;
		open?: boolean;
		constraintName?: string;
		onSaved?: () => void;
	} = $props();

	const isEdit = $derived(constraintName !== '');

	const CONSTRAINT_TYPES = ['PRIMARY KEY', 'UNIQUE', 'CHECK', 'FOREIGN KEY', 'EXCLUDE'] as const;
	const REFERENTIAL_ACTIONS = ['NO ACTION', 'RESTRICT', 'CASCADE', 'SET NULL', 'SET DEFAULT'];

	let name = $state('');
	let type = $state<string>('CHECK');
	let comment = $state('');
	let notValid = $state(false);

	// The payload is always the constraint body postgres itself would print. The
	// builder composes that body from fields; hand mode holds it verbatim, which is
	// how EXCLUDE, MATCH FULL, DEFERRABLE and anything else the builder does not
	// model still round-trips through this form.
	let byHand = $state(false);
	let handDefinition = $state('');

	let columns = $state<string[]>(['']);
	let checkExpression = $state('');
	let referencedTable = $state('');
	let referencedColumns = $state<string[]>(['']);
	let onUpdate = $state('NO ACTION');
	let onDelete = $state('NO ACTION');

	let tables = $state<string[]>([]);
	let columnNames = $state<string[]>([]);
	let referencedColumnNames = $state<string[]>([]);
	let loading = $state(false);
	let saving = $state(false);
	let showErrors = $state(false);

	const supportsNotValid = $derived(type === 'CHECK' || type === 'FOREIGN KEY');

	$effect(() => {
		if (open) {
			load();
		}
	});

	async function load() {
		loading = true;
		saving = false;
		showErrors = false;
		name = '';
		type = 'CHECK';
		comment = '';
		notValid = false;
		byHand = false;
		handDefinition = '';
		columns = [''];
		checkExpression = '';
		referencedTable = '';
		referencedColumns = [''];
		onUpdate = 'NO ACTION';
		onDelete = 'NO ACTION';
		referencedColumnNames = [];

		try {
			const [options, meta, existing] = await Promise.all([
				GetSchemaEditorOptions(tabID),
				GetTableColumnsMeta(tabID, tableName),
				constraintName
					? GetConstraintDefinition(tabID, tableName, constraintName)
					: Promise.resolve(null)
			]);

			tables = options.tables ?? [];
			columnNames = (meta ?? []).map((column) => column.name);

			if (existing) {
				name = existing.name;
				type = existing.type;
				comment = existing.comment;
				notValid = existing.notValid;
				handDefinition = existing.definition;
				byHand = !applyParsed(existing.type, existing.definition);
				if (!byHand && referencedTable) {
					loadReferencedColumns(referencedTable);
				}
			}
		} catch (error) {
			toast.error('Failed to open the constraint editor', { description: String(error) });
			open = false;
		} finally {
			loading = false;
		}
	}

	function loadReferencedColumns(table: string) {
		if (table === '') {
			referencedColumnNames = [];
			return;
		}
		GetTableColumnsMeta(tabID, table)
			.then((meta) => {
				referencedColumnNames = (meta ?? []).map((column) => column.name);
			})
			// A table the user has not finished typing is not an error worth a toast;
			// the picker just falls back to free text.
			.catch(() => (referencedColumnNames = []));
	}

	function splitList(list: string): string[] {
		return list
			.split(',')
			.map((entry) => entry.trim())
			.filter(Boolean);
	}

	// Reads a body postgres printed back into the builder's fields. Returns false when
	// the shape is one the builder does not model, which puts the form in hand mode
	// with the original text rather than dropping anything.
	function applyParsed(constraintType: string, definition: string): boolean {
		const body = definition.trim();

		if (constraintType === 'PRIMARY KEY' || constraintType === 'UNIQUE') {
			const match = body.match(/^(?:PRIMARY KEY|UNIQUE)\s*\(([^)]*)\)$/i);
			if (!match) return false;
			columns = splitList(match[1]);
			return columns.length > 0;
		}

		if (constraintType === 'CHECK') {
			const match = body.match(/^CHECK\s*\(([\s\S]*)\)$/i);
			if (!match) return false;
			// pg_get_constraintdef wraps the expression in its own parentheses, so the
			// captured text is usually "(price > 0)" — unwrap one balanced layer.
			checkExpression = unwrapOnce(match[1].trim());
			return checkExpression !== '';
		}

		if (constraintType === 'FOREIGN KEY') {
			const match = body.match(
				/^FOREIGN KEY\s*\(([^)]*)\)\s*REFERENCES\s+([^\s(]+)\s*\(([^)]*)\)([\s\S]*)$/i
			);
			if (!match) return false;
			columns = splitList(match[1]);
			referencedTable = match[2].replaceAll('"', '');
			referencedColumns = splitList(match[3]);

			let tail = match[4].trim();
			const updateMatch = tail.match(/ON UPDATE (NO ACTION|RESTRICT|CASCADE|SET NULL|SET DEFAULT)/i);
			const deleteMatch = tail.match(/ON DELETE (NO ACTION|RESTRICT|CASCADE|SET NULL|SET DEFAULT)/i);
			if (updateMatch) {
				onUpdate = updateMatch[1].toUpperCase();
				tail = tail.replace(updateMatch[0], '');
			}
			if (deleteMatch) {
				onDelete = deleteMatch[1].toUpperCase();
				tail = tail.replace(deleteMatch[0], '');
			}
			// MATCH FULL, DEFERRABLE and the rest are left to hand mode.
			return tail.trim() === '';
		}

		return false;
	}

	function unwrapOnce(text: string): string {
		if (!text.startsWith('(') || !text.endsWith(')')) return text;
		let depth = 0;
		for (let index = 0; index < text.length; index++) {
			if (text[index] === '(') depth++;
			else if (text[index] === ')') {
				depth--;
				// The opening paren closes before the end, so the outer pair is not a wrapper.
				if (depth === 0 && index !== text.length - 1) return text;
			}
		}
		return text.slice(1, -1).trim();
	}

	function quoteIdentifier(identifier: string): string {
		return /^[a-z_][a-z0-9_]*$/.test(identifier)
			? identifier
			: `"${identifier.replaceAll('"', '""')}"`;
	}

	const composed = $derived.by(() => {
		const keyColumns = columns.map((column) => column.trim()).filter(Boolean);
		switch (type) {
			case 'PRIMARY KEY':
				return `PRIMARY KEY (${keyColumns.join(', ')})`;
			case 'UNIQUE':
				return `UNIQUE (${keyColumns.join(', ')})`;
			case 'CHECK':
				// pg wraps the expression in a pair of its own, which applyParsed strips
				// off again on the way in.
				return `CHECK ((${checkExpression.trim()}))`;
			case 'FOREIGN KEY': {
				const parent = referencedColumns.map((column) => column.trim()).filter(Boolean);
				// No space before the parenthesis: that is how postgres prints it.
				let body = `FOREIGN KEY (${keyColumns.join(', ')}) REFERENCES ${quoteIdentifier(
					referencedTable.trim()
				)}(${parent.join(', ')})`;
				if (onUpdate !== 'NO ACTION') body += ` ON UPDATE ${onUpdate}`;
				if (onDelete !== 'NO ACTION') body += ` ON DELETE ${onDelete}`;
				return body;
			}
			default:
				return '';
		}
	});

	const definition = $derived(byHand ? handDefinition : composed);

	function changeType(next: string) {
		type = next;
		// EXCLUDE has no builder, so it is always written by hand.
		if (next === 'EXCLUDE') {
			if (!byHand) handDefinition = 'EXCLUDE USING gist ()';
			byHand = true;
		}
		if (!supportsNotValid) notValid = false;
	}

	function editByHand() {
		handDefinition = composed;
		byHand = true;
	}

	const nameError = $derived(name.trim() === '' ? 'A constraint name is required.' : '');
	const definitionError = $derived.by(() => {
		if (definition.trim() === '') return 'A constraint definition is required.';
		if (byHand) return '';
		const keyColumns = columns.map((column) => column.trim()).filter(Boolean);
		if (type !== 'CHECK' && keyColumns.length === 0) return 'Select at least one column.';
		if (type === 'CHECK' && checkExpression.trim() === '') return 'Enter a check expression.';
		if (type === 'FOREIGN KEY') {
			if (referencedTable.trim() === '') return 'Select the table this key points at.';
			if (referencedColumns.filter((column) => column.trim() !== '').length === 0)
				return 'Select at least one referenced column.';
		}
		return '';
	});
	const hasErrors = $derived(nameError !== '' || definitionError !== '');

	function save() {
		showErrors = true;
		if (hasErrors) return;

		const payload: model.ConstraintDefinition = {
			name: name.trim(),
			type,
			definition: definition.trim(),
			comment: comment.trim(),
			notValid: supportsNotValid && notValid,
			isValidated: true
		};

		saving = true;
		const request = isEdit
			? UpdateConstraint(tabID, tableName, constraintName, payload)
			: AddConstraint(tabID, tableName, payload);

		request
			.then(() => {
				toast.success(isEdit ? 'Constraint updated' : 'Constraint added', {
					description: `${payload.name} on ${tableName}.`
				});
				open = false;
				onSaved?.();
			})
			.catch((error) => {
				toast.error(isEdit ? 'Failed to update the constraint' : 'Failed to add the constraint', {
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
				{isEdit ? `Edit ${constraintName}` : `Add a constraint to ${tableName}`}
			</Sheet.Title>
			<Sheet.Description>
				{isEdit
					? 'Changing anything but the name replaces the constraint, in one transaction.'
					: 'Postgres checks the existing rows as the constraint is added, unless you skip that below.'}
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
						<Field.FieldLabel for="constraint-name">Name</Field.FieldLabel>
						<Input
							id="constraint-name"
							bind:value={name}
							placeholder={`${tableName}_column_check`}
							aria-invalid={showErrors && nameError !== ''}
						/>
						{#if showErrors && nameError}
							<Field.FieldError>{nameError}</Field.FieldError>
						{/if}
					</Field.Field>

					<Field.Field>
						<Field.FieldLabel for="constraint-type">Type</Field.FieldLabel>
						<Select.Root type="single" value={type} onValueChange={changeType}>
							<Select.Trigger id="constraint-type" class="w-full">{type}</Select.Trigger>
							<Select.Content>
								{#each CONSTRAINT_TYPES as constraintType (constraintType)}
									<Select.Item value={constraintType}>{constraintType}</Select.Item>
								{/each}
							</Select.Content>
						</Select.Root>
					</Field.Field>

					{#if !byHand}
						{#if type === 'CHECK'}
							<Field.Field data-invalid={showErrors && definitionError !== ''}>
								<Field.FieldLabel for="constraint-check">Rows must satisfy</Field.FieldLabel>
								<Textarea
									id="constraint-check"
									rows={3}
									class="font-mono text-xs"
									bind:value={checkExpression}
									aria-invalid={showErrors && definitionError !== ''}
									placeholder="price > 0 AND price < 1000"
								/>
							</Field.Field>
						{:else}
							<Field.Field data-invalid={showErrors && definitionError !== ''}>
								<Field.FieldLabel>
									{type === 'FOREIGN KEY' ? 'Columns in this table' : 'Columns'}
								</Field.FieldLabel>
								<SchemaExpressionList
									bind:entries={columns}
									{columnNames}
									idPrefix="constraint-column"
								/>
							</Field.Field>
						{/if}

						{#if type === 'FOREIGN KEY'}
							<Field.Field>
								<Field.FieldLabel for="constraint-ref-table">References table</Field.FieldLabel>
								<SchemaCombobox
									id="constraint-ref-table"
									bind:value={
										() => referencedTable,
										(next) => {
											referencedTable = next;
											referencedColumns = [''];
											loadReferencedColumns(next);
										}
									}
									options={tables.map((table) => ({ value: table }))}
									placeholder="Select a table"
									searchPlaceholder="Search tables"
									emptyText="No table matches."
								/>
							</Field.Field>

							<Field.Field>
								<Field.FieldLabel>Referenced columns</Field.FieldLabel>
								<SchemaExpressionList
									bind:entries={referencedColumns}
									columnNames={referencedColumnNames}
									idPrefix="constraint-ref-column"
								/>
								<Field.FieldDescription>
									Must be covered by a primary key or unique constraint on
									{referencedTable || 'the referenced table'}.
								</Field.FieldDescription>
							</Field.Field>

							<Field.Field>
								<Field.FieldLabel for="constraint-on-delete">
									When a referenced row is deleted
								</Field.FieldLabel>
								<Select.Root
									type="single"
									value={onDelete}
									onValueChange={(next) => (onDelete = next)}
								>
									<Select.Trigger id="constraint-on-delete" class="w-full">
										{onDelete}
									</Select.Trigger>
									<Select.Content>
										{#each REFERENTIAL_ACTIONS as action (action)}
											<Select.Item value={action}>{action}</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</Field.Field>

							<Field.Field>
								<Field.FieldLabel for="constraint-on-update">
									When a referenced key is changed
								</Field.FieldLabel>
								<Select.Root
									type="single"
									value={onUpdate}
									onValueChange={(next) => (onUpdate = next)}
								>
									<Select.Trigger id="constraint-on-update" class="w-full">
										{onUpdate}
									</Select.Trigger>
									<Select.Content>
										{#each REFERENTIAL_ACTIONS as action (action)}
											<Select.Item value={action}>{action}</Select.Item>
										{/each}
									</Select.Content>
								</Select.Root>
							</Field.Field>
						{/if}
					{/if}

					<Field.Field data-invalid={showErrors && definitionError !== ''}>
						<div class="flex items-center justify-between gap-2">
							<Field.FieldLabel for="constraint-definition">Definition</Field.FieldLabel>
							{#if !byHand}
								<Button variant="ghost" size="sm" class="h-7 text-xs" onclick={editByHand}>
									Write it myself
								</Button>
							{/if}
						</div>
						{#if byHand}
							<Textarea
								id="constraint-definition"
								rows={3}
								class="font-mono text-xs"
								bind:value={handDefinition}
								aria-invalid={showErrors && definitionError !== ''}
								placeholder="CHECK (price > 0)"
							/>
						{:else}
							<pre
								id="constraint-definition"
								class="bg-muted/50 text-muted-foreground overflow-x-auto rounded-md border px-3 py-2 font-mono text-xs">{definition ||
									'—'}</pre>
						{/if}
						{#if showErrors && definitionError}
							<Field.FieldError>{definitionError}</Field.FieldError>
						{/if}
					</Field.Field>

					{#if supportsNotValid}
						<Field.Field orientation="horizontal">
							<Field.FieldContent>
								<Field.FieldLabel for="constraint-not-valid">
									Skip checking existing rows
								</Field.FieldLabel>
								<Field.FieldDescription>
									New and changed rows are checked from now on; the rows already in the table
									are left alone.
								</Field.FieldDescription>
							</Field.FieldContent>
							<Switch id="constraint-not-valid" bind:checked={notValid} />
						</Field.Field>
					{/if}

					<Field.Field>
						<Field.FieldLabel for="constraint-comment">Comment</Field.FieldLabel>
						<Textarea id="constraint-comment" rows={2} bind:value={comment} />
					</Field.Field>
				</Field.FieldGroup>
			{/if}
		</div>

		<Sheet.Footer class="flex-row justify-end gap-2 border-t px-6 py-4 sm:space-x-0">
			<Button variant="outline" onclick={() => (open = false)} disabled={saving}>Cancel</Button>
			<Button onclick={save} disabled={loading || saving} class="relative">
				<span class={cn(saving && 'invisible')}>
					{isEdit ? 'Save changes' : 'Add constraint'}
				</span>
				{#if saving}
					<span class="absolute inset-0 flex items-center justify-center">
						<Spinner />
					</span>
				{/if}
			</Button>
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
