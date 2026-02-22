<script lang="ts">
	import AppSidebar from '$lib/components/app/sidebar/sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import Tabs from '$lib/components/app/main_screen/tabs.svelte';
	import * as Resizable from '$lib/components/ui/resizable/index.js';
	import X from '@lucide/svelte/icons/x';
	import SquarePen from '@lucide/svelte/icons/square-pen';
	import Plus from '@lucide/svelte/icons/plus';
	import Ellipsis from '@lucide/svelte/icons/ellipsis';
	import Paperclip from '@lucide/svelte/icons/paperclip';

	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Kbd from '$lib/components/ui/kbd/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';

	// Active tab properties
	let tabID = $state(0);
	let tabName = $state('');
	let tabType = $state('');

	// Table tab active db properties
	let tabTableDBPoolID = $state('');
	let tabConnName = $state('');
	let tabDBName = $state('');
	let tabConnID = $state(0);

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
		connID?: number,
		dbName?: string,
		tableDBPoolID?: string,
		connName?: string
	) {
		if (tabsComponent && tabsComponent.addTab) {
			tabsComponent.addTab(tableName, connID, dbName, tableDBPoolID, connName);
		}
	}

	// Right Sidebar
	let collapsed = $state(false);
	let chatPane: ReturnType<typeof Resizable.Pane>;
	let chatPaneSize = $state(0);

	function toggleChatPane() {
		if (collapsed) {
			chatPane.expand();
			if (chatPaneSize > 0) {
				chatPane.resize(chatPaneSize);
			} else {
				chatPane.resize(30);
				chatPaneSize = 30;
			}
		} else {
			chatPaneSize = chatPane.getSize();
			chatPane.collapse();
		}
	}
</script>

<Resizable.PaneGroup direction="horizontal">
	<Resizable.Pane defaultSize={100}>
		<Sidebar.Provider>
			<AppSidebar
				bind:tabID
				bind:tabName
				bind:tabTableDBPoolID
				bind:tabConnID
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
					bind:tabConnName
					bind:tabDBName
					bind:tabConnID
					bind:select
					bind:limit
					bind:offset
					bind:where
					bind:orderBy
					bind:groupBy
					bind:tableColumns
					{toggleChatPane}
				/>
			</Sidebar.Inset>
		</Sidebar.Provider>
	</Resizable.Pane>

	<Resizable.Handle />
	
	<Resizable.Pane
		defaultSize={0}
		maxSize={40}
		collapsible={true}
		collapsedSize={0}
		onCollapse={() => (collapsed = true)}
		onExpand={() => (collapsed = false)}
		bind:this={chatPane}
		class="flex h-full"
	>
		<div class="mx-1 my-2 flex w-full flex-col">
			<!-- header -->
			<div class="flex flex-[1] items-center justify-between">
				<SquarePen size={18} class="m-2" />
				<p>Chat</p>
				<X size={18} class="m-2" onclick={toggleChatPane} />
			</div>
			<div class="flex flex-[20] items-center justify-center">
				<div class="m-2 flex flex-col items-center">
					<p class="text-muted-foreground text-sm">
						Use
						<Kbd.Group>
							<Kbd.Root>+</Kbd.Root>
						</Kbd.Group>
						to add tables or columns as context
					</p>
				</div>
			</div>
			<div class="ml-2 flex w-full flex-[1] items-center justify-start">
				<Badge class="m-1 my-2 bg-green-700 text-white">
					<p class="font-medium font-normal">Explain</p>
				</Badge>
				<Badge class="m-1 my-2 bg-blue-700 text-white">
					<p class="font-medium font-normal">Optimise</p>
				</Badge>
				<Badge class="m-1 my-2 bg-red-700 text-white">
					<p class="font-medium font-normal">Fix Syntax</p>
				</Badge>
			</div>
			<div
				class="bg-secondary/40 h-42 flex flex-[5] flex-col items-center justify-center rounded-xl border p-2"
			>
				<div class="m-2 flex h-36 w-full flex-[3] items-center justify-start">
					<div class=" flex h-11 items-center justify-center rounded-lg">
						<Badge
							variant="secondary"
							class="mr-2 flex h-11 items-center justify-center rounded-xl bg-gray-800 pr-0 pt-0"
						>
							<p class="font-medium font-normal">Line 10-13</p>
							<X size={15} color="red" class="mb-1 ml-1 self-start rounded-full hover:bg-red-700" />
						</Badge>
					</div>
					<div class=" flex h-11 items-center justify-center rounded-lg">
						<Badge
							variant="secondary"
							class="mr-2 flex h-11 items-center justify-center rounded-lg bg-gray-800 pr-0 pt-0"
						>
							<p class="font-medium font-normal">transaction</p>

							<X size={14} color="red" class="mb-1 ml-1 self-start rounded-full hover:bg-red-700" />
						</Badge>
					</div>
				</div>
				<div class="flex h-24 max-h-24 w-full flex-[3] items-center justify-center">
					<Textarea value="Explain" class="h-24 max-h-24 bg-black" placeholder="Ask anything..." />
				</div>
				<div class="flex w-full flex-[1] items-end justify-start">
					<Plus size={16} class="m-2" />
					<Paperclip size={16} class="m-2" />
					<Ellipsis size={16} class="m-2" />
				</div>
			</div>
		</div>
	</Resizable.Pane>
</Resizable.PaneGroup>
