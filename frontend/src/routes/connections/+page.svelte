<script lang="ts">
	import NewConnection from './new/+page.svelte';
	import data from './data.ts';
	import ConnectionsTable from '$lib/components/app/connections/connections-table.svelte';
</script>

<div class="flex h-full flex-col items-center justify-center overflow-hidden">
	<div class="flex h-full w-full flex-col self-center overflow-hidden">
		<div class="flex h-12 flex-[1] items-center justify-center p-4">
			<h1 class="font-mono text-6xl font-bold">Connections</h1>
		</div>
		<div class="flex w-full flex-col flex-[9] items-center justify-between mb-4 overflow-auto">
			<ConnectionsTable {data} />
		</div>
	</div>
</div>
