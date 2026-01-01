<script lang="ts">
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import X from 'lucide-svelte/icons/x';
	import Plus from 'lucide-svelte/icons/plus';
	import Chat from 'lucide-svelte/icons/message-circle-more';
	import * as Resizable from '$lib/components/ui/resizable/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { onMount, type ComponentProps } from 'svelte';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { LineScale } from 'svelte-spins';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Collapsible from '$lib/components/ui/collapsible/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';

	import * as Select from '$lib/components/ui/select/index.js';

	// Import our custom components
	import SqlEditor from '$lib/components/app/main_screen/sql_editor.svelte';
	import {
		AddTab,
		DeleteTab,
		SetActiveTab,
		UpdateTabEditorContent,
		SaveActiveDBProps,
		GetAllTabs
	} from '$lib/wailsjs/go/app/Tabs';
	import { activeDBs, suggestions } from '$lib/state.svelte';
	import {
		tabsMap,
		selectedDBDisplay,
		currentColor,
		activePoolID,
		selectedQuery
	} from '$lib/state.svelte';

	import { toast } from 'svelte-sonner';
	import { ExecuteQuery, GetTableData } from '$lib/wailsjs/go/app/Connections.js';

	import DataTable from './data-table.svelte';
	import { columns, rows } from '$lib/state.svelte';
	import ManageTable from './manage_table.svelte';

	let editorHeight = $state(50); // Percentage of the container height
	let outputHeight = $state(50); // Percentage of the container height

	let queryLoading = $state(false);
	let tabLoading = $state(false);

	// Handle Tabs

	// Active tab properties
	let {
		tabID = $bindable(0),
		tabName = $bindable(''),
		tabType = $bindable(''),
		tabDBName = $bindable(''),
		tabTableDBPoolID = $bindable(''),
		tabConnName = $bindable(''),
		tabConnID = $bindable(0),
		select = $bindable(''),
		limit = $bindable(''),
		offset = $bindable(''),
		where = $bindable(''),
		orderBy = $bindable(''),
		groupBy = $bindable(''),
		tableColumns = $bindable([]),
		toggleChatPane
	} = $props();
	let editor = $state('');

	let isSelectDropdownOpen = $state(false);
	let isWhereDropdownOpen = $state(false);
	let isOrderByDropdownOpen = $state(false);
	let isGroupByDropdownOpen = $state(false);

	// Table view tab state (for Data/Structure/Indexes)
	let tableViewTab = $state('data');

	onMount(() => {
		getAllTabs();
	});

	function getAllTabs() {
		tableViewTab = 'data';

		GetAllTabs().then((tabs) => {
			if (!tabs) {
				return;
			}
			for (const tab of tabs) {
				tabsMap.set(tab.ID, tab);

				// Set active tab properties
				if (tab.IsActive) {
					tabID = tab.ID;
					tabName = tab.Name;
					tabType = tab.Type;

					// Properties for table view tab
					tabDBName = tab.DBName || '';
					tabTableDBPoolID = tab.ActiveDBID || '';
					tabConnName = tab.ConnectionName || '';
					tabConnID = tab.ConnectionID || 0;

					select = tab.Select;
					limit = tab.Limit;
					offset = tab.Offset;
					where = tab.Where;
					orderBy = tab.OrderBy;
					groupBy = tab.GroupBy;
					tableColumns = tab.TableColumnsList;

					editor = tab.Editor;

					if ($activeDBs.length == 0) {
						$selectedDBDisplay = 'Connect to a database';
						$currentColor = '';
						$activePoolID = '';
					} else {
						$selectedDBDisplay = tab.ActiveDB || 'Connect to a database';
						$activePoolID = tab.ActiveDBID || '';
						$currentColor = tab.ActiveDBColor || '';
					}

					// Update columns
					if (tab.columns) {
						for (const column of tab.columns) {
							columns.set([
								...$columns,
								{
									accessorKey: column,
									header: column
								}
							]);
						}
					}

					// Update rows
					if (tab.rows) {
						for (const row of tab.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							rows.set([...$rows, cell]);
						}
					}
				}
			}
		});

		columns.set([]);
		rows.set([]);
	}

	export function addTab(
		tableName: string = '',
		connID: number = 0,
		dbName: string = '',
		tableDBPoolID: string = '',
		connName: string = ''
	) {
		$selectedQuery = '';

		tableViewTab = 'data';

		tabLoading = true;
		let tableType = 'editor';
		let tabDisplayName = 'Editor';

		// If table's pool id is provided, use that instead of the active pool id
		// This is used when we want to open a table in a new tab
		// In case of table, we want the pool id of the table's database

		// First assign the active global pool id
		let poolID: string = $activePoolID;

		// If table's pool id is provided, use that instead of the active pool id
		if (tableName.trim() != '') {
			tableType = 'table';
			tabDisplayName = tableName;
			poolID = tableDBPoolID;
		}
		// Send default values for now in activeDBID and activeDB
		AddTab(
			poolID,
			$selectedDBDisplay,
			$currentColor,
			tabDisplayName,
			tableType,
			connID,
			dbName,
			connName
		)
			.then((tab) => {
				queryLoading = false;
				tabsMap.set(tab.ID, tab);

				tabID = tab.ID;
				tabName = tab.Name;
				tabType = tab.Type;

				// Properties for table view tab
				tabDBName = tab.DBName || '';
				tabTableDBPoolID = tab.ActiveDBID || '';
				tabConnName = tab.ConnectionName || '';
				tabConnID = tab.ConnectionID || 0;

				select = tab.Select;
				limit = tab.Limit;
				offset = tab.Offset;
				where = tab.Where;
				orderBy = tab.OrderBy;
				groupBy = tab.GroupBy;
				tableColumns = tab.TableColumnsList;

				editor = tab.Editor;
			})
			.catch((error) => {
				toast.error('Failed to add tab', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
		
		tabLoading = false;

		$selectedDBDisplay = $selectedDBDisplay;
		$activePoolID = $activePoolID;
		$currentColor = $currentColor;

		columns.set([]);
		rows.set([]);
	}

	function deleteTab(id: number) {
		$selectedQuery = '';
		// Delete the old tab from the map
		tabsMap.delete(id);
		tableViewTab = 'data';

		DeleteTab(id)
			.then((tab) => {
				if (tab) {
					queryLoading = false;

					// Set the new tab as active
					tabsMap.set(tab.ID, tab);

					tabID = tab.ID;
					tabName = tab.Name;
					tabType = tab.Type;

					// Properties for table view tab
					tabDBName = tab.DBName || '';
					tabTableDBPoolID = tab.ActiveDBID || '';
					tabConnName = tab.ConnectionName || '';
					tabConnID = tab.ConnectionID || 0;

					select = tab.Select;
					limit = tab.Limit;
					offset = tab.Offset;
					where = tab.Where;
					orderBy = tab.OrderBy;
					groupBy = tab.GroupBy;
					tableColumns = tab.TableColumnsList;

					editor = tab.Editor;

					$selectedDBDisplay = tab.ActiveDB || 'Connect to a database';
					$activePoolID = tab.ActiveDBID || '';
					$currentColor = tab.ActiveDBColor || '';

					// Update columns
					if (tab.columns) {
						for (const column of tab.columns) {
							columns.set([
								...$columns,
								{
									accessorKey: column,
									header: column
								}
							]);
						}
					}

					// Update rows
					if (tab.rows) {
						for (const row of tab.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							rows.set([...$rows, cell]);
						}
					}
				}
			})
			.catch((error) => {
				toast.error('Failed to delete tab', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		columns.set([]);
		rows.set([]);
	}

	function setActiveTab(id: number) {
		$selectedQuery = '';
		tableViewTab = 'data';

		SetActiveTab(id)
			.then((tab) => {
				queryLoading = false;
				tabsMap.set(tab.ID, tab);

				tabID = tab.ID;
				tabName = tab.Name;
				tabType = tab.Type;

				// Properties for table view tab
				tabDBName = tab.DBName || '';
				tabTableDBPoolID = tab.ActiveDBID || '';
				tabConnName = tab.ConnectionName || '';
				tabConnID = tab.ConnectionID || 0;

				select = tab.Select;
				limit = tab.Limit;
				offset = tab.Offset;
				where = tab.Where;
				orderBy = tab.OrderBy;
				groupBy = tab.GroupBy;
				tableColumns = tab.TableColumnsList;

				editor = tab.Editor;

				if ($activeDBs.length == 0) {
					$selectedDBDisplay = 'Connect to a database';
					$currentColor = '';
					$activePoolID = '';
				} else {
					$selectedDBDisplay = tab.ActiveDB || 'Connect to a database';
					$activePoolID = tab.ActiveDBID || '';
					$currentColor = tab.ActiveDBColor || '';
				}

				// Update columns
				if (tab.columns) {
					for (const column of tab.columns) {
						columns.set([
							...$columns,
							{
								accessorKey: column,
								header: column
							}
						]);
					}
				}

				// Update rows
				if (tab.rows) {
					for (const row of tab.rows) {
						let cell: Record<string, any> = {};
						for (const resultCell of row) {
							if (resultCell.column && resultCell.value) {
								cell[resultCell.column] = resultCell.value;
							}
						}
						rows.set([...$rows, cell]);
					}
				}
			})
			.catch((error) => {
				toast.error('Failed to set active tab', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		columns.set([]);
		rows.set([]);
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

	function executeQuery() {
		if ($selectedQuery.trim() == '') {
			console.log('selected query is empty');
			console.log($selectedQuery.trim());
			toast.error('Please select a query to execute', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}
		if ($activePoolID == '') {
			toast.error('Please select a database to execute the query', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		queryLoading = true;
		// Execute query
		ExecuteQuery($activePoolID, $selectedQuery, tabID)
			.then((result) => {
				if (!result.ok) {
					queryLoading = false;
					toast.error('Query Failed', {
						description: result.message,
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
					return;
				}

				// Update columns
				if (result.columns) {
					for (const column of result.columns) {
						columns.set([
							...$columns,
							{
								accessorKey: column,
								header: column
							}
						]);
					}
				}

				// Update rows
				if (result.rows) {
					for (const row of result.rows) {
						let cell: Record<string, any> = {};
						for (const resultCell of row) {
							if (resultCell.column && resultCell.value) {
								cell[resultCell.column] = resultCell.value;
							}
						}
						rows.set([...$rows, cell]);
					}
				}
				queryLoading = false;
			})
			.catch((error) => {
				queryLoading = false;
				// Handle errors from the ExecuteQuery call
				toast.error('Query Failed', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		columns.set([]);
		rows.set([]);
	}

	function getTableData() {
		if (tabTableDBPoolID == '') {
			toast.error('Please select a database to execute the query', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		queryLoading = true;

		// Execute query
		GetTableData(tabTableDBPoolID, tabID, tabName, select, limit, offset, where, orderBy, groupBy)
			.then((result) => {
				if (!result.ok) {
					queryLoading = false;
					toast.error('Query Failed', {
						description: result.message,
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
					return;
				}

				// Update columns
				if (result.columns) {
					for (const column of result.columns) {
						columns.set([
							...$columns,
							{
								accessorKey: column,
								header: column
							}
						]);
					}
				}

				// Update rows
				if (result.rows) {
					for (const row of result.rows) {
						let cell: Record<string, any> = {};
						for (const resultCell of row) {
							if (resultCell.column && resultCell.value) {
								cell[resultCell.column] = resultCell.value;
							}
						}
						rows.set([...$rows, cell]);
					}
				}
				queryLoading = false;
			})
			.catch((error) => {
				queryLoading = false;
				// Handle errors from the ExecuteQuery call
				toast.error('Query Failed', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		columns.set([]);
		rows.set([]);
	}

	document.addEventListener('keydown', handleKeyDown);
	function handleKeyDown(event: KeyboardEvent) {
		if (event.altKey && event.key === 'Enter') {
			event.preventDefault();
			if (tabType == 'table') {
				console.log('get table data');
				getTableData();
			} else {
				console.log('execute query');
				executeQuery();
			}
		}
		// Command+S (Mac) or Ctrl+S (Windows/Linux)
		if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
			event.preventDefault();
			// Your custom logic here
			toast.success('Not Needed! 😂', {
				description: 'Your queries are saved automatically 😂',
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		}
	}

	function selectActiveDB(activeDBDisplay: string, poolID: string, activeDBColor: string) {
		$selectedDBDisplay = activeDBDisplay;
		$activePoolID = poolID;
		$currentColor = activeDBColor;

		SaveActiveDBProps(tabID, $activePoolID, $selectedDBDisplay, $currentColor);
	}

	$effect(() => {
		if ($activeDBs.length == 0) {
			$selectedDBDisplay = 'Connect to a database';
			$currentColor = '';
		}
	});

	// Call UpdateTabEditorContent on editor change
	let editorUpdateTimer: any;
	$effect(() => {
		// Explicitly reference editor to ensure reactivity
		const _ = editor;
		const selectQuery = select;
		const limitQuery = limit;
		const offsetQuery = offset;
		const whereQuery = where;
		const orderByQuery = orderBy;
		const groupByQuery = groupBy;

		// Clear any existing timeout to debounce rapid changes
		if (editorUpdateTimer) clearTimeout(editorUpdateTimer);

		// Set a new timeout to update the content after typing stops
		editorUpdateTimer = setTimeout(() => {
			UpdateTabEditorContent(tabID, editor, select, limit, offset, where, orderBy, groupBy);
		}, 500);
	});

	// Write a function to call addTab when pressed cmd + t
	$effect(() => {
		document.addEventListener('keydown', (event: KeyboardEvent) => {
			if (event.key === 't' && event.metaKey) {
				addTab();
			}
		});
	});
</script>

<div class="my-2 flex h-full flex-1 flex-col rounded-md">
	<Tabs.Root value={tabID.toString()} class="flex h-full flex-1 flex-col overflow-hidden">
		<!-- Tabs visible in the header -->
		<header class="flex h-10 items-center justify-center">
			<div class="flex w-full items-center justify-between">
				<div class="flex h-auto items-center justify-center overflow-x-auto">
					<Sidebar.Trigger />
					<Separator orientation="vertical" />
					<Tabs.List class="thin-scrollbar scrollbar-thin overflow-x-auto overflow-y-hidden">
						{#each Array.from(tabsMap.entries()) as [key, tab]}
							<div class="mr-2 flex rounded-sm bg-slate-800">
								<Tabs.Trigger
									value={tab.ID.toString()}
									class="flex items-center justify-center h-auto"
									onclick={() => setActiveTab(tab.ID)}
								>
									{tab.Name}
								</Tabs.Trigger>
								<button
									class="rounded-r-sm bg-slate-900 px-2 py-1 text-slate-300 hover:text-red-700"
									onclick={() => deleteTab(tab.ID)}
								>
									<X size={16} />
								</button>
							</div>
						{/each}
						{#if tabLoading}
							<Button variant="outline" size="sm">
								<Spinner />
								Processing
							</Button>
						{/if}
						<Button
							variant="outline" size="sm"
							onclick={() => addTab()}
						>
							<Plus size={16} color="white" />
						</Button>
					</Tabs.List>
				</div>
				<!-- <div class="flex">
					<Button variant="secondary" size="sm" onclick={toggleChatPane}>
						<Chat size={16} />
						{#if tabType == 'table'}
							Chat with {tabName}
						{:else}
							Chat with {$selectedDBDisplay}
						{/if}
					</Button>
				</div> -->
			</div>
		</header>

		{#if tabsMap.size > 0}
			<!-- Main Content on screen -->

			{#if tabType == 'table'}
				<div class="flex h-screen flex-1 flex-col">
					<div class="flex h-full flex-1 flex-col justify-center">
						<!-- Breadcrumb -->
						<div class="mt-1 flex items-center justify-between">
							<div class="flex items-center px-2">
								<Breadcrumb.Root>
									<Breadcrumb.List>
										<Breadcrumb.Item>
											<Breadcrumb.Link>{tabConnName}</Breadcrumb.Link>
										</Breadcrumb.Item>
										<Breadcrumb.Separator />
										<Breadcrumb.Item>
											<Breadcrumb.Link>{tabDBName}</Breadcrumb.Link>
										</Breadcrumb.Item>
										<Breadcrumb.Separator />
										<Breadcrumb.Item>
											<Breadcrumb.Page>{tabName}</Breadcrumb.Page>
										</Breadcrumb.Item>
									</Breadcrumb.List>
								</Breadcrumb.Root>
								{#if tabTableDBPoolID === ''}
									<Label class="ml-2 text-red-500">Disconnected</Label>
								{:else}
									<Label class="ml-2 text-green-500">Connected</Label>
								{/if}
							</div>
							<div class="flex px-2">
								<Tabs.Root value={tableViewTab}>
									<Tabs.List class="flex items-center justify-center gap-2">
										<Tabs.Trigger value="data" onclick={() => (tableViewTab = 'data')}
											>Data</Tabs.Trigger
										>
										<Tabs.Trigger value="manage" onclick={() => (tableViewTab = 'manage')}
											>Manage</Tabs.Trigger
										>
									</Tabs.List>
								</Tabs.Root>
							</div>
						</div>

						<!-- Content based on selected tab -->
						{#if tableViewTab === 'data'}
							<div class="flex h-screen flex-1 flex-col">
								<div class="h-18 flex flex-col">
									<Collapsible.Root>
										<div class="flex flex-1 items-center p-1 pt-0">
											<div class="flex flex-1 items-center gap-2 p-1">
												<Label for="where">Where</Label>
												<DropdownMenu.Root bind:open={isWhereDropdownOpen}>
													<DropdownMenu.Trigger
														disabled={true}
														class="flex flex-1 items-center gap-2 p-1"
													>
														<Input
															type="text"
															id="where"
															placeholder="Where..."
															class="w-full focus-visible:ring-0"
															bind:value={where}
															onkeyup={(e) => {
																if (e.key === "'" || e.key === '"') {
																	where += e.key;
																}
																if (e.key === '(') {
																	where += ')';
																}
															}}
															onkeydown={(e) => {
																if (e.key === 'Enter' && !e.altKey) {
																	isWhereDropdownOpen = true;
																}
																if (e.key === 'Escape') {
																	isWhereDropdownOpen = false;
																}
															}}
														/>
													</DropdownMenu.Trigger>
													<DropdownMenu.Content align="start" class="w-96 overflow-auto">
														<DropdownMenu.Group class="max-h-96 overflow-auto">
															<DropdownMenu.Label>Columns</DropdownMenu.Label>
															<DropdownMenu.Separator />

															{#each tableColumns as column (column)}
																<DropdownMenu.Item
																	onclick={() => {
																		if (
																			where.length > 0 &&
																			!where.endsWith(',') &&
																			!where.endsWith(', ')
																		) {
																			where += ', ';
																		}
																		where += column;
																		isWhereDropdownOpen = false;
																	}}>{column}</DropdownMenu.Item
																>
															{/each}
														</DropdownMenu.Group>
													</DropdownMenu.Content>
												</DropdownMenu.Root>
											</div>
											<Collapsible.Trigger>
												<Button size="sm" variant="secondary">Advanced</Button>
											</Collapsible.Trigger>
										</div>
										<Collapsible.Content>
											<div class="flex flex-1 items-center gap-2 p-1">
												<Label for="select">Select</Label>
												<DropdownMenu.Root bind:open={isSelectDropdownOpen}>
													<DropdownMenu.Trigger
														disabled={true}
														class="flex flex-1 items-center gap-2 p-1"
													>
														<Input
															type="text"
															id="select"
															placeholder="Select..."
															class="w-full focus-visible:ring-0"
															bind:value={select}
															onkeyup={(e) => {
																if (e.key === "'" || e.key === '"') {
																	select += e.key;
																}
																if (e.key === '(') {
																	select += ')';
																}
															}}
															onkeydown={(e) => {
																if (e.key === 'Enter' && !e.altKey) {
																	isSelectDropdownOpen = true;
																}
																if (e.key === 'Escape') {
																	isSelectDropdownOpen = false;
																}
															}}
														/>
													</DropdownMenu.Trigger>
													<DropdownMenu.Content align="start" class="w-96 overflow-auto">
														<DropdownMenu.Group class="max-h-96 overflow-auto">
															<DropdownMenu.Label>Columns</DropdownMenu.Label>
															<DropdownMenu.Separator />

															{#each tableColumns as column (column)}
																<DropdownMenu.Item
																	onclick={() => {
																		if (
																			select.length > 0 &&
																			!select.endsWith(',') &&
																			!select.endsWith(', ')
																		) {
																			select += ', ';
																		}
																		select += column;
																		isSelectDropdownOpen = false;
																	}}>{column}</DropdownMenu.Item
																>
															{/each}
														</DropdownMenu.Group>
													</DropdownMenu.Content>
												</DropdownMenu.Root>
											</div>
											<div class="flex flex-1 items-center p-1 pt-0">
												<div class="flex flex-1 items-center gap-2 p-1">
													<Label for="orderBy">Order</Label>
													<DropdownMenu.Root bind:open={isOrderByDropdownOpen}>
														<DropdownMenu.Trigger
															disabled={true}
															class="flex flex-1 items-center gap-2 p-1"
														>
															<Input
																type="text"
																id="orderBy"
																placeholder="Order By"
																class="w-full focus-visible:ring-0"
																bind:value={orderBy}
																onkeyup={(e) => {
																	if (e.key === "'" || e.key === '"') {
																		orderBy += e.key;
																	}
																	if (e.key === '(') {
																		orderBy += ')';
																	}
																}}
																onkeydown={(e) => {
																	if (e.key === 'Enter' && !e.altKey) {
																		isOrderByDropdownOpen = true;
																	}
																	if (e.key === 'Escape') {
																		isOrderByDropdownOpen = false;
																	}
																}}
															/>
														</DropdownMenu.Trigger>
														<DropdownMenu.Content align="start" class="w-96 overflow-auto">
															<DropdownMenu.Group class="max-h-96 overflow-auto">
																<DropdownMenu.Label>Columns</DropdownMenu.Label>
																<DropdownMenu.Separator />

																{#each tableColumns as column (column)}
																	<DropdownMenu.Item
																		onclick={() => {
																			if (
																				orderBy.length > 0 &&
																				!orderBy.endsWith(',') &&
																				!orderBy.endsWith(', ')
																			) {
																				orderBy += ', ';
																			}
																			orderBy += column;
																			isOrderByDropdownOpen = false;
																		}}>{column}</DropdownMenu.Item
																	>
																{/each}
															</DropdownMenu.Group>
														</DropdownMenu.Content>
													</DropdownMenu.Root>
												</div>
												<div class="flex flex-1 items-center p-1">
													<Label for="groupBy">Group</Label>
													<DropdownMenu.Root bind:open={isGroupByDropdownOpen}>
														<DropdownMenu.Trigger
															disabled={true}
															class="flex flex-1 items-center p-1"
														>
															<Input
																type="text"
																id="groupBy"
																placeholder="Group By"
																class="w-full focus-visible:ring-0"
																bind:value={groupBy}
																onkeyup={(e) => {
																	if (e.key === "'" || e.key === '"') {
																		groupBy += e.key;
																	}
																	if (e.key === '(') {
																		groupBy += ')';
																	}
																}}
																onkeydown={(e) => {
																	if (e.key === 'Enter' && !e.altKey) {
																		isGroupByDropdownOpen = true;
																	}
																	if (e.key === 'Escape') {
																		isGroupByDropdownOpen = false;
																	}
																}}
															/>
														</DropdownMenu.Trigger>
														<DropdownMenu.Content align="start" class="w-96 overflow-auto">
															<DropdownMenu.Group class="max-h-96 overflow-auto">
																<DropdownMenu.Label>Columns</DropdownMenu.Label>
																<DropdownMenu.Separator />

																{#each tableColumns as column (column)}
																	<DropdownMenu.Item
																		onclick={() => {
																			if (
																				groupBy.length > 0 &&
																				!groupBy.endsWith(',') &&
																				!groupBy.endsWith(', ')
																			) {
																				groupBy += ', ';
																			}
																			groupBy += column;
																			isGroupByDropdownOpen = false;
																		}}>{column}</DropdownMenu.Item
																	>
																{/each}
															</DropdownMenu.Group>
														</DropdownMenu.Content>
													</DropdownMenu.Root>
												</div>
												<div class="flex items-center gap-2 p-1">
													<Label for="limit">Limit</Label>
													<Input
														type="text"
														id="limit"
														placeholder="Limit"
														class="w-24 focus-visible:ring-0"
														bind:value={limit}
													/>
												</div>
												<div class="flex items-center gap-2 p-1">
													<Label for="offset">Offset</Label>
													<Input
														type="text"
														id="offset"
														placeholder="Offset"
														class="w-24 focus-visible:ring-0"
														bind:value={offset}
													/>
												</div>
											</div>
										</Collapsible.Content>
									</Collapsible.Root>
								</div>
								<div class="flex flex-1 overflow-auto border">
									<div class="h-full w-full overflow-auto">
										{#if $columns.length > 0}
											<DataTable data={$rows} columns={$columns} />
										{:else if queryLoading}
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
										{/if}
									</div>
								</div>
							</div>
						{:else if tableViewTab === 'manage'}
							<div class="mt-2 flex flex-1 flex-col overflow-hidden">
								<ManageTable activePoolID={$activePoolID} {tabName} />
							</div>
						{/if}
					</div>
				</div>
			{:else}
				<div class="flex h-screen flex-1 flex-col rounded-md">
					<Tabs.Content value={tabID.toString()} class="flex flex-col flex-1 overflow-hidden">
						<div class="flex h-full flex-col">
							<!-- Active DB Selector and Execute Query Button -->
							<div class="flex items-center justify-between h-10 mb-2">
								<Select.Root type="single" name="activeDatabase">
									<Select.Trigger
										class="{getColorClass(
											$currentColor
										)} prevent:default w-auto bg-opacity-20 hover:bg-opacity-25 mt-0.5 ml-0.5"
									>
										{$selectedDBDisplay}
									</Select.Trigger>
									<Select.Content>
										<Select.Group>
											{#each $activeDBs as activeDB}
												<Select.Item
													onclick={() =>
														selectActiveDB(
															activeDB.ConnectionName + ' - ' + activeDB.Name,
															activeDB.PoolID,
															activeDB.Color
														)}
													class="{getColorClass(activeDB.Color)} bg-opacity-20 hover:bg-opacity-25"
													value={activeDB.ID}
													label={activeDB.Name}
													>{activeDB.ConnectionName} - {activeDB.Name}</Select.Item
												>
											{/each}
										</Select.Group>
									</Select.Content>
								</Select.Root>
							</div>

							<!-- Resizable Panes for Editor and Output -->
							<Resizable.ResizablePaneGroup direction="vertical" class="h-full">
								<!-- SQL Editor Pane -->
								<Resizable.Pane
									defaultSize={editorHeight}
									minSize={10}
									class="rsz-pane"
								>
									<SqlEditor
										bind:value={editor}
										bind:selectedQuery={$selectedQuery}
										suggestions={Array.from($suggestions)}
									/>
								</Resizable.Pane>

								<Resizable.ResizableHandle withHandle />

								<!-- Output Pane -->
								<Resizable.Pane
									defaultSize={outputHeight}
									minSize={10}
									class="rsz-pane"
								>
									<div class="h-full">
										{#if $columns.length > 0}
											<DataTable data={$rows} columns={$columns} />
										{:else if queryLoading}
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
											<Skeleton class="my-3 h-[40px] w-full" />
										{/if}
									</div>
								</Resizable.Pane>
							</Resizable.ResizablePaneGroup>
						</div>
					</Tabs.Content>
				</div>
			{/if}
		{/if}
	</Tabs.Root>
</div>

<style>
	:global(.thin-scrollbar) {
		/* Firefox */
		scrollbar-width: thin;
		scrollbar-color: #4a5568 #2d3748;
	}

	:global(.thin-scrollbar::-webkit-scrollbar) {
		height: 1px; /* Horizontal scrollbar height */
		width: 1px; /* Vertical scrollbar width */
	}

	:global(.thin-scrollbar::-webkit-scrollbar-track) {
		background: #2d3748;
		border-radius: 1px;
	}

	:global(.thin-scrollbar::-webkit-scrollbar-thumb) {
		background: #4a5568;
		border-radius: 1px;
	}

	:global(.thin-scrollbar::-webkit-scrollbar-thumb:hover) {
		background: #718096;
	}
</style>
