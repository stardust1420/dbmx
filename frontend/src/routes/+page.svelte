<script lang="ts">
	import AppSidebar from '$lib/components/app/sidebar/sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import Tabs from '$lib/components/app/main_screen/tabs.svelte';

	// Active tab properties
	let tabID = $state(0);
	let tabName = $state('');
	let tabType = $state('');

	// Table tab properties
	let tabTableDBPoolID = $state('');
	let tabPostgresConnName = $state('');
	let tabDBName = $state('');
	let tabPostgresConnID = $state(0);

	let select = $state('');
	let limit = $state('');
	let offset = $state('');
	let where = $state('');
	let orderBy = $state('');
	let groupBy = $state('');
	let tableColumns = $state([]);

	// Reference to the Tabs component
	let tabsComponent: Tabs;

	// Function to handle adding a new tab from sidebar
	function handleAddTab(
		tableName?: string,
		postgresConnID?: number,
		dbName?: string,
		tableDBPoolID?: string,
		postgresConnName?: string
	) {
		if (tabsComponent && tabsComponent.addTab) {
			tabsComponent.addTab(tableName, postgresConnID, dbName, tableDBPoolID, postgresConnName);
		}
	}
</script>

<Sidebar.Provider>
	<AppSidebar
		bind:tabID
		bind:tabName
		bind:tabTableDBPoolID
		bind:tabPostgresConnID
		bind:tabDBName
		onAddTab={handleAddTab}
	/>
	<Sidebar.Inset>
		<Tabs
			bind:this={tabsComponent}
			bind:tabID
			bind:tabName
			bind:tabType
			bind:tabTableDBPoolID
			bind:tabPostgresConnName
			bind:tabDBName
			bind:tabPostgresConnID
			bind:select
			bind:limit
			bind:offset
			bind:where
			bind:orderBy
			bind:groupBy
			bind:tableColumns
		/>
	</Sidebar.Inset>
</Sidebar.Provider>
