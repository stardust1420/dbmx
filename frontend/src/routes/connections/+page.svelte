<script lang="ts">
	import ConnectionsTable from '$lib/components/app/connections/connections-table.svelte';

	import DataTableCellViewer from '$lib/components/app/connections/connections-table-cell-viewer.svelte';
	import * as Drawer from '$lib/components/ui/drawer/index.js';
	import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
	import { Button } from '$lib/components/ui/button/index.js';
	import Plus from '@lucide/svelte/icons/plus';

	let engine = $state('PostgreSQL');

</script>

<div class="flex h-full flex-col items-center justify-center overflow-hidden">
	<div class="flex h-full w-full flex-col self-center overflow-hidden">
		<div class="flex h-12 flex-[2] items-center justify-center pt-4">
			<h1 class="font-mono text-6xl font-bold">Connections</h1>
		</div>
		<div class="flex flex-[1] items-center justify-end mx-36">
			<Drawer.Root direction="right">
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="mr-6">
						<Button variant="outline">
							<Plus class="size-4" />
							New Connection
						</Button>
					</DropdownMenu.Trigger>
					<DropdownMenu.Content class="w-40" align="start">
						<DropdownMenu.Label>Engines</DropdownMenu.Label>
						<DropdownMenu.Item class="flex w-full" onclick={() => engine = 'PostgreSQL'}>
							<Drawer.Trigger class="flex w-full">
								PostgreSQL
							</Drawer.Trigger>
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => engine = 'MySQL'}>
							MySQL
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => engine = 'ClickHouse'}>
							ClickHouse
						</DropdownMenu.Item>
						<DropdownMenu.Item onclick={() => engine = 'SQLite'}>
							SQLite
						</DropdownMenu.Item>
					</DropdownMenu.Content>
				</DropdownMenu.Root>

				<DataTableCellViewer newConnection={true} title={engine} description="Add a New Connection" />
			</Drawer.Root>

		</div>
		<div class="flex flex-col h-full flex-[15] items-center justify-between mb-4 mx-12 overflow-auto">
			<ConnectionsTable />
		</div>
	</div>
</div>
