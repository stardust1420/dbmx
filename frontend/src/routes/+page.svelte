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
	import * as Tooltip from "$lib/components/ui/tooltip/index.js";

	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Kbd from '$lib/components/ui/kbd/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import Button from '$lib/components/ui/button/button.svelte';

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

	function toggleChatPane() {
		if (chatPaneCollapsed) {
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

	// Right Sidebar
	let chatPaneCollapsed: boolean = $state(false);
	let chatPane: ReturnType<typeof Resizable.Pane>;
	let chatPaneSize: number = $state(0);
</script>

<Resizable.PaneGroup direction="horizontal">
	<Resizable.Pane defaultSize={100} class="mr-1">
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
					bind:chatPaneCollapsed
					{toggleChatPane}
				/>
			</Sidebar.Inset>
		</Sidebar.Provider>
	</Resizable.Pane>

	<Resizable.Handle />
	
	<Resizable.Pane
		defaultSize={0}
		maxSize={50}
		collapsible={true}
		collapsedSize={0}
		onCollapse={() => (chatPaneCollapsed = true)}
		onExpand={() => (chatPaneCollapsed = false)}
		bind:this={chatPane}
		class="flex rounded-lg bg-neutral-900 my-2"
	>
		<div class="flex w-full h-full flex-col">
			<!-- header -->
			<div class="flex flex-[1] items-center justify-between m-1">
				<SquarePen size={18} class="m-2" />
				<p>Stardust AI</p>
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
			<div
				class="flex flex-[1] flex-col items-center justify-center rounded-t-3xl bg-black"
			>
				<div class="flex w-full flex-[5] items-center justify-center">
					<Textarea class="max-h-48 m-1 focus-visible:ring-0 border-0" placeholder="Ask anything..." />
				</div>
				<div class="flex w-full flex-[1] items-end justify-start">
				<Tooltip.Provider>
					<Tooltip.Root>
						<Tooltip.Trigger>
							<Button variant="ghost"><Plus size={12} /></Button>
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p>Add context</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</Tooltip.Provider>
				</div>
			</div>
		</div>
	</Resizable.Pane>
</Resizable.PaneGroup>
