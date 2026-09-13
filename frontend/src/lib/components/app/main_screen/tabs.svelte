<script lang="ts">
	import * as Tabs from '$lib/components/ui/tabs/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import X from 'lucide-svelte/icons/x';
	import Plus from 'lucide-svelte/icons/plus';
	import Chat from 'lucide-svelte/icons/message-circle-more';
	import * as Resizable from '$lib/components/ui/resizable/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { onMount } from 'svelte';
	import { LaserLoader } from '$lib/components/ui/laser-loader/index.js';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';


	// Import our custom components
	import SqlEditor from '$lib/components/app/main_screen/sql_editor.svelte';
	import ClauseBar from '$lib/components/app/main_screen/clause-bar.svelte';
	import {
		AddTab,
		DeleteTab,
		UpdateTabEditorContent,
		GetAllTabs
	} from '$lib/wailsjs/go/app/Tabs';
	import { suggestions } from '$lib/state.svelte';
	import {
		tabsMap,
		selectedQuery
	} from '$lib/state.svelte';

	import { toast } from 'svelte-sonner';
	import { CancelQuery, ExecuteQuery, GetTableData } from '$lib/wailsjs/go/app/Connections.js';

	import { columnTypeMeta } from '$lib/components/ui/data-table/index.js';
	import DataTable from './data-table.svelte';
	import { columns, rows, totalRows, currentPage, currentPageSize } from '$lib/state.svelte';
	import ManageTable from './manage_table.svelte';
	import DataTableManual from './data-table-manual.svelte';
	import { Play, Square } from 'lucide-svelte';

	let editorHeight = $state(50); // Percentage of the container height
	let outputHeight = $state(50); // Percentage of the container height

	let queryLoading = $state(false);
	let tabLoading = $state(false);

	// Ask the backend to stop whatever this tab is running. The in-flight call
	// resolves on its own with a cancelled result, which is what clears the
	// running state, so there is nothing to unwind here.
	function cancelQuery() {
		const currentTabID = tabID;
		CancelQuery(currentTabID).catch((error) => {
			toast.error('Could not stop the query', { description: String(error) });
		});
	}

	// A cancelled query is something the user asked for, not a failure worth an
	// error toast.
	function reportQueryOutcome(message: string) {
		if (message?.toLowerCase().startsWith('query cancelled')) {
			toast.info(message);
			return;
		}
		toast.error('Query Failed', {
			description: message,
			action: {
				label: 'OK',
				onClick: () => console.info('OK')
			}
		});
	}

	// Drag-and-drop tab reordering state
	let dragTabId: number | null = $state(null);
	let dragOverTabId: number | null = $state(null);
	// Ordered tab IDs for rendering (controls display order)
	let tabOrder: number[] = $state([]);

	function onTabDragStart(e: DragEvent, id: number) {
		dragTabId = id;
		if (e.dataTransfer) {
			e.dataTransfer.effectAllowed = 'move';
		}
	}

	function getColorClass(color: string): string {
		const colorMap: Record<string, string> = {
			'bg-purple-500': 'border-t dark:border-purple-700 light:border-purple-500',
			'bg-indigo-500': 'border-t dark:border-indigo-700 light:border-indigo-500',
			'bg-emerald-500': 'border-t dark:border-emerald-700 light:border-emerald-500',
			'bg-red-500': 'border-t dark:border-red-700 light:border-red-500',
			'bg-blue-500': 'border-t dark:border-blue-700 light:border-blue-500',
			'bg-green-500': 'border-t dark:border-green-700 light:border-green-500',
			'bg-yellow-500': 'border-t dark:border-yellow-700 light:border-yellow-500',
			'bg-orange-500': 'border-t dark:border-orange-700 light:border-orange-500',
			'bg-pink-500': 'border-t dark:border-pink-700 light:border-pink5900'
		};
		return colorMap[color] || '';
	}

	function onTabDragOver(e: DragEvent, id: number) {
		e.preventDefault();
		if (dragTabId === null || dragTabId === id) return;
		dragOverTabId = id;

		// Reorder tabs in real-time
		const fromIndex = tabOrder.indexOf(dragTabId);
		const toIndex = tabOrder.indexOf(id);
		if (fromIndex === -1 || toIndex === -1 || fromIndex === toIndex) return;
		const newOrder = [...tabOrder];
		newOrder.splice(fromIndex, 1);
		newOrder.splice(toIndex, 0, dragTabId);
		tabOrder = newOrder;
	}

	function onTabDragEnd() {
		dragTabId = null;
		dragOverTabId = null;
	}

	// The tab strip iterates over the tabs themselves, not over their ids. tabOrder
	// and tabsMap are updated one after the other, so an id in the order can spend a
	// moment with no tab behind it; resolving here means the strip simply renders one
	// tab fewer instead of rendering a row for a tab that is already gone.
	const openTabs = $derived(
		tabOrder
			.map((id) => tabsMap.get(id))
			.filter((tab) => tab !== undefined)
	);

	// Keep tabOrder in sync with tabsMap additions/removals
	function syncTabOrder() {
		const mapKeys = Array.from(tabsMap.keys());
		// Add any new keys not yet in the order
		for (const key of mapKeys) {
			if (!tabOrder.includes(key)) {
				tabOrder = [...tabOrder, key];
			}
		}
		// Remove keys that no longer exist
		tabOrder = tabOrder.filter((id) => tabsMap.has(id));
	}

	// Handle Tabs

	// Active tab properties
	let {
		tabID = $bindable(0),
		tabName = $bindable(''),
		tabType = $bindable(''),
		tabDBName = $bindable(''),
		tabDBPoolID = $bindable(''),
		tabConnName = $bindable(''),
		tabConnID = $bindable(0),
		select = $bindable(''),
		limit = $bindable(''),
		offset = $bindable(''),
		where = $bindable(''),
		orderBy = $bindable(''),
		groupBy = $bindable(''),
		tableColumns = $bindable([]),
		aiChat = $bindable([]),
		chatPaneCollapsed = $bindable(false),
		toggleChatPane
	} = $props();

	// Read the running flag off the tab itself rather than a component-local one,
	// so switching tabs shows each tab's own state instead of whatever the last
	// query left behind. Declared after the props because it reads tabID.
	const activeTabRunning = $derived(tabsMap.get(tabID)?.IsQueryRunning ?? false);

	// tabsMap is a SvelteMap, and SvelteMap only notifies its readers when the
	// value stored against a key changes identity. Tabs are read out of the map,
	// mutated in place and put straight back, so the map is handed the very object
	// it already holds and nothing re-renders. Storing a fresh object with the same
	// prototype is what makes those mutations visible: without it a finished query
	// leaves its spinner spinning and its stop button up until something else
	// happens to invalidate the read.
	function commitTab(id: number, tab: NonNullable<ReturnType<typeof tabsMap.get>>) {
		tabsMap.set(id, Object.assign(Object.create(Object.getPrototypeOf(tab)), tab));
	}

	let editor = $state('');



	// Table view tab state (for Data/Structure/Indexes)
	let tableViewTab = $state('data');

	onMount(() => {
		getAllTabs() 
	});


	function getAllTabs() {
		tableViewTab = 'data';

		if (tabsMap.size > 0) {
			// For each tab in tabs map
			for (const tab of tabsMap.values()) {

				// Set active tab properties
				if (tab.IsActive) {
					if (tab.IsQueryRunning) {
						queryLoading = true;
					} else {
						queryLoading = false;
					}
					tabID = tab.ID;
					tabName = tab.Name;
					tabType = tab.Type;
					tabDBName = tab.DBName || '';
					tabDBPoolID = tab.ActiveDBID || '';
					tabConnName = tab.ConnectionName || '';
					tabConnID = tab.ConnectionID || 0;

					select = tab.Select;
					limit = tab.Limit;
					offset = tab.Offset;
					where = tab.Where;
					orderBy = tab.OrderBy;
					groupBy = tab.GroupBy;
					tableColumns = tab.TableColumnsList ?? [];
					aiChat = tab.AIChat || [];

					editor = tab.Editor;

					// Update columns
					if (tab.columns) {
						columns.set(tab.columns.map((column, index) => ({
							accessorKey: column,
							id: String(index),
							header: column,
							meta: columnTypeMeta(tab.columnTypes, index)
						})));
					}

					// Process rows once and cache them
					if (tab.rows) {
						let processedRows: any[] = [];
						for (const row of tab.rows) {
							let cell: Record<string, any> = {};
							if (Array.isArray(row)) {
								for (const resultCell of row) {
									if (resultCell.column && resultCell.value) {
										cell[resultCell.column] = resultCell.value;
									}
								}
								processedRows.push(cell);
							}
						}
						// Cache the processed rows in the tab object
						// We need to cast to any to avoid TS error if model definition isn't updated instantly in IDE
						(tab as any).processedRows = processedRows;
						rows.set(processedRows);
					}
				}
			}
			syncTabOrder();
			
		} else {
			GetAllTabs().then((tabs) => {
				if (!tabs) {
					return;
				}
				for (const tab of tabs) {
					tabsMap.set(tab.ID, tab);

					// Set active tab properties
					if (tab.IsActive) {
						if (tab.IsQueryRunning) {
							queryLoading = true;
						} else {
							queryLoading = false;
						}
						tabID = tab.ID;
						tabName = tab.Name;
						tabType = tab.Type;
						tabDBName = tab.DBName || '';
						tabDBPoolID = tab.ActiveDBID || '';
						tabConnName = tab.ConnectionName || '';
						tabConnID = tab.ConnectionID || 0;

						select = tab.Select;
						limit = tab.Limit;
						offset = tab.Offset;
						where = tab.Where;
						orderBy = tab.OrderBy;
						groupBy = tab.GroupBy;
						tableColumns = tab.TableColumnsList ?? [];
						aiChat = tab.AIChat || [];

						editor = tab.Editor;

						// Update columns
						// if (tab.columns) {
						// 	for (const column of tab.columns) {
						// 		columns.set([
						// 			...$columns,
						// 			{
						// 				accessorKey: column,
						// 				header: column
						// 			}
						// 		]);
						// 	}
						// }

						// Process rows once and cache them
						// if (tab.rows) {
						// 	let processedRows: any[] = [];
						// 	for (const row of tab.rows) {
						// 		let cell: Record<string, any> = {};
						// 		if (Array.isArray(row)) {
						// 			for (const resultCell of row) {
						// 				if (resultCell.column && resultCell.value) {
						// 					cell[resultCell.column] = resultCell.value;
						// 				}
						// 			}
						// 			processedRows.push(cell);
						// 		}
						// 	}
						// 	// Cache the processed rows in the tab object
						// 	// We need to cast to any to avoid TS error if model definition isn't updated instantly in IDE
						// 	(tab as any).processedRows = processedRows;
						// 	tabsMap.set(tab.ID, tab);
						// }

						// // Update active rows from cache
						// if ((tab as any).processedRows) {
						// 	rows.set((tab as any).processedRows);
						// }
					}
				}
				syncTabOrder();
			});
		}

		columns.set([]);
		rows.set([]);
	}

	export function addTab(
		connID: number,
		activeDBPoolID: string,
		dbName: string,
		addTabType: string,
		tableName: string,
		editorFromTable: boolean
	) {
		// Clear the output area for new tab. This is important to avoid confusion as the new tab loads and fetches its own data
		columns.set([]);
		rows.set([]);

		$selectedQuery = '';

		tableViewTab = 'data';

		tabLoading = true;

		AddTab(
			connID,
			activeDBPoolID,
			dbName,
			addTabType,
			tableName,
			editorFromTable
		)
			.then((tab) => {
				queryLoading = false;
				tabsMap.set(tab.ID, tab);
				syncTabOrder();
				tabID = tab.ID;
				tabName = tab.Name;
				tabType = tab.Type;
				tabDBName = tab.DBName || '';
				tabDBPoolID = tab.ActiveDBID || '';
				tabConnName = tab.ConnectionName || '';
				tabConnID = tab.ConnectionID || 0;

				select = tab.Select;
				limit = tab.Limit;
				offset = tab.Offset;
				where = tab.Where;
				orderBy = tab.OrderBy;
				groupBy = tab.GroupBy;
				tableColumns = tab.TableColumnsList ?? [];
				aiChat = tab.AIChat || [];

				editor = tab.Editor;
				tabLoading = false;

				if (tabType === 'table') {
					getTableData();
				}
				if (editorFromTable) {
					$selectedQuery = editor
					executeQuery();
				}
			})
			.catch((error) => {
				toast.error('Failed to add tab', {
					description: String(error),
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
				tabLoading = false;
			});
	}

	function deleteTab(id: number) {
		// Check if the tab being deleted has query running and if yes, prevent deletion and show toast
		const tabToDelete = tabsMap.get(id);
		if (tabToDelete && tabToDelete.IsQueryRunning) {
			toast.error('Cannot delete tab', {
				description: 'A query is currently running in this tab. Please wait for it to finish before deleting.',
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		$selectedQuery = '';

		let wasActive = (id === tabID);

		// Delete the old tab from the map
		tabsMap.delete(id);
		syncTabOrder();

		// If the tab was active, switch to another tab
		if (wasActive) {
			if (tabsMap.size > 0) {
				// Get the first available tab
				const firstTabID = tabsMap.keys().next().value;
				if (firstTabID) {
					setActiveTab(firstTabID);
				}
			} else {
				// No tabs left
				tabID = 0;
				tabName = '';
				editor = '';
				aiChat = [];
				columns.set([]);
				rows.set([]);
			}
		} else {
			// If not active, just ensure UI state is clean if needed, 
			// but we didn't change the active view so do nothing.
		}

		DeleteTab(id)
			.then(() => {
				// Backend delete successful
			})
			.catch((error) => {
				toast.error('Failed to delete tab', {
					description: String(error),
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	}

	function setActiveTab(id: number) {
		$selectedQuery = '';
		tableViewTab = 'data';

		// Set the active tab from the map
		const tab = tabsMap.get(id);
		if (!tab) {
			return;
		}

		if (tab.IsQueryRunning) {
			queryLoading = true;
		} else {
			queryLoading = false;
		}

		tab.IsActive = true;

		tabID = tab.ID;
		tabName = tab.Name;
		tabType = tab.Type;

		// Properties for table view tab
		tabDBName = tab.DBName || '';
		tabDBPoolID = tab.ActiveDBID || '';
		tabConnName = tab.ConnectionName || '';
		tabConnID = tab.ConnectionID || 0;

		select = tab.Select;
		limit = tab.Limit;
		offset = tab.Offset;
		where = tab.Where;
		orderBy = tab.OrderBy;
		groupBy = tab.GroupBy;
		tableColumns = tab.TableColumnsList ?? [];
		aiChat = tab.AIChat || [];

		editor = tab.Editor;

		totalRows.set(tab.totalRows);
		currentPage.set(tab.currentPage);
		currentPageSize.set(Number(tab.Limit));

		// Update columns
		columns.set([]);
		if (tab.columns) {
			columns.set(tab.columns.map((column, index) => ({
				accessorKey: column,
				id: String(index),
				header: column,
				meta: columnTypeMeta(tab.columnTypes, index)
			})));
		}

		// Update rows using cached data (O(1))
		rows.set([]);
		if ((tab as any).processedRows && (tab as any).processedRows.length > 0) {
			rows.set((tab as any).processedRows);
		} else if (tab.rows) {
			// Fallback: If not cached yet (legacy/first load edge case), process and cache now
			let newRows: any[] = [];
			for (const row of tab.rows) {
				let cell: Record<string, any> = {};
				if (Array.isArray(row)) {
					for (const resultCell of row) {
						if (resultCell.column && resultCell.value) {
							cell[resultCell.column] = resultCell.value;
						}
					}
					newRows.push(cell);
				}
			}
			(tab as any).processedRows = newRows;
			rows.set(newRows);
		}

		lastQueryExecutionTime = tab.LastQueryExecutionTime

		tabsMap.set(tabID, tab);

		// SetActiveTab(id)
		// 	.then((tab) => {
		// 		// We don't update UI from this response anymore as it's slow/empty
		// 		// Maybe update metadata if needed, but tabsMap already has it from GetAllTabs or local updates
		// 	})
		// 	.catch((error) => {
		// 		toast.error('Failed to set active tab', {
		// 			description: String(error),
		// 			action: {
		// 				label: 'OK',
		// 				onClick: () => console.info('OK')
		// 			}
		// 		});
		// 	});
	}

	let executeQueryTableName = $state("")
	let lastQueryExecutionTime = $state(0)

	function executeQuery(isExplain?: boolean) {
		if (tabType == 'table') {
			return;
		}
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

		// Set the current output to empty
		columns.set([]);
		rows.set([]);

		// Save in a const to prevent reactive tabID from changing in the middle of execution and causing issues with tabsMap updates
		const currentTabID = tabID;

		// Update loading state in the current tab to show spinner on the tab itself
		let currentTab = tabsMap.get(currentTabID);
		if (currentTab) {
			currentTab.IsQueryRunning = true;
			// Delete the previous output data
			currentTab.columns = [];
			currentTab.columnTypes = [];
			currentTab.rows = [];
			(currentTab as any).processedRows = []; // Clear cached rows as well
			commitTab(currentTabID, currentTab);
		}

		queryLoading = true;
		let explain = isExplain || false;
		// Execute query
		ExecuteQuery(tabID, $selectedQuery, explain)
			.then((result) => {
				let currentTab = tabsMap.get(currentTabID);

				// Update the map with cached rows
				if (!result.ok) {
					queryLoading = false;

					// Update the tab state in memory
					if (currentTab) {
						currentTab.IsQueryRunning = false;
						commitTab(currentTabID, currentTab);
					}

					reportQueryOutcome(result.message);
					return;
				}

				// If current tab is still the tab for which the query was run
				if (currentTabID == tabID) {
					// Update columns
					if (result.columns) {
						columns.set(result.columns.map((column, index) => ({
							accessorKey: column,
							id: String(index),
							header: column,
							meta: columnTypeMeta(result.columnTypes, index)
						})));
					}

					// Update rows. Also update the in-memory tabsMap!
					if (result.rows) {
						let newRows: any[] = [];
						for (const row of result.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							newRows.push(cell);
						}
						rows.set(newRows);
					}

					lastQueryExecutionTime = result.executionTime
				}
				
				if (currentTab) {
					currentTab.columns = result.columns;
					currentTab.columnTypes = result.columnTypes;
					currentTab.rows = result.rows; // result.rows is Cell[][]
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = result.executionTime || 0;
					commitTab(currentTabID, currentTab);
				}
				queryLoading = false;
				executeQueryTableName = result.tableName;

				// Show warning toast if partial results were returned
				if (result.message) {
					toast.warning('Partial Results', {
						description: result.message,
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
				}
			})
			.catch((error) => {
				queryLoading = false;

				// Update the tab state in memory
				let currentTab = tabsMap.get(currentTabID);
				if (currentTab) {
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = 0
					commitTab(currentTabID, currentTab);
				}

				// Handle errors from the ExecuteQuery call
				toast.error('Query Failed', {
					description: String(error),
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
		if (tabType == 'editor') {
			return;
		}

		if (tabDBPoolID == '') {
			toast.error('Please select a database to execute the query', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		queryLoading = true;

		// Set the current output to empty
		columns.set([]);
		rows.set([]);

		// Save in a const to prevent reactive tabID from changing in the middle of execution and causing issues with tabsMap updates
		const currentTabID = tabID;

		// Update loading state in the current tab to show spinner on the tab itself
		let currentTab = tabsMap.get(currentTabID);
		if (currentTab) {
			currentTab.IsQueryRunning = true;
			// Delete the previous output data
			currentTab.columns = [];
			currentTab.columnTypes = [];
			currentTab.rows = [];
			(currentTab as any).processedRows = []; // Clear cached rows as well
			commitTab(currentTabID, currentTab);
		}

		// Execute query
		GetTableData(tabID, tabName, select, limit, offset, where, orderBy, groupBy, false)
			.then((result) => {
				let currentTab = tabsMap.get(currentTabID);

				if (!result.ok) {
					queryLoading = false;

					// Update the tab state in memory
					if (currentTab) {
						currentTab.IsQueryRunning = false;
						commitTab(currentTabID, currentTab);
					}

					reportQueryOutcome(result.message);
					return;
				}

				// If current tab is still the tab for which the query was run
				if (currentTabID == tabID) {
					totalRows.set(result.totalRows);
					currentPage.set(0);
					currentPageSize.set(20);

					// Update columns
					if (result.columns) {
						columns.set(result.columns.map((column, index) => ({
							accessorKey: column,
							id: String(index),
							header: column,
							meta: columnTypeMeta(result.columnTypes, index),
							size: 270
						})));
					}

					// Update rows and map
					if (result.rows) {
						let newRows: any[] = [];
						for (const row of result.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							newRows.push(cell);
						}
						rows.set(newRows);
					}

					lastQueryExecutionTime = result.executionTime
				}
				// Update the map with cached rows
				if (currentTab) {
					currentTab.columns = result.columns;
					currentTab.columnTypes = result.columnTypes;
					currentTab.rows = result.rows; // result.rows is Cell[][]
					currentTab.totalRows = result.totalRows;
					currentTab.Limit = '20';
					currentTab.currentPage = 0;
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = result.executionTime || 0
					commitTab(currentTabID, currentTab);
				}
				queryLoading = false;
			})
			.catch((error) => {
				queryLoading = false;

				// Update the tab state in memory
				let currentTab = tabsMap.get(currentTabID);
				if (currentTab) {
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = 0
					commitTab(currentTabID, currentTab);
				}

				// Handle errors from the ExecuteQuery call
				toast.error('Query Failed', {
					description: String(error),
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		// This is not required to do in server side pagination
		// columns.set([]);
		// rows.set([]);

	}

	function getTablePageData(limit: string, offset: string) {
		if (tabDBPoolID == '') {
			toast.error('Please select a database to execute the query', {
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		queryLoading = true;

		// Set the current output to empty
		// columns.set([]);
		// rows.set([]);

		// Save in a const to prevent reactive tabID from changing in the middle of execution and causing issues with tabsMap updates
		const currentTabID = tabID;

		// Update loading state in the current tab to show spinner on the tab itself
		let currentTab = tabsMap.get(currentTabID);
		if (currentTab) {
			currentTab.IsQueryRunning = true;
			// Delete the previous output data
			currentTab.columns = [];
			currentTab.columnTypes = [];
			currentTab.rows = [];
			(currentTab as any).processedRows = []; // Clear cached rows as well
			commitTab(currentTabID, currentTab);
		}

		// Execute query
		GetTableData(tabID, tabName, select, limit, offset, where, orderBy, groupBy, true)
			.then((result) => {
				let currentTab = tabsMap.get(currentTabID);

				if (!result.ok) {
					queryLoading = false;

					// Update the tab state in memory
					if (currentTab) {
						currentTab.IsQueryRunning = false;
						commitTab(currentTabID, currentTab);
					}

					reportQueryOutcome(result.message);
					return;
				}

				// If current tab is still the tab for which the query was run
				if (currentTabID == tabID) {
					// Update columns
					if (result.columns) {
						columns.set(result.columns.map((column, index) => ({
							accessorKey: column,
							id: String(index),
							header: column,
							meta: columnTypeMeta(result.columnTypes, index)
						})));
					}

					// Update rows and map
					if (result.rows) {
						let newRows: any[] = [];
						for (const row of result.rows) {
							let cell: Record<string, any> = {};
							for (const resultCell of row) {
								if (resultCell.column && resultCell.value) {
									cell[resultCell.column] = resultCell.value;
								}
							}
							newRows.push(cell);
						}
						rows.set(newRows);
					}

					lastQueryExecutionTime = result.executionTime
				}

				// Update the map with cached rows
				if (currentTab) {
					currentTab.columns = result.columns;
					currentTab.columnTypes = result.columnTypes;
					currentTab.rows = result.rows; // result.rows is Cell[][]
					currentTab.Limit = limit;
					currentTab.currentPage = $currentPage;
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = result.executionTime || 0
					commitTab(currentTabID, currentTab);
				}
				queryLoading = false;
			})
			.catch((error) => {
				queryLoading = false;

				// Update the tab state in memory
				let currentTab = tabsMap.get(currentTabID);
				if (currentTab) {
					currentTab.IsQueryRunning = false;
					currentTab.LastQueryExecutionTime = 0
					commitTab(currentTabID, currentTab);
				}

				// Handle errors from the ExecuteQuery call
				toast.error('Query Failed', {
					description: String(error),
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});

		// This is not required to do in server side pagination
		// columns.set([]);
		// rows.set([]);

	}

	// Call UpdateTabEditorContent on editor change
	let editorUpdateTimer: any;
	$effect(() => {
		// Explicitly reference editor to ensure reactivity
		const currentEditorContent = editor;
		const selectQuery = select;
		const limitQuery = limit;
		const offsetQuery = offset;
		const whereQuery = where;
		const orderByQuery = orderBy;
		const groupByQuery = groupBy;

		// Update local map instantly
		let currentTab = tabsMap.get(tabID);
		if (currentTab) {
			currentTab.Editor = currentEditorContent;
			currentTab.Select = selectQuery;
			currentTab.Limit = limitQuery;
			currentTab.Offset = offsetQuery;
			currentTab.Where = whereQuery;
			currentTab.OrderBy = orderByQuery;
			currentTab.GroupBy = groupByQuery;
			// Update other properties if needed, though they seem bound to local state anyway
			tabsMap.set(tabID, currentTab);
		}

		// Clear any existing timeout to debounce rapid changes
		if (editorUpdateTimer) clearTimeout(editorUpdateTimer);

		// Set a new timeout to update the content after typing stops
		editorUpdateTimer = setTimeout(() => {
			UpdateTabEditorContent(tabID, currentEditorContent, selectQuery, limitQuery, offsetQuery, whereQuery, orderByQuery, groupByQuery);
		}, 500);
	});

	function handleKeyDown(event: KeyboardEvent) {
        // Alt + Enter
        if (event.altKey && event.key === 'Enter') {
			if (tabType == 'editor') {
				return;
			}
            event.preventDefault();
			console.log('get table data');
			getTableData();
        }
    }
</script>

<svelte:document onkeydown={handleKeyDown} />

<!-- One control with two states: the same spot runs the query and stops it, so a
     running query always has its stop button where the run button just was. -->
{#snippet runControl(run: () => void, runLabel: string)}
	{#if activeTabRunning}
		<button
			class="mx-2 flex items-center self-center rounded-full border border-red-500 p-1 text-red-500 transition-colors hover:bg-red-500/10 hover:text-red-600"
			onclick={cancelQuery}
			title="Stop the running query"
			aria-label="Stop the running query"
		>
			<Square size={16} />
		</button>
	{:else}
		<button
			class="mx-2 flex items-center self-center rounded-full border border-green-500 p-1 text-green-500 transition-colors hover:bg-green-500/10 hover:text-green-600"
			onclick={run}
			title={runLabel}
			aria-label={runLabel}
		>
			<Play size={16} />
		</button>
	{/if}
{/snippet}

<div class="flex h-full flex-1 flex-col rounded-md bg-background">
	<Tabs.Root value={tabID.toString()} class="flex h-full flex-1 flex-col overflow-hidden">
		<!-- Tabs visible in the header - Chrome style -->
		<header class="flex h-11 items-end bg-background pt-1">
			<div class="flex w-full items-end justify-between">
				<div class="flex h-auto items-end overflow-x-auto">
					<div class="flex items-center self-center px-1 mb-1">
						<Sidebar.Trigger />
					</div>
					<div class="thin-scrollbar scrollbar-thin flex items-end overflow-x-auto overflow-y-hidden gap-0.5 px-1">
						{#each openTabs as tab (tab.ID)}
							<div
								class="{getColorClass(tab.ActiveDBColor || "")} group relative flex items-center rounded-t-lg px-3 py-1.5 transition-all duration-150 select-none
									{tab.ID === tabID
										? 'bg-muted text-foreground z-10'
										: 'bg-background text-muted-foreground hover:bg-muted/50 hover:text-foreground'}"
								draggable="true"
								ondragstart={(e) => onTabDragStart(e, tab.ID)}
								ondragover={(e) => onTabDragOver(e, tab.ID)}
								ondragend={onTabDragEnd}
								role="tab"
								tabindex="0"
							>
								<button
									class="flex items-center gap-1.5 pr-2 text-sm font-medium truncate max-w-[160px]"
									onclick={() => setActiveTab(tab.ID)}
									title={tab.Name}
								>
									{tab.Name}
								</button>
								{#if tab.IsQueryRunning}
									<span class="mr-1 flex items-center" title="Query running">
										<Spinner class="size-3" />
									</span>
								{/if}
								<button
									class="ml-1 rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground hover:text-destructive hover:bg-muted"
									onclick={(e) => { e.stopPropagation(); deleteTab(tab.ID); }}
								>
									<X size={14} />
								</button>
								<!-- Active tab connector to content below -->
								{#if tab.ID === tabID}
									<div class="absolute bottom-0 left-0 right-0 h-[2px] bg-muted"></div>
								{/if}
							</div>
						{/each}
						{#if tabLoading}
							<div class="flex items-center self-center ml-2">
								<Button variant="ghost" size="sm" class="text-muted-foreground">
									<Spinner />
									Processing
								</Button>
							</div>
						{/if}
					</div>
				</div>
				<div class="flex mr-2 self-center">
					{#if chatPaneCollapsed}
						<Button class="text-xs p-1" variant="secondary" size="xs" onclick={toggleChatPane}>
							<Chat size={16} />
							Ask Stardust AI
						</Button>
					{/if}
				</div>
			</div>
		</header>

		<div class="flex h-screen flex-1 flex-col rounded-3xl bg-muted">
		{#if tabsMap.size > 0}
			<!-- Main Content on screen -->

			{#if tabType == 'table'}
					<div class="flex h-full flex-1 flex-col justify-center">
						<!-- Breadcrumb -->
						<div class="mt-1 flex items-center justify-between mx-2">
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
											<Breadcrumb.Page class='{tabDBPoolID === '' ? "text-red-500" : "text-green-500"}'>{tabName}</Breadcrumb.Page>
										</Breadcrumb.Item>
										{#if tabDBPoolID}
											{@render runControl(getTableData, 'Load table data')}
										{/if}
									</Breadcrumb.List>
								</Breadcrumb.Root>
							</div>
							<div class="flex px-2">
								<Tabs.Root value={tableViewTab}>
									<Tabs.List class="flex items-center justify-center gap-2">
										<Tabs.Trigger class="h-8 bg-background" value="data" onclick={() => (tableViewTab = 'data')}
											>Data</Tabs.Trigger
										>
										<Tabs.Trigger class="h-8 bg-background" value="manage" onclick={() => (tableViewTab = 'manage')}
											>Manage</Tabs.Trigger
										>
									</Tabs.List>
								</Tabs.Root>
							</div>
						</div>

						<!-- Content based on selected tab -->
						{#if tableViewTab === 'data'}
							<div class="flex h-screen flex-1 flex-col">
								<div class="px-2 pb-1 pt-1">
									<ClauseBar
										{tabID}
										tableName={tabName}
										columnNames={tableColumns}
										bind:select
										bind:where
										bind:orderBy
										bind:groupBy
										onrun={getTableData}
									/>
								</div>
								<div class="relative flex h-full flex-1 overflow-hidden">
									<div class="flex h-full w-full overflow-hidden mx-1 pb-1">
										{#if queryLoading}
										<LaserLoader />
									{/if}
									{#if $columns.length > 0}
											{#key tabID}
												<DataTableManual
													tabID={tabID}
													tableName={tabName}
													getTablePageData={getTablePageData}
													{lastQueryExecutionTime}
													{select}
													{where}
													{orderBy}
													{groupBy}
												/>
											{/key}
										{/if}
									</div>
								</div>
							</div>
						{:else if tableViewTab === 'manage'}
							<div class="mt-2 flex flex-1 flex-col overflow-hidden">
								<ManageTable 
									tabID={tabID}
									tabName={tabName} />
							</div>
						{/if}
					</div>
			{:else}
					<Tabs.Content value={tabID.toString()} class="flex flex-col flex-1 overflow-hidden">
						<div class="flex h-full flex-col">
							 <Breadcrumb.Root class="mx-4 mt-2">
								<Breadcrumb.List>
									<Breadcrumb.Item>
										<Breadcrumb.Link>{tabConnName}</Breadcrumb.Link>
									</Breadcrumb.Item>
									<Breadcrumb.Separator />
									<Breadcrumb.Item>
										<Breadcrumb.Link class='{tabDBPoolID === '' ? "text-red-500" : "text-green-500"}'>{tabDBName}</Breadcrumb.Link>
									</Breadcrumb.Item>
									{#if tabDBPoolID}
										{@render runControl(() => executeQuery(), 'Run query')}
									{/if}
								</Breadcrumb.List>
							</Breadcrumb.Root>
							

							<!-- Resizable Panes for Editor and Output -->
							<Resizable.ResizablePaneGroup direction="vertical" class="h-full">
								<!-- SQL Editor Pane -->
								<Resizable.Pane
									defaultSize={editorHeight}
									minSize={10}
									class="rsz-pane rounded-t-3xl rounded-lg mx-1 pt-1"
								>
									<SqlEditor
										bind:value={editor}
										bind:selectedQuery={$selectedQuery}
										suggestions={Array.from($suggestions)}
										executeQuery={executeQuery}
									/>
								</Resizable.Pane>

								<Resizable.ResizableHandle withHandle />

								<!-- Output Pane -->
								<Resizable.Pane
									defaultSize={outputHeight}
									minSize={10}
									class="rsz-pane"
								>
									<div class="relative h-full mx-1 pb-1">
										{#if queryLoading}
											<LaserLoader />
										{/if}
										{#if $columns.length > 0}
											{#key tabID}
												<DataTable
													tabID={tabID}
													tableName={executeQueryTableName} 
													executeQuery={executeQuery}
													{lastQueryExecutionTime}
												/>
											{/key}
										{/if}
									</div>
								</Resizable.Pane>
							</Resizable.ResizablePaneGroup>
						</div>
					</Tabs.Content>
			{/if}
			
			{/if}
		</div>
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
