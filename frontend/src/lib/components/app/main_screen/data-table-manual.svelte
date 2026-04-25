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
	import { FlexRender } from '$lib/components/ui/data-table/index.js';
	import ChevronsLeftIcon from '@tabler/icons-svelte/icons/chevrons-left';
	import ChevronLeftIcon from '@tabler/icons-svelte/icons/chevron-left';
	import ChevronRightIcon from '@tabler/icons-svelte/icons/chevron-right';
	import ChevronsRightIcon from '@tabler/icons-svelte/icons/chevrons-right';
	import { toast } from 'svelte-sonner';
	import { untrack } from 'svelte';
	import { columns, rows, totalRows, currentPage, currentPageSize } from '$lib/state.svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import type { model } from '$lib/wailsjs/go/models';
	import { UpdateCells } from '$lib/wailsjs/go/app/Connections';


	let {
		tabTableDBPoolID,
        tableName,
        getTablePageData
	} = $props();

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
		enableRowSelection: true,
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

        // Use untrack so that getTableData's side effects (updating rows/columns/totalRows)
        // don't create reactive dependencies that would re-trigger this effect
        untrack(() => {
            getTablePageData(String($currentPageSize), String(offset));
        });
    });

	let editedCellsMap = $state(new SvelteMap<string, string>());
	let editingCellValue: any = $state(null);

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
        // Command/Ctrl + S
        if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
            event.preventDefault();
			if (updateCellPayload.length > 0) {
				UpdateCells(tabTableDBPoolID, updateCellPayload)
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
						description: error
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
					description: 'Your changes are cleared successfully.',
				});
			}
		}
    }

</script>

<svelte:document onkeydown={handleKeyDown} />

<div class="h-full overflow-auto bg-black rounded-3xl">
	<div class="flex h-full flex-col">
		<div class="position-sticky top-0 flex flex-1 overflow-auto">
			<Table.Root class="border">
				<Table.Header class="bg-black text-xs font-medium">
					{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
						<Table.Row>
							{#each headerGroup.headers as header (header.id)}
								<Table.Head colspan={header.colSpan} class="text-center">
									{#if !header.isPlaceholder}
										<FlexRender
											content={header.column.columnDef.header}
											context={header.getContext()}
										/>
									{/if}
								</Table.Head>
							{/each}
						</Table.Row>
					{/each}
				</Table.Header>
				<Table.Body class="text-sm">
					{#each table.getRowModel().rows as row (row.id)}
						<Table.Row>
							{#each row.getVisibleCells() as cell (cell.id)}
								<Table.Cell
									class={`hover:bg-muted ${
										editedCellsMap.has(cell.id) ? 'bg-red-900 hover:bg-red-800' : ''
									} h-12 px-4 text-start focus-within:px-2 transition-[padding] w-fit`}
									ondblclick={() => {
										editingCell = cell.id;
										editingCellValue = editedCellsMap.get(cell.id) ?? String(cell.getValue());
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
													addUpdateCellPayload(cell.id, Number(row.original["id"]), cell.column.id, editingCellValue);
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
						<Table.Row>
							<Table.Cell colspan={$columns.length} class="h-24 text-center">No results.</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		<div
			class="position-sticky bottom-0 my-0.5 flex w-full items-center justify-between px-4 py-1"
		>
			<div class="text-muted-foreground hidden flex-1 text-sm lg:flex">
				Total Rows: {$totalRows}
			</div>
			<div class="text-muted-foreground hidden flex-1 text-sm lg:flex">
				Current Page Rows: {$rows?.length ?? 0} of {$totalRows}
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
	}
	:global(table td:first-child) {
		width: 200px;
	}
	:global(table td) {
		text-align: center;
		vertical-align: middle;
		text-overflow: ellipsis;
		white-space: nowrap;
		border-right: 1px solid #262734;
		max-width: 350px;
		overflow: hidden;
	}
	:global(table td:last-child) {
		border-right: none; /* Remove border on last column */
	}
</style>
