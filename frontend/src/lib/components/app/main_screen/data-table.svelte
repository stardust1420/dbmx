<script lang="ts" generics="TData, TValue">
	import {
		getCoreRowModel,
		getFacetedRowModel,
		getFacetedUniqueValues,
		getFilteredRowModel,
		getPaginationRowModel,
		getSortedRowModel,
		type ColumnDef,
		type ColumnFiltersState,
		type PaginationState,
		type RowSelectionState,
		type SortingState,
		type VisibilityState
	} from '@tanstack/table-core';
	import { createSvelteTable } from '$lib/components/ui/data-table/data-table.svelte.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import {
		ColumnTypeTag,
		FlexRender,
		isCellSentinel,
		isJsonColumn
	} from '$lib/components/ui/data-table/index.js';
	import CellValueEditor from './cell-value-editor.svelte';
	import ChevronsLeftIcon from '@tabler/icons-svelte/icons/chevrons-left';
	import ChevronLeftIcon from '@tabler/icons-svelte/icons/chevron-left';
	import ChevronRightIcon from '@tabler/icons-svelte/icons/chevron-right';
	import ChevronsRightIcon from '@tabler/icons-svelte/icons/chevrons-right';
	import { toast } from 'svelte-sonner';
	import { columns, rows } from '$lib/state.svelte';

	let {
		tabID,
		tableName,
		executeQuery,
		lastQueryExecutionTime = 0
	} = $props();

	let pagination = $state<PaginationState>({ pageIndex: 0, pageSize: 20 });
	let sorting = $state<SortingState>([]);
	let columnFilters = $state<ColumnFiltersState>([]);
	let rowSelection = $state<RowSelectionState>({});
	let columnVisibility = $state<VisibilityState>({});
	let editingCell = $state<string | null>(null);
	const table = createSvelteTable({
		get data() {
			return $rows;
		},
		get columns() {
			return $columns;
		},
		state: {
			get pagination() {
				return pagination;
			},
			get sorting() {
				return sorting;
			},
			get columnVisibility() {
				return columnVisibility;
			},
			get rowSelection() {
				return rowSelection;
			},
			get columnFilters() {
				return columnFilters;
			}
		},
		enableRowSelection: true,
		getCoreRowModel: getCoreRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFacetedRowModel: getFacetedRowModel(),
		getFacetedUniqueValues: getFacetedUniqueValues(),
		getFilteredRowModel: getFilteredRowModel(),
		onPaginationChange: (updater) => {
			if (typeof updater === 'function') {
				pagination = updater(pagination);
			} else {
				pagination = updater;
			}
		},
		onSortingChange: (updater) => {
			if (typeof updater === 'function') {
				sorting = updater(sorting);
			} else {
				sorting = updater;
			}
		},
		onColumnFiltersChange: (updater) => {
			if (typeof updater === 'function') {
				columnFilters = updater(columnFilters);
			} else {
				columnFilters = updater;
			}
		},
		onColumnVisibilityChange: (updater) => {
			if (typeof updater === 'function') {
				columnVisibility = updater(columnVisibility);
			} else {
				columnVisibility = updater;
			}
		},
		onRowSelectionChange: (updater) => {
			if (typeof updater === 'function') {
				rowSelection = updater(rowSelection);
			} else {
				rowSelection = updater;
			}
		}
	});

	import { SvelteMap } from 'svelte/reactivity';
	import type { model } from '$lib/wailsjs/go/models';
	import { UpdateCells } from '$lib/wailsjs/go/app/Connections';
	import { Clock } from 'lucide-svelte';
	import ExportMenu from './export-menu.svelte';



	let editedCellsMap = $state(new SvelteMap<string, string>());
	let editingCellValue: any = $state(null);

	/**
	 * A JSON cell is too wide to read or edit on the one line the grid gives it,
	 * so double clicking one opens the cell editor instead of the inline input.
	 * `original` is the value as the query returned it, kept so that editing a
	 * cell back to what it was drops the pending change the way the inline
	 * editor does.
	 */
	type JsonEditorTarget = {
		cellId: string;
		columnName: string;
		dataType: string;
		value: string;
		original: string;
		rowId: number | null;
		/** The cell the popover hangs off, so the value opens where it lives. */
		anchor: HTMLElement | null;
	};

	let jsonEditorOpen = $state(false);
	let jsonEditorTarget = $state<JsonEditorTarget | null>(null);

	function openJsonEditor(
		cell: any,
		row: any,
		currentValue: string,
		anchor: HTMLElement | null
	) {
		const rowId = row.original['id'];
		jsonEditorTarget = {
			cellId: cell.id,
			columnName: cell.column.columnDef.header as string,
			dataType: cell.column.columnDef.meta?.columnType?.dataType ?? '',
			value: currentValue,
			original: String(cell.getValue()),
			rowId: rowId === undefined ? null : Number(rowId),
			anchor
		};
		jsonEditorOpen = true;
	}

	function applyJsonEdit(next: string) {
		const target = jsonEditorTarget;
		if (!target || target.rowId === null) return;
		if (next === target.original) {
			editedCellsMap.delete(target.cellId);
			removeUpdateCellPayload(target.cellId);
			return;
		}
		editedCellsMap.set(target.cellId, next);
		addUpdateCellPayload(target.cellId, target.rowId, target.columnName, next);
	}

	let updateCellPayload = $state<model.UpdateCell[]>([]);

	const exportFileName = $derived(String(tableName ?? '').trim() || 'query-results');

	function addUpdateCellPayload(cellId: string, rowId: number, columnId: string, value: any) {
		let payload = updateCellPayload.find((item) => item.CellID === cellId);
		if (payload) {
			payload.Value = value;
		} else {
			updateCellPayload.push({
				CellID: cellId,
				TableName: tableName,
				RowID: rowId,
				ColumnName: columnId,
				Value: value
			});
		}
	}

	function removeUpdateCellPayload(cellId: string) {
		updateCellPayload = updateCellPayload.filter((item) => item.CellID !== cellId);
	}

	function handleKeyDown(event: KeyboardEvent) {
		// The cell editor owns the keyboard while it is open, so Escape closes the
		// dialog rather than also discarding every pending edit in the grid.
		if (jsonEditorOpen) return;

        // Command/Ctrl + S
        if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
            event.preventDefault();
			if (updateCellPayload.length > 0) {
				UpdateCells(tabID, updateCellPayload)
				.then((res) => {
					if (res) {
						toast.success('Saving Changes', {
							description: 'Your changes are saved successfully.',
						});
						updateCellPayload = [];
						editedCellsMap.clear();
						executeQuery();
					}
				})
				.catch(error => {
					toast.error('Failed to save changes.', {
						description: String(error)
					});
				});
			} else {
				toast.error('No changes to save.', {
					description: 'There are no changes to save.',
				});
			}
        }

		// Clear the update payload on escape or cmd + z
		if (event.key === 'Escape' || ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'z')) {
            event.preventDefault();
			if (updateCellPayload.length > 0) {
				updateCellPayload = [];
				editedCellsMap.clear();
				toast.success('Changes Cleared', {
					description: 'Your changes have been cleared successfully.',
				});
			}
		}
    }
</script>

<svelte:document onkeydown={handleKeyDown} />

<CellValueEditor
	bind:open={jsonEditorOpen}
	anchor={jsonEditorTarget?.anchor ?? null}
	cellId={jsonEditorTarget?.cellId ?? ''}
	columnName={jsonEditorTarget?.columnName ?? ''}
	dataType={jsonEditorTarget?.dataType ?? ''}
	value={jsonEditorTarget?.value ?? ''}
	readOnly={jsonEditorTarget?.rowId === null}
	readOnlyReason="This row has no id column, so the value can be read and copied but not edited. Select the primary key in the query to edit it."
	onApply={applyJsonEdit}
/>

<div class="h-full w-full overflow-auto">
	<div class="flex h-full flex-col">
		<div class="position-sticky top-0 flex flex-1 overflow-auto rounded-lg rounded-b-3xl">
			<Table.Root class="dbmx-grid border rounded-lg rounded-b-3xl overflow-hidden">
				<Table.Header class="bg-background text-xs font-medium">
					{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
						<Table.Row class="data-[state=selected]:bg-blue-900/50">
							{#each headerGroup.headers as header (header.id)}
								<!-- The cells below are text-start with px-4, so the header reads down
								     the same left edge as the values it names. The checkbox gutter keeps
								     its own zero padding, which .select-column sets at higher specificity. -->
								<Table.Head colspan={header.colSpan}>
									{#if !header.isPlaceholder}
										<span class="inline-flex w-full items-center gap-1.5">
											<!-- The th clips its own overflow, but it cannot put an ellipsis on text
											     nested inside a flex box, so the name carries the truncation itself
											     and the icon holds its width beside it. -->
											<span class="min-w-0 truncate">
												<FlexRender
													content={header.column.columnDef.header}
													context={header.getContext()}
												/>
											</span>
											<ColumnTypeTag columnType={header.column.columnDef.meta?.columnType} />
										</span>
									{/if}
								</Table.Head>
							{/each}
						</Table.Row>
					{/each}
				</Table.Header>
				<Table.Body class="text-sm bg-background">
					{#each table.getRowModel().rows as row (row.id)}
						<Table.Row class="data-[state=selected]:bg-blue-900/50">
							{#each row.getVisibleCells() as cell (cell.id)}
								<Table.Cell
									class={`${
										editingCell === cell.id
											? 'dbmx-editing bg-cyan-500/20'
											: editedCellsMap.has(cell.id)
												? 'dbmx-editing bg-destructive hover:bg-destructive'
												: 'hover:bg-muted/50'
									} h-12 px-4 text-start focus-within:px-2 transition-[padding] w-fit`}
									ondblclick={(event: MouseEvent & { currentTarget: HTMLElement }) => {
										const currentValue = editedCellsMap.get(cell.id) ?? String(cell.getValue());
										if (isJsonColumn(cell.column.columnDef.meta?.columnType)) {
											openJsonEditor(cell, row, currentValue, event.currentTarget);
											return;
										}
										editingCell = cell.id;
										editingCellValue = currentValue;
									}}
								>
									{#if editingCell === cell.id}
										<form
											onsubmit={(e) => {
												e.preventDefault();
												if (cell.getValue() != editingCellValue) {
													if (row.original["id"] === undefined) {
														toast.error("No primary key found for the row. Please make sure the primary key is selected while querying")
														return;
													}
													editedCellsMap.set(cell.id, editingCellValue);
													addUpdateCellPayload(cell.id, Number(row.original["id"]), cell.column.columnDef.header as string, editingCellValue);
												} else {
													editedCellsMap.delete(cell.id);
													removeUpdateCellPayload(cell.id);
												}
												editingCell = null;
												editingCellValue = "";
											}}
										>
											<Input
												class="w-full px-2 text-start bg-transparent dark:bg-transparent border-0 rounded-none shadow-none focus-visible:ring-0"
												bind:value={editingCellValue}
												autofocus
												onfocusout={() => {
													editingCell = null;
													editingCellValue = "";
												}}
											/>
										</form>
									{:else}
										{#if editedCellsMap.has(cell.id)}
											{editedCellsMap.get(cell.id)}
										{:else if isCellSentinel(cell.getValue())}
											<span class="text-muted-foreground/60">{cell.getValue()}</span>
										{:else}
											<FlexRender
												content={cell.column.columnDef.cell}
												context={cell.getContext()}
											/>
										{/if}
									{/if}
								</Table.Cell>
							{/each}
						</Table.Row>
					{:else}
						<Table.Row class="data-[state=selected]:bg-blue-900/50">
							<Table.Cell colspan={$columns.length} class="h-24 text-center">No results.</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<div
			class="position-sticky bottom-0 mt-1 bg-background flex w-full items-center justify-between px-4 py-1 rounded-3xl"
		>
			<div class="flex flex-1 items-center gap-4">
				<ExportMenu {tabID} fileName={exportFileName} />
				<span class="text-muted-foreground hidden text-sm lg:flex"
					>{table.getFilteredRowModel().rows.length} row(s)</span
				>
				{#if lastQueryExecutionTime > 0}
					<span class="text-green-500 text-sm lg:flex"> <Clock size=16 class='mx-2 self-center' color='yellow' /> {lastQueryExecutionTime} ms</span>
				{/if}
			</div>
			<div class="flex w-full items-center gap-8 lg:w-fit">
				<div class="hidden items-center gap-2 lg:flex">
					<Label for="rows-per-page" class="text-sm font-medium">Rows per page</Label>
					<Select.Root
						type="single"
						bind:value={
							() => `${table.getState().pagination.pageSize}`, (v) => table.setPageSize(Number(v))
						}
					>
						<Select.Trigger class="w-20" id="rows-per-page">
							{table.getState().pagination.pageSize}
						</Select.Trigger>
						<Select.Content side="top">
							{#each [20, 30, 40, 50] as pageSize (pageSize)}
								<Select.Item value={pageSize.toString()}>
									{pageSize}
								</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				</div>
				<div class="flex w-fit items-center justify-center text-sm font-medium">
					Page {table.getState().pagination.pageIndex + 1} of
					{table.getPageCount()}
				</div>
				<div class="ml-auto flex items-center gap-2 lg:ml-0">
					<Button
						variant="outline"
						class="hidden h-8 w-8 p-0 lg:flex"
						onclick={() => table.setPageIndex(0)}
						disabled={!table.getCanPreviousPage()}
					>
						<span class="sr-only">Go to first page</span>
						<ChevronsLeftIcon />
					</Button>
					<Button
						variant="outline"
						class="size-8"
						size="icon"
						onclick={() => table.previousPage()}
						disabled={!table.getCanPreviousPage()}
					>
						<span class="sr-only">Go to previous page</span>
						<ChevronLeftIcon />
					</Button>
					<Button
						variant="outline"
						class="size-8"
						size="icon"
						onclick={() => table.nextPage()}
						disabled={!table.getCanNextPage()}
					>
						<span class="sr-only">Go to next page</span>
						<ChevronRightIcon />
					</Button>
					<Button
						variant="outline"
						class="hidden size-8 lg:flex"
						size="icon"
						onclick={() => table.setPageIndex(table.getPageCount() - 1)}
						disabled={!table.getCanNextPage()}
					>
						<span class="sr-only">Go to last page</span>
						<ChevronsRightIcon />
					</Button>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	:global(table th),
	:global(table th:first-child) {
		height: 32px;
		min-width: 100px;
		max-width: 400px;
		width: fit-content;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	:global(table td:first-child) {
		width: fit-content;
	}
	:global(table td) {
		text-align: center;
		vertical-align: middle;
		text-overflow: ellipsis;
		white-space: nowrap;
		border-right: 1px solid hsl(var(--border));
		min-width: 100px;
		max-width: 400px;
		width: fit-content;
		overflow: hidden;
	}
	:global(table td:last-child) {
		border-right: none; /* Remove border on last column */
	}

	/* Hovered cell: a crisp 1px box, no fill, so the tint on a cell holding an unsaved
	   edit still shows through. The checkbox gutter is left out. */
	:global(table.dbmx-grid tbody td:not(.select-column):hover) {
		border: 1px solid hsl(var(--grid-accent));
	}
	:global(table.dbmx-grid tbody td:has(+ td:not(.select-column):hover)) {
		border-right-color: hsl(var(--grid-accent));
	}

	/* On the final row the cell's bottom edge is shared with the table's own outer
	   border, which the table paints itself, over the cell's -- so the hover box
	   came out with three sides. It is the same contest the selected last row loses
	   above, and it is settled the same way: an overlay box, which paints in a later
	   stage and lands on top. Only the bottom line is redrawn; the other three sides
	   the cell wins outright. */
	:global(table.dbmx-grid tbody tr:last-child td:not(.select-column):hover) {
		position: relative;
	}
	:global(table.dbmx-grid tbody tr:last-child td:not(.select-column):hover::after) {
		content: '';
		position: absolute;
		inset: 0;
		border-bottom: 1px solid hsl(var(--grid-accent));
		pointer-events: none;
	}

	/* The cell being edited gets the same 1px box as a hover, in cyan to match its
	   fill, and holds it whether or not the pointer is over it. The neighbour and the
	   last-row overlay follow the hover rules' reasoning above. Each selector below is
	   paired with a :hover variant: without it the hover rule is the more specific of
	   the two on the cell being edited, and would paint its own blue back over this. */
	:global(table.dbmx-grid tbody td.dbmx-editing),
	:global(table.dbmx-grid tbody td.dbmx-editing:hover) {
		border: 1px solid hsl(var(--grid-edit-accent));
	}
	:global(table.dbmx-grid tbody td:has(+ td.dbmx-editing)),
	:global(table.dbmx-grid tbody td:has(+ td.dbmx-editing:hover)) {
		border-right-color: hsl(var(--grid-edit-accent));
	}
	:global(table.dbmx-grid tbody tr:last-child td.dbmx-editing) {
		position: relative;
	}
	:global(table.dbmx-grid tbody tr:last-child td.dbmx-editing::after),
	:global(table.dbmx-grid tbody tr:last-child td.dbmx-editing:hover::after) {
		content: '';
		position: absolute;
		inset: 0;
		border-bottom: 1px solid hsl(var(--grid-edit-accent));
		pointer-events: none;
	}
</style>
