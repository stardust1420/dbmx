<script lang="ts" module>
	import * as Drawer from '$lib/components/ui/drawer/index.js';
	import DataTableCellViewer from '$lib/components/app/connections/connections-table-cell-viewer.svelte';
	import * as Dialog from "$lib/components/ui/dialog/index.js";

	let data = $state<model.ConnectionTable[]>([]);

	let dialogOpen = $state(false);

	export const columns: ColumnDef<model.ConnectionTable>[] = [
		{
			accessorKey: 'ID',
			header: '#',
			enableHiding: false,
			size: 50
		},
		{
			accessorKey: 'Name',
			header: 'Name',
			enableHiding: false,
			size: 200
		},
		{
			accessorKey: 'Env',
			header: 'Env',
			enableHiding: false,
			size: 150
		},
		{
			accessorKey: 'Engine',
			header: 'Engine',
			enableHiding: false,
		},
		{
			accessorKey: 'Host',
			header: 'Host',
			enableHiding: false,
			size: 300
		},
		{
			accessorKey: 'Database',
			header: 'Database',
			enableHiding: false,
		},
		{
			id: 'actions',
			header: 'Actions',
			cell: (props) => renderSnippet(DataTableActions, props.row.getValue('ID')),
			size: 30
		}
	];
</script>

<script lang="ts">
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
		type Row,
		type RowSelectionState,
		type SortingState,
		type VisibilityState
	} from '@tanstack/table-core';
	import type { Schema } from './schemas.js';
	import type { Attachment } from 'svelte/attachments';
	import { RestrictToVerticalAxis } from '@dnd-kit/abstract/modifiers';
	import { createSvelteTable } from '$lib/components/ui/data-table/data-table.svelte.js';
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import {
		FlexRender,
		renderComponent,
		renderSnippet
	} from '$lib/components/ui/data-table/index.js';
	import GripVerticalIcon from '@tabler/icons-svelte/icons/grip-vertical';
	import ChevronsLeftIcon from '@tabler/icons-svelte/icons/chevrons-left';
	import ChevronLeftIcon from '@tabler/icons-svelte/icons/chevron-left';
	import ChevronRightIcon from '@tabler/icons-svelte/icons/chevron-right';
	import ChevronsRightIcon from '@tabler/icons-svelte/icons/chevrons-right';
	import CircleCheckFilledIcon from '@tabler/icons-svelte/icons/circle-check-filled';
	import LoaderIcon from '@tabler/icons-svelte/icons/loader';
	import DotsVerticalIcon from '@tabler/icons-svelte/icons/dots-vertical';
	import Trash2Icon from 'lucide-svelte/icons/trash-2';

	import { toast } from 'svelte-sonner';
	import DataTableCheckbox from './connections-table-checkbox.svelte';
	import { GetAllConnections } from '$lib/wailsjs/go/app/Connections.js';
	import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
	import { model } from '$lib/wailsjs/go/models.js';
	import { onMount } from 'svelte';


	let pagination = $state<PaginationState>({ pageIndex: 0, pageSize: 10 });
	let sorting = $state<SortingState>([]);
	let columnFilters = $state<ColumnFiltersState>([]);
	let rowSelection = $state<RowSelectionState>({});
	let columnVisibility = $state<VisibilityState>({});
	const table = createSvelteTable({
		get data() {
			return data;
		},
		columns,
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
		getRowId: (row) => row.ID.toString(),
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
		},
		enableColumnResizing: true,
		columnResizeMode: 'onChange', // or 'onEnd'
	});

	onMount(() => {
		// Get All Connections
		GetAllConnections()
			.then((connections) => {
				data = connections;
			})
			.catch((error) => {
				toast.error('Failed to get connections', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	});
</script>

<div class="overflow-auto">
	<Table.Root class="border">
		<Table.Header class="bg-muted sticky top-0 z-10 overflow-auto">
			{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
				<Table.Row>
					{#each headerGroup.headers as header (header.id)}
						<Table.Head colspan={header.colSpan} style="width: {header.getSize()}px" class="text-center">
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

		<Table.Body class="**:data-[slot=table-cell]:first:w-8 overflow-auto">
				{#if table.getRowModel().rows?.length}
					{#each table.getRowModel().rows as row (row.id)}
						<Drawer.Root direction="right">
							<Table.Row data-state={row.getIsSelected() && 'selected'}>
								{#each row.getVisibleCells() as cell (cell.id)}
									<Table.Cell
										style="width: {cell.column.getSize()}px"
										class="h-12 group hover:bg-muted"
									>
									<!-- Single trigger: clicking any cell opens the one global drawer -->
									<Drawer.Trigger class="w-full">
										<span class="group-hover:underline">
										<FlexRender
											content={cell.column.columnDef.cell}
											context={cell.getContext()}
										/>
										</span>
									</Drawer.Trigger>
									</Table.Cell>
								{/each}
							</Table.Row>

							<DataTableCellViewer
								connectionID={row.getValue('ID')}
								newConnection={false}
								title={String(row.getValue('Name'))}
								description="View Connection Details"
							/>
						</Drawer.Root>

					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">
						No Connections Found. Add a new connection.
						</Table.Cell>
					</Table.Row>
				{/if}
		</Table.Body>

	</Table.Root>
</div>
<div class="flex gap-10 items-center justify-center p-2">
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
					{#each [10, 20, 30, 40, 50] as pageSize (pageSize)}
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
		<div class="ms-auto flex items-center gap-2 lg:ms-0">
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

{#snippet DataTableActions(connectionID: number)}
	<Dialog.Root bind:open={dialogOpen}>
		<Dialog.Trigger>
			<Button variant="destructive" size="icon">
				<Trash2Icon  size={16}/>
			</Button>
		</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Are you absolutely sure?</Dialog.Title>
				<Dialog.Description>
					This action cannot be undone. This will permanently delete this connection.
				</Dialog.Description>
			</Dialog.Header>
			<Dialog.Footer>
				<Button variant="outline" onclick={() => (dialogOpen = false)}>
					Cancel
				</Button>
				<Button variant="destructive" onclick={() => console.log('Delete')}>
					Delete
				</Button>
			</Dialog.Footer>
		</Dialog.Content>
	</Dialog.Root>
{/snippet}
