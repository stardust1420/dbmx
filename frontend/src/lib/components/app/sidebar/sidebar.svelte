<script lang="ts">
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';
	import Plus from 'lucide-svelte/icons/plus';
	import Activity from 'lucide-svelte/icons/activity';
	import Table2 from 'lucide-svelte/icons/table-2';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import { type ComponentProps } from 'svelte';
	import * as ContextMenu from '$lib/components/ui/context-menu/index.js';
	import NavUser from './user.svelte';
	import { goto } from '$app/navigation';

	function goToRoute(route: string) {
		goto(route);
	}


	let userData = {
		name: 'Stardust',
		email: 'stardust@dbmx.com',
		avatar: 'https://api.dicebear.com/9.x/glass/svg?seed=Kimberly'
	};

	let {
		ref = $bindable(null),
		tabID = $bindable(0),
		tabName = $bindable(''),
		tabTableDBPoolID = $bindable(''),
		tabConnID = $bindable(0),
		tabDBName = $bindable(''),
		onAddTab,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & {
		onAddTab?: (
			tableName?: string,
			connID?: number,
			dbName?: string,
			tableDBPoolID?: string,
			connName?: string
		) => void;
	} = $props();

	import {
		postgresConnectionsMap,
		connectionDatabasesMap,
		databasesMap,
		loadingMap,
		dbLoadingMap,
		activeDBs,
		suggestions,
		selectedDBDisplay,
		currentColor,
		activePoolID,
		tabsMap
	} from '$lib/state.svelte';

	import {
		EstablishPostgresConnection,
		EstablishPostgresDatabaseConnection,
		GetAllConnections,
		RefreshPostgresDatabase,
		TerminatePostgresDatabaseConnection
	} from '$lib/wailsjs/go/app/Connections';
	import Button from '$lib/components/ui/button/button.svelte';
	import { toast } from 'svelte-sonner';
	import { SaveActiveDBProps } from '$lib/wailsjs/go/app/Tabs';

	const refresh = () => {
		GetAllConnections()
			.then((connections) => {
				// Save a local set of connections returned
				let existingConnectionsMap = new Map(connections.map((connection) => [connection.ID, connection]));

				for (let connection of connections) {
					postgresConnectionsMap.set(connection.ID, connection);
				}

				// Check if any connection from the postgresConnectionsMap is not present in the existingConnectionsMap
				for (let connection of postgresConnectionsMap) {
					if (!existingConnectionsMap.has(connection[0])) {
						postgresConnectionsMap.delete(connection[0]);
					}
				}
			})
			.catch((error) => {
				// Handle errors from the EstablishPostgresDatabaseConnection call
				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	};

	import { onMount } from 'svelte';


	onMount(async () => {
		refresh();
	});

	function establishDatabaseConnection(id: number, dbID: string) {
		let db = databasesMap.get(dbID);

		if (!db) {
			toast.error('Database not found', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		// Check if connection is already established
		if (db.IsActive) {
			dbLoadingMap.set(dbID, false);
			return;
		}

		dbLoadingMap.set(dbID, true);

		// Establish connection
		EstablishPostgresDatabaseConnection(id, db.Name)
			.then((db) => {
				dbLoadingMap.set(dbID, false);
				databasesMap.set(dbID, db);

				$activeDBs.push(db);
				// Add tables and columns to suggestions set

				if (db.Tables) {
					suggestions.update((s) => {
						const next = new Set(s);
						db.Tables.forEach((t) => next.add(t));
						return next; // return a NEW Set
					});
				}
				if (db.Columns) {
					suggestions.update((s) => {
						const next = new Set(s);
						db.Columns.forEach((c) => next.add(c));
						return next; // return a NEW Set
					});
				}

				// If the server and database to which connection was established is the current table's database
				// Set the tab table db pool id so that it's connected
				if (tabConnID === db.ConnectionID && tabDBName === db.Name) {
					tabTableDBPoolID = db.PoolID;
				}

				$selectedDBDisplay = db.ConnectionName + ' - ' + db.Name;
				$currentColor = db.Color;
				$activePoolID = db.PoolID;
				SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

				// Update tabsMap
				let currentTab = tabsMap.get(tabID);
				if (currentTab) {
					currentTab.ActiveDB = $selectedDBDisplay;
					currentTab.ActiveDBID = $activePoolID;
					currentTab.ActiveDBColor = $currentColor;
					tabsMap.set(tabID, currentTab);
				}

				toast.success('Connected to ' + db.Name, {});
			})
			.catch((error) => {
				dbLoadingMap.set(dbID, false);

				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function establishConnection(id: number) {
		loadingMap.set(id, true);
		// Check if connection is already established
		if ((connectionDatabasesMap.get(id)?.length || 0) > 0) {
			loadingMap.set(id, false);
			return;
		}

		// Establish connection
		EstablishPostgresConnection(id)
			.then((dbs) => {
				for (let db of dbs) {
					databasesMap.set(db.ID, db);
					if (db.IsActive) {
						$activeDBs.push(db);
						// Add tables and columns to suggestions set
						if (db.Tables) {
							suggestions.update((s) => {
								const next = new Set(s);
								db.Tables.forEach((t) => next.add(t));
								return next; // return a NEW Set
							});
						}
						if (db.Columns) {
							suggestions.update((s) => {
								const next = new Set(s);
								db.Columns.forEach((c) => next.add(c));
								return next; // return a NEW Set
							});
						}

						// If the server and database to which connection was established is the current table's database
						// Set the tab table db pool id so that it's connected
						if (tabConnID === db.ConnectionID && tabDBName === db.Name) {
							tabTableDBPoolID = db.PoolID;
						}

						// Set active DB
						$selectedDBDisplay = db.ConnectionName + ' - ' + db.Name;
						$currentColor = db.Color;
						$activePoolID = db.PoolID;
						SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

						// Update tabsMap
						let currentTab = tabsMap.get(tabID);
						if (currentTab) {
							currentTab.ActiveDB = $selectedDBDisplay;
							currentTab.ActiveDBID = $activePoolID;
							currentTab.ActiveDBColor = $currentColor;
							tabsMap.set(tabID, currentTab);
						}

						toast.success('Connected to ' + db.Name, {});
					}
				}
				connectionDatabasesMap.set(
					id,
					dbs.map((db) => db.ID)
				);
				loadingMap.set(id, false);
			})
			.catch((error) => {
				loadingMap.set(id, false);
				console.log(error);
				toast.error('Connection Failed', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function getColorClass(color: string): string {
		const colorMap: Record<string, string> = {
			'bg-purple-500': 'bg-purple-500',
			'bg-indigo-500': 'bg-indigo-500',
			'bg-emerald-500': 'bg-emerald-500',
			'bg-red-500': 'bg-red-500',
			'bg-blue-500': 'bg-blue-500',
			'bg-green-500': 'bg-green-500',
			'bg-yellow-500': 'bg-yellow-500',
			'bg-orange-500': 'bg-orange-500',
			'bg-pink-500': 'bg-pink-500'
		};
		return colorMap[color] || '';
	}

	function terminateDBConnection(dbID: string) {
		let db = databasesMap.get(dbID);

		if (!db) {
			toast.error('Database not found', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		if (!db.IsActive) {
			toast.error('Connection is not active', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		TerminatePostgresDatabaseConnection(db.PoolID, db.ConnectionID)
			.then((success) => {
				if (success) {
					$activeDBs.splice(
						$activeDBs.findIndex((db) => db.ID === dbID),
						1
					); // Use splice to remove the item
					// Remove tables and columns from suggestions set (immutable update)
					const toDelete = new Set([...(db?.Tables ?? []), ...(db?.Columns ?? [])]);
					if (toDelete.size > 0) {
						$suggestions = new Set([...$suggestions].filter((x) => !toDelete.has(x)));
					}

					db.IsActive = false;
					db.Tables = [];
					db.Columns = [];
					databasesMap.delete(dbID);
					databasesMap.set(dbID, db);

					// If the active tab is connected to this db, disconnect it
					if (tabTableDBPoolID === db.PoolID) {
						tabTableDBPoolID = '';
					}

					if ($activeDBs.length == 0) {
						$selectedDBDisplay = 'Connect to a database';
						$currentColor = '';
						$activePoolID = '';
					} else {
						// Set first db as active DB
						$selectedDBDisplay = $activeDBs[0].ConnectionName + ' - ' + $activeDBs[0].Name;
						$currentColor = $activeDBs[0].Color;
						$activePoolID = $activeDBs[0].PoolID;
					}

					SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);

					// Update tabsMap
					let currentTab = tabsMap.get(tabID);
					if (currentTab) {
						currentTab.ActiveDB = $selectedDBDisplay;
						currentTab.ActiveDBID = $activePoolID;
						currentTab.ActiveDBColor = $currentColor;
						tabsMap.set(tabID, currentTab);
					}

					toast.success('Disconnected ' + db.Name, {});
				}
			})
			.catch((error) => {
				toast.error('Failed to disconnect', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function refreshDB(dbID: string) {
		dbLoadingMap.set(dbID, true);

		let db = databasesMap.get(dbID);
		if (!db) {
			dbLoadingMap.set(dbID, false);
			return;
		}

		// Establish connection
		RefreshPostgresDatabase(db.ConnectionID, dbID, db.Name, db.PoolID)
			.then((db) => {
				dbLoadingMap.set(dbID, false);
				databasesMap.set(db.ID, db);
			})
			.catch((error) => {
				dbLoadingMap.set(dbID, false);

				toast.error('Failed to refresh', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	// Function to handle adding a new tab
	function handleAddTab(
		tableName?: string,
		connID?: number,
		dbName?: string,
		tableDBPoolID?: string,
		connName?: string
	) {
		if (onAddTab) {
			onAddTab(tableName, connID, dbName, tableDBPoolID, connName);
		}
	}
</script>

<Sidebar.Root
	bind:ref
	{tabID}
	{tabName}
	{tabTableDBPoolID}
	{tabConnID}
	{tabDBName}
	{...restProps}
	variant="floating"
>
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>
				<div class="flex w-full items-center justify-between">
					Connections
					<div class="flex gap-1">
						<Button size="sm" variant="ghost" onclick={() => refresh()}>
							<RefreshCw />
						</Button>
					</div>
				</div>
			</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					<!-- Check if postgresConnectionsMap is empty -->
					{#if postgresConnectionsMap.size === 0}
						<Sidebar.MenuItem class="h-96 flex flex-col items-center justify-center">
							No connections found
							<Button size="sm" variant="link" onclick={() => goToRoute('/connections')}>
								<Plus /> Add Connection
							</Button>
						</Sidebar.MenuItem>
					{:else}
						{#each Array.from(postgresConnectionsMap.entries()) as [key, connection]}
							<Sidebar.MenuItem
								class="{getColorClass(connection.Color)} bg-opacity-20 hover:bg-opacity-25"
							>
								<Collapsible.Root>
									<Collapsible.Trigger onclick={() => establishConnection(connection.ID)}>
										<Sidebar.MenuButton class="transform-none select-none transition-none">
											<ChevronRight className="transition-transform" />

											{connection.Name}
										</Sidebar.MenuButton>
									</Collapsible.Trigger>
									<Collapsible.Content>
										<Sidebar.MenuSub>
											{#if loadingMap.get(connection.ID)}
												<Sidebar.MenuSkeleton />
												<Sidebar.MenuSkeleton />
												<Sidebar.MenuSkeleton />
												<Sidebar.MenuSkeleton />
											{:else if (connectionDatabasesMap.get(connection.ID) || []).length > 0}
												{#each connectionDatabasesMap.get(connection.ID) || [] as databaseID}
													<Collapsible.Root open={databasesMap.get(databaseID)?.IsActive}>
														<Collapsible.Trigger
															onclick={() => establishDatabaseConnection(connection.ID, databaseID)}
														>
															<ContextMenu.Root>
																<ContextMenu.Trigger>
																	<Sidebar.MenuButton
																		class="transform-none select-none transition-none"
																		aria-disabled={!databasesMap.get(databaseID)?.IsActive}
																	>
																		<ChevronRight className="transition-transform" />
																		{databasesMap.get(databaseID)?.Name}
																		{#if databasesMap.get(databaseID)?.IsActive}
																			<Activity color="#4fff4d" />
																		{/if}
																	</Sidebar.MenuButton>
																</ContextMenu.Trigger>
																<ContextMenu.Content>
																	<ContextMenu.Item onclick={() => terminateDBConnection(databaseID)}
																		>Disconnect
																	</ContextMenu.Item>
																	<ContextMenu.Item onclick={() => refreshDB(databaseID)}
																		>Refresh
																	</ContextMenu.Item>
																</ContextMenu.Content>
															</ContextMenu.Root>
														</Collapsible.Trigger>
														<Collapsible.Content>
															<Sidebar.MenuSub>
																{#if dbLoadingMap.get(databaseID)}
																	<Sidebar.MenuSkeleton />
																	<Sidebar.MenuSkeleton />
																	<Sidebar.MenuSkeleton />
																	<Sidebar.MenuSkeleton />
																{:else if (databasesMap.get(databaseID)?.Tables || []).length > 0}
																	{#each databasesMap.get(databaseID)?.Tables || [] as table}
																		<ContextMenu.Root>
																			<ContextMenu.Trigger>
																				<Sidebar.MenuButton
																					ondblclick={() =>
																						handleAddTab(
																							table,
																							connection.ID,
																							databasesMap.get(databaseID)?.Name,
																							databasesMap.get(databaseID)?.PoolID,
																							connection.Name
																						)}
																				>
																					<Table2 color="#fd6868" strokeWidth={2} size={25} />
																					<p>{table}</p>
																				</Sidebar.MenuButton>
																			</ContextMenu.Trigger>
																			<ContextMenu.Content>
																				<ContextMenu.Item
																					onclick={() =>
																						handleAddTab(
																							table,
																							connection.ID,
																							databasesMap.get(databaseID)?.Name,
																							databasesMap.get(databaseID)?.PoolID,
																							connection.Name
																						)}
																				>
																					<Plus class="mr-2 h-4 w-4" />
																					Open in New Tab
																				</ContextMenu.Item>
																			</ContextMenu.Content>
																		</ContextMenu.Root>
																	{/each}
																{:else}
																	No tables found
																{/if}
															</Sidebar.MenuSub>
														</Collapsible.Content>
													</Collapsible.Root>
												{/each}
											{:else}
												No databases found
											{/if}
										</Sidebar.MenuSub>
									</Collapsible.Content>
								</Collapsible.Root>
							</Sidebar.MenuItem>
						{/each}
					{/if}
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<NavUser user={userData} />
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
