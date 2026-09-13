<script lang="ts" generics="TData, TValue">
	import {
		getCoreRowModel,
		getFacetedRowModel,
		getFacetedUniqueValues,
		getFilteredRowModel,
		getSortedRowModel,
		type Cell,
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
	import { ColumnTypeTag, FlexRender, isJsonColumn } from '$lib/components/ui/data-table/index.js';
	import CellValueEditor from './cell-value-editor.svelte';
	import ChevronsLeftIcon from '@tabler/icons-svelte/icons/chevrons-left';
	import ChevronLeftIcon from '@tabler/icons-svelte/icons/chevron-left';
	import ChevronRightIcon from '@tabler/icons-svelte/icons/chevron-right';
	import ChevronsRightIcon from '@tabler/icons-svelte/icons/chevrons-right';
	import { toast } from 'svelte-sonner';
	import { untrack } from 'svelte';
	import { columns, rows, totalRows, currentPage, currentPageSize } from '$lib/state.svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import type { model } from '$lib/wailsjs/go/models';
	import { DeleteRows, UpdateCells } from '$lib/wailsjs/go/app/Connections';
	import { Clock, Plus, Trash2 } from 'lucide-svelte';
	import AddRowSheet from './add-row-sheet.svelte';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import ExportMenu from './export-menu.svelte';


	let {
		tabID,
        tableName,
        getTablePageData,
		lastQueryExecutionTime = 0,
		select = '',
		where = '',
		orderBy = '',
		groupBy = ''
	} = $props();

	let sorting = $state<SortingState>([]);
	let columnFilters = $state<ColumnFiltersState>([]);
	let rowSelection = $state<RowSelectionState>({});
	let columnVisibility = $state<VisibilityState>({});
	let editingCell = $state<string | null>(null);

	// The grid addresses a row by a column literally named "id", the same contract the
	// inline cell editor and UpdateCells already rely on. A result set that omits it
	// (a projection, a join, an aggregate) yields rows that cannot be deleted.
	function hasRowID(row: any): boolean {
		return row?.['id'] !== undefined && row?.['id'] !== null;
	}

	const table = createSvelteTable({
		get data() {
			return $rows;
		},
		get columns() {
			return $columns;
		},
		state: {
			get pagination() {
				return { pageIndex: $currentPage, pageSize: $currentPageSize };
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
		// Key the selection by the row's own primary key rather than its position, so a
		// selection can never outlive a page fetch and end up pointing at a different row.
		getRowId: (row: any, index: number) => (hasRowID(row) ? String(row['id']) : `unkeyed-${index}`),
		// A row with no id cannot be addressed by DeleteRows, so make it unselectable up
		// front instead of letting the delete fail later.
		enableRowSelection: (row) => hasRowID(row.original),
		getCoreRowModel: getCoreRowModel(),
		manualPagination: true,
        get rowCount() {
            return $totalRows;
        },
		getSortedRowModel: getSortedRowModel(),
		getFacetedRowModel: getFacetedRowModel(),
		getFacetedUniqueValues: getFacetedUniqueValues(),
		getFilteredRowModel: getFilteredRowModel(),
		onPaginationChange: (updater) => {
			let nextState: PaginationState;
			if (typeof updater === 'function') {
				nextState = updater({ pageIndex: $currentPage, pageSize: $currentPageSize });
			} else {
				nextState = updater;
			}
			currentPage.set(nextState.pageIndex);
			currentPageSize.set(nextState.pageSize);
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

    // Fetch data whenever pagination changes (user-driven only)
    let isInitialMount = true;
    $effect(() => {
        // Track only pagination state via stores
		let offset = $currentPage * $currentPageSize;

        // Skip the initial mount — parent already loaded the data
        if (isInitialMount) {
            isInitialMount = false;
            return;
        }

        // A fetched page replaces every row, so no selection survives it.
        rowSelection = {};

        // Use untrack so that getTableData's side effects (updating rows/columns/totalRows)
        // don't create reactive dependencies that would re-trigger this effect
        untrack(() => {
            getTablePageData(String($currentPageSize), String(offset));
        });
    });

	let editedCellsMap = $state(new SvelteMap<string, string>());
	let editingCellValue: any = $state(null);

	/**
	 * A JSON cell is too wide to read or edit on the one line the grid gives it,
	 * so double clicking one opens the cell editor instead of the inline input.
	 * `original` is the value as the page load returned it, kept so that editing
	 * a cell back to what it was drops the pending change the way the inline
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

	let addRowOpen = $state(false);
	let deleteConfirmOpen = $state(false);
	let deleting = $state(false);

	const pageRows = $derived(table.getRowModel().rows);
	const selectedRows = $derived(pageRows.filter((row) => row.getIsSelected()));
	const selectableRowCount = $derived(pageRows.filter((row) => row.getCanSelect()).length);
	const allRowsSelected = $derived(
		selectableRowCount > 0 && selectedRows.length === selectableRowCount
	);

	function deleteSelectedRows() {
		const ids = selectedRows.map((row) => String((row.original as any)['id']));
		if (ids.length === 0) {
			return;
		}

		deleting = true;
		DeleteRows(tabID, tableName, ids)
			.then((deletedCount) => {
				// Drop only the pending edits that belonged to the deleted rows; edits the
				// user has queued on surviving rows stay queued.
				const staleCellIDs = updateCellPayload
					.filter((item) => ids.includes(String(item.RowID)))
					.map((item) => item.CellID);
				updateCellPayload = updateCellPayload.filter(
					(item) => !staleCellIDs.includes(item.CellID)
				);
				for (const cellID of staleCellIDs) {
					editedCellsMap.delete(cellID);
				}

				deleteConfirmOpen = false;
				rowSelection = {};
				totalRows.update((total) => Math.max(0, total - deletedCount));
				getTablePageData(String($currentPageSize), String($currentPage * $currentPageSize));

				toast.success(`Deleted ${deletedCount} ${deletedCount === 1 ? 'row' : 'rows'}`, {
					description: `${deletedCount === 1 ? 'The row was' : 'The rows were'} removed from ${tableName}.`
				});
			})
			.catch((error) => {
				toast.error('Failed to delete rows.', {
					description: String(error)
				});
			})
			.finally(() => {
				deleting = false;
			});
	}

	// Refresh the page the user is on. The page fetch skips the COUNT(*), so the
	// total is adjusted here for the single row that was just inserted.
	function onRowInserted() {
		totalRows.update((total) => total + 1);
		getTablePageData(String($currentPageSize), String($currentPage * $currentPageSize));
	}

	let updateCellPayload = $state<model.UpdateCell[]>([]);

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
		// The add-row sheet owns the keyboard while it is open, so Escape closes it
		// instead of discarding pending cell edits behind it.
		if (addRowOpen || deleteConfirmOpen || jsonEditorOpen) {
			return;
		}

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
						getTablePageData(String($currentPageSize), String($currentPage * $currentPageSize));
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

<AddRowSheet {tabID} {tableName} bind:open={addRowOpen} onInserted={onRowInserted} />

<Dialog.Root bind:open={deleteConfirmOpen}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>
				Delete {selectedRows.length}
				{selectedRows.length === 1 ? 'row' : 'rows'}?
			</Dialog.Title>
			<Dialog.Description>
				{selectedRows.length === 1 ? 'This row' : 'These rows'} will be permanently deleted from
				<span class="font-medium">{tableName}</span>. This cannot be undone.
			</Dialog.Description>
		</Dialog.Header>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (deleteConfirmOpen = false)} disabled={deleting}>
				Cancel
			</Button>
			<Button variant="destructive" onclick={deleteSelectedRows} disabled={deleting}>
				{deleting ? 'Deleting...' : 'Delete'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<div class="h-full w-full overflow-auto">
	<div class="flex h-full flex-col">
		<div class="position-sticky top-0 flex flex-1 overflow-auto rounded-3xl">
			<Table.Root class="dbmx-grid border rounded-3xl overflow-hidden">
				<Table.Header class="bg-background text-xs font-medium">
					{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
						<Table.Row class="data-[state=selected]:bg-blue-900/50">
							<Table.Head class="select-column">
								<Checkbox
									class="mx-auto"
									checked={allRowsSelected}
									indeterminate={selectedRows.length > 0 && !allRowsSelected}
									disabled={selectableRowCount === 0}
									aria-label="Select all rows on this page"
									onCheckedChange={(value: boolean) => table.toggleAllRowsSelected(value === true)}
								/>
							</Table.Head>
							{#each headerGroup.headers as header (header.id)}
								<!-- The cells below are text-start with px-4, so the header reads down
								     the same left edge as the values it names. The checkbox gutter keeps
								     its own zero padding, which .select-column sets at higher specificity. -->
								<Table.Head colspan={header.colSpan} class="px-4">
									{#if !header.isPlaceholder}
										<span class="items-center justify-center">
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
						<Table.Row class="data-[state=selected]:bg-blue-900/50" data-state={row.getIsSelected() ? 'selected' : undefined}>
							<Table.Cell class="select-column">
								<Checkbox
									class="mx-auto"
									checked={row.getIsSelected()}
									disabled={!row.getCanSelect()}
									aria-label={row.getCanSelect()
										? 'Select row'
										: 'This row has no id column, so it cannot be selected'}
									onCheckedChange={(value: boolean) => row.toggleSelected(value === true)}
								/>
							</Table.Cell>
							{#each row.getVisibleCells() as cell (cell.id)}
								<Table.Cell
									class={`${
										editedCellsMap.has(cell.id) ? 'bg-destructive/20 hover:bg-destructive/30' : ''
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
												class="hover:bg-input/30 px-2 focus-visible:bg-background dark:hover:bg-input/30 dark:focus-visible:bg-input/30 w-full bg-transparent text-start shadow-none focus-visible:border dark:bg-transparent"
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
							<Table.Cell colspan={$columns.length + 1} class="h-24 text-center">No results.</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<div
			class="position-sticky bottom-0 mt-1 bg-background flex w-full items-center justify-between px-4 py-1 rounded-3xl"
		>
			<div class="flex flex-1 items-center gap-3">
				<Button variant="outline" size="sm" class="h-8" onclick={() => (addRowOpen = true)}>
					<Plus data-icon="inline-start" />
					Add Row
				</Button>
				<ExportMenu
					{tabID}
					fileName={tableName}
					{tableName}
					{select}
					{where}
					{orderBy}
					{groupBy}
				/>
				{#if selectedRows.length > 0}
					<Button
						variant="destructive"
						size="sm"
						class="h-8"
						onclick={() => (deleteConfirmOpen = true)}
					>
						<Trash2 data-icon="inline-start" />
						Delete {selectedRows.length}
						{selectedRows.length === 1 ? 'row' : 'rows'}
					</Button>
				{/if}
				<span class="text-muted-foreground hidden text-sm lg:flex">
					Total Rows: {$totalRows}
				</span>
			</div>
			{#if lastQueryExecutionTime > 0}
				<span class="text-green-500 text-sm lg:flex flex-1"> <Clock size=16 class='mx-2 self-center' color='yellow' /> {lastQueryExecutionTime} ms</span>
			{/if}
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
					Page: {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
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
		/* width: fit-content; */
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
	/* The checkbox column is sized to its content. It has to opt out of the blanket
	   100px min-width above, which the data columns depend on, and it must come last
	   to outrank the equally specific `table td:first-child` rule. */
	:global(table th.select-column),
	:global(table td.select-column) {
		width: 2.75rem;
		min-width: 2.75rem;
		max-width: 2.75rem;
		/* Both the padding and the divider are kept out of the layout so the content box
		   is the full 2.75rem and the checkbox's `mx-auto` lands on the cell's true
		   middle. The cells are border-box, so a border-right and side padding would
		   otherwise shrink the content box and shift its centre left. The divider is
		   drawn as an inset shadow, which paints the same 1px line but takes no space. */
		padding: 0;
		border-right: none;
		box-shadow: inset -1px 0 0 hsl(var(--border));
		/* The checkbox root is `display: flex`, i.e. a block-level box that text-align
		   cannot move; `mx-auto` at the call site is what actually centres it. */
		text-align: center;
	}
	/* Hover and selection ship as a muted background tint, which is the same colour as
	   the panel the grid sits on and so reads as no feedback at all. They are redrawn
	   here as a tinted fill plus a crisp 1px edge. Rules are scoped to .dbmx-grid both to
	   outrank the shared Table.Row utilities and to keep them off every other table in
	   the app.

	   The table is border-collapse: collapse, which decides every shared grid line, so
	   which element declares an edge matters as much as its colour:
	     - horizontal edges are borders on the cell, because the collapse cascade resolves
	       a width/style tie by element (cell > row > table) and so beats the border-b that
	       Table.Row puts on every row;
	     - the block's outer left and right edges are inset shadows instead, because those
	       lines are shared with the table's own outer border and painting inside the cell
	       sidesteps that contest entirely;
	     - a hovered cell's left edge is shared with the previous cell's border-right, and
	       between two cells the cascade gives the line to the one further left, so the
	       neighbour is recoloured rather than the hovered cell.
	   Every line stays 1px and every boundary already had one, so nothing shifts. */
	:global(table.dbmx-grid) {
		/* One place to retune the highlight. The radius has to track the rounded-3xl on
		   Table.Root, which is what the table's overflow clips to. */
		--grid-accent: 217 91% 60%;
		--grid-radius: 1.5rem;
	}

	/* A selected row is marked by its edges alone, with no fill. The background still has
	   to be stated: leaving it out would let Table.Row's own data-[state=selected]:bg-muted
	   paint the muted tint back in. */
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td) {
		background-color: transparent;
		border-top: 1px solid hsl(var(--grid-accent));
		border-bottom: 1px solid hsl(var(--grid-accent));
	}

	/* Row hover: a tint that is actually distinguishable from the panel behind the grid.
	   It comes after the selected-row rule, which it ties with on specificity, so that a
	   selected row still responds to the pointer. */
	:global(table.dbmx-grid tbody tr:hover > td) {
		background-color: hsl(var(--grid-accent) / 0.07);
	}
	/* The outer left and right edges are drawn as positioned pseudo-elements rather than
	   a border or an inset shadow on the cell. In a collapsed-border table the shared
	   lines -- including the table's own outer border, which these two edges sit on -- are
	   painted by the table, over cell backgrounds and box-shadows alike, so anything
	   drawn on the cell there is covered. A positioned box paints in a later stage and
	   lands on top. The first cell keeps its own divider box-shadow untouched this way. */
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td) {
		position: relative;
	}
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td:first-child::before),
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td:last-child::before) {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		width: 1px;
		background-color: hsl(var(--grid-accent));
		pointer-events: none;
	}
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td:first-child::before) {
		left: 0;
	}
	:global(table.dbmx-grid tbody tr[data-state='selected'] > td:last-child::before) {
		right: 0;
	}

	/* The final row's lower edge sits on the table's own outer bottom border and loses it
	   the same way, so it is drawn as a pseudo-element too. Only this one row needs it --
	   every interior row boundary is won outright by the cell border above.

	   It is an overlay box rather than a 1px line because this row also has to follow the
	   table's rounded bottom corners: a corner is where the side edge meets the bottom
	   one, which two separate straight lines cannot round. So each end cell draws its
	   side and bottom edge as one bordered box carrying the radius, and the plain side
	   line is suppressed there to keep it from cutting across the curve. */
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td) {
		/* Drawn by the overlay below, curve included. */
		border-bottom-color: transparent;
	}
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td::after) {
		content: '';
		position: absolute;
		inset: 0;
		border-bottom: 1px solid hsl(var(--grid-accent));
		pointer-events: none;
	}
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td:first-child::after) {
		border-left: 1px solid hsl(var(--grid-accent));
		border-bottom-left-radius: var(--grid-radius);
	}
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td:last-child::after) {
		border-right: 1px solid hsl(var(--grid-accent));
		border-bottom-right-radius: var(--grid-radius);
	}
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td:first-child::before),
	:global(table.dbmx-grid tbody tr[data-state='selected']:last-child > td:last-child::before) {
		display: none;
	}

	/* Adjacent selected rows read as one block: the boundary between two of them falls
	   back to the ordinary divider colour instead of an accent edge, so a run of
	   selections is bounded once rather than outlined row by row. Both sides of the
	   shared line have to be set, since either cell's border can win the collapse. */
	:global(table.dbmx-grid tbody tr[data-state='selected'] + tr[data-state='selected'] > td) {
		border-top-color: hsl(var(--border));
	}
	:global(table.dbmx-grid tbody tr[data-state='selected']:has(+ tr[data-state='selected']) > td) {
		border-bottom-color: hsl(var(--border));
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
</style>
