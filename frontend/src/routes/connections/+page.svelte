<script lang="ts">
	import ConnectionsTable from '$lib/components/app/connections/connections-table.svelte';

	import DataTableCellViewer from '$lib/components/app/connections/connections-table-cell-viewer.svelte';
	import { data } from './data';

</script>

<div class="flex h-full flex-col items-center justify-center overflow-hidden">
	<div class="flex h-full w-full flex-col self-center overflow-hidden">
		<div class="flex h-12 flex-[2] items-center justify-center pt-4">
			<h1 class="font-mono text-6xl font-bold">Connections</h1>
		</div>
		<div class="flex flex-[1] items-center justify-end mx-36">
			<DataTableCellViewer scheme="postgres" />
		</div>
		<div class="flex flex-col h-full flex-[15] items-center justify-between mb-4 mx-12 overflow-auto">
			<ConnectionsTable {data} />
		</div>
	</div>
</div>
