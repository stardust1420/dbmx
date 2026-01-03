<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { GetTableInfo } from '$lib/wailsjs/go/app/Connections.js';
	import { toast } from 'svelte-sonner';
	import type { ColumnDef, RowData } from '@tanstack/table-core';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import DataTable from './data-table.svelte';

	// Accept pool id and table name as props
	let { activePoolID = '', tabName = '' } = $props();
	let selectedView = $state('structure');

	// Columns and rows for the query output data table
	let structureColumns = $state<ColumnDef<RowData, unknown>[]>([]);
	let structureRows = $state<RowData[]>([]);

	let indexesColumns = $state<ColumnDef<RowData, unknown>[]>([]);
	let indexesRows = $state<RowData[]>([]);

	let rulesColumns = $state<ColumnDef<RowData, unknown>[]>([]);
	let rulesRows = $state<RowData[]>([]);

	// Call getTableInfo on mount
	import { onMount } from 'svelte';
	onMount(() => {
		getTableInfo();
	});

	function getTableInfo() {
		if (activePoolID == '') {
			toast.error('Please connect to a database', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}
		if (tabName == '') {
			toast.error('Please select a table to execute the query', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		GetTableInfo(activePoolID, tabName)
			.then((response) => {
				// Update structure columns

				if (response.structure) {
					if (response.structure.columns) {
						for (const column of response.structure.columns) {
							structureColumns = [
								...structureColumns,
								{
									accessorKey: column,
									header: column
								}
							];
						}
					}

					// Update rows
					if (response.structure.rows) {
						for (const row of response.structure.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							structureRows = [...structureRows, cell];
						}
					}
				}

				if (response.indexes) {
					// Update columns
					if (response.indexes.columns) {
						for (const column of response.indexes.columns) {
							indexesColumns = [
								...indexesColumns,
								{
									accessorKey: column,
									header: column
								}
							];
						}
					}

					// Update rows
					if (response.indexes.rows) {
						for (const row of response.indexes.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							indexesRows = [...indexesRows, cell];
						}
					}
				}

				if (response.rules) {
					// Update columns
					if (response.rules.columns) {
						for (const column of response.rules.columns) {
							rulesColumns = [
								...rulesColumns,
								{
									accessorKey: column,
									header: column
								}
							];
						}
					}

					// Update rows
					if (response.rules.rows) {
						for (const row of response.rules.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							rulesRows = [...rulesRows, cell];
						}
					}
				}
			})
			.catch((error) => {
				toast.error('Failed to get table info', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		structureColumns = [];
		structureRows = [];
		indexesColumns = [];
		indexesRows = [];
		rulesColumns = [];
		rulesRows = [];
	}
</script>

<div class="flex h-full flex-row overflow-hidden">
	<!-- Tab Management -->
	<div class="mr-2 flex h-full flex-[1] flex-col overflow-hidden">
		<div class="mb-0.5 flex flex-1 items-center justify-center">
			<Button
				variant={selectedView === 'structure' ? 'default' : 'outline'}
				class="flex h-full w-full flex-col items-center justify-center"
				onclick={() => (selectedView = 'structure')}
			>
				{#each 'structure'.split('') as letter}
					<span class="flex-shrink-0 text-[16px] leading-none">{letter}</span>
				{/each}
			</Button>
		</div>
		<div class="mt-0.5 flex flex-1 items-center justify-center">
			<Button
				variant={selectedView === 'indexes' ? 'default' : 'outline'}
				class="flex h-full w-full flex-col items-center justify-center"
				onclick={() => (selectedView = 'indexes')}
			>
				{#each 'indexes'.split('') as letter}
					<span class="flex-shrink-0 text-[16px] leading-none">{letter}</span>
				{/each}
			</Button>
		</div>
		<div class="mt-0.5 flex flex-1 items-center justify-center">
			<Button
				variant={selectedView === 'rules' ? 'default' : 'outline'}
				class="flex h-full w-full flex-col items-center justify-center"
				onclick={() => (selectedView = 'rules')}
			>
				{#each 'rules'.split('') as letter}
					<span class="flex-shrink-0 text-[16px] leading-none">{letter}</span>
				{/each}
			</Button>
		</div>
	</div>

	<!-- Tab Content -->
	<div class="flex h-full min-h-0 flex-[15] overflow-hidden">
		{#if selectedView === 'structure'}
			<div class="flex h-full min-h-0 flex-1 flex-col overflow-hidden">
				<div class="flex h-1 min-h-0 flex-1 flex-col overflow-hidden">
					{#if structureColumns.length > 0}
						<DataTable data={structureRows} columns={structureColumns} />
					{:else}
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
					{/if}
				</div>
			</div>
		{:else if selectedView === 'indexes'}
			<div class="flex h-full min-h-0 flex-1 flex-col overflow-hidden">
				<div class="flex h-1 min-h-0 flex-1 flex-col overflow-hidden">
					{#if indexesColumns.length > 0}
						<DataTable data={indexesRows} columns={indexesColumns} />
					{:else}
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
					{/if}
				</div>
			</div>
		{:else if selectedView === 'rules'}
			<div class="flex h-full min-h-0 flex-1 flex-col overflow-hidden">
				<div class="flex h-1 min-h-0 flex-1 flex-col overflow-hidden">
					{#if rulesColumns.length > 0}
						<DataTable data={rulesRows} columns={rulesColumns} />
					{:else}
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
						<Skeleton class="my-3 h-[40px] w-full" />
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- <div>
	<h1>Manage Table</h1>
</div> -->
