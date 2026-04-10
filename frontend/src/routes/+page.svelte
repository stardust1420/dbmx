<script lang="ts">
	import AppSidebar from '$lib/components/app/sidebar/sidebar.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import Tabs from '$lib/components/app/main_screen/tabs.svelte';
	import * as Resizable from '$lib/components/ui/resizable/index.js';
	import X from '@lucide/svelte/icons/x';
	import Play from '@lucide/svelte/icons/play';
	import Plus from '@lucide/svelte/icons/plus';
	import Ellipsis from '@lucide/svelte/icons/ellipsis';
	import Paperclip from '@lucide/svelte/icons/paperclip';
	import Sparkles from '@lucide/svelte/icons/sparkles';
	import User from '@lucide/svelte/icons/user';
	import * as Tooltip from "$lib/components/ui/tooltip/index.js";
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { toast } from 'svelte-sonner';


	import { Textarea } from '$lib/components/ui/textarea/index.js';
	import * as Kbd from '$lib/components/ui/kbd/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Select from "$lib/components/ui/select/index.js";
	import Label from '$lib/components/ui/label/label.svelte';
	import { tick } from 'svelte';
	import { model } from '$lib/wailsjs/go/models';
	import { tabsMap } from '$lib/state.svelte';
	import { SaveNewChatMessage } from '$lib/wailsjs/go/app/Tabs';


	const models = [
		{ value: "gemini-3.1-pro-high", label: "Gemini 3.1 Pro (High)" },
		{ value: "gemini-3.1-pro-low", label: "Gemini 3.1 Pro (Low)" },
		{ value: "gemini-3-flash-preview", label: "Gemini 3 Flash" },
		{ value: "claude-sonnet-4.6-thinking", label: "Claude Sonnet 4.6 (Thinking)" },
		{ value: "claude-opus-4.6-thinking", label: "Claude Opus 4.6 (Thinking)" },
		{ value: "gpt-oss-120b-medium", label: "GPT OSS 120B (Medium)" }
	];
 
	let value = $state("Claude Sonnet 4.6 (Thinking)");
	
	const triggerContent = $derived(
		models.find((f) => f.value === value)?.label ?? "Claude Sonnet 4.6 (Thinking)"
	);

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
	let aiChat: model.AIMsg[] = $state([]);

	$effect(() => {
		console.log(aiChat);
	});

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
			if (tabsMap.size === 0) {
				toast.error('No tab is open', {
						description: "Please open a tab to use AI features",
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
				return;
			}
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

	// Chat state

	let chatInput = $state('');
	let isAiTyping = $state(false);
	let chatScrollContainer: HTMLDivElement;

	async function scrollToBottom() {
		await tick();
		if (chatScrollContainer) {
			chatScrollContainer.scrollTop = chatScrollContainer.scrollHeight;
		}
	}

	function formatTime(date: Date): string {
		return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
	}

	let saveMsg = (currentTabID: number, msg: model.AIMsg) => {
		// If current tab id is equal to tabID, update the aiChat
		if (currentTabID === tabID) {
			aiChat.push(msg);
		}

		// Save in the tabs map
		let currentTab = tabsMap.get(currentTabID);
		if (currentTab) {
			if (!currentTab.AIChat) {
				// Initialize the array if null
				currentTab.AIChat = [];
			}
			currentTab.AIChat.push(msg);
			tabsMap.set(currentTabID, currentTab);
		}

		// Call backend function to save in DB
		SaveNewChatMessage(currentTabID, msg)
			.then(() => {
				console.log('Message saved successfully');
			})
			.catch((err) => {
				console.error('Failed to save message:', err);
			});
	}

	async function sendMessage() {
		if (tabsMap.size === 0) {
			toast.error('No tab is open', {
					description: "Please open a tab to use AI features",
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			return;
		}

		// Save the current tab id to make sure data doesn't get updated in wrong tab when quickly switching tabs
		const currentTabId = tabID;

		const text = chatInput.trim();
		if (!text) return;

		const msg = new model.AIMsg();
		msg.ID = crypto.randomUUID().toString();
		msg.Role = 'user';
		msg.Content = text;
		msg.CreatedAt = new Date().toISOString();

		saveMsg(currentTabId, msg);

		chatInput = '';
		await scrollToBottom();

		isAiTyping = true;
		await scrollToBottom();

		// Simulate AI response (replace with real API call)
		setTimeout(async () => {
			isAiTyping = false;
			const aiMsg = new model.AIMsg();
			aiMsg.ID = crypto.randomUUID().toString();
			aiMsg.Role = 'assistant';
			aiMsg.Content = 'I can help you with that! Let me analyze your database schema and get back to you.';
			aiMsg.CreatedAt = new Date().toISOString();

			saveMsg(currentTabId, aiMsg);

			await scrollToBottom();
		}, 1500);
	}

	function handleChatKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
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
					bind:aiChat
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
		class="flex rounded-lg bg-black my-2"
	>
		<div class="flex w-full h-full flex-col">
			<!-- header -->
			<div class="flex flex-[1] items-center justify-between">
				<p></p>
				<p>Stardust AI</p>
				<X size={18} class="m-2" onclick={toggleChatPane} />
			</div>
			<div class="flex flex-[20] flex-col overflow-hidden">
				<div
					bind:this={chatScrollContainer}
					class="chat-scroll flex-1 overflow-y-auto px-3 py-4"
				>
					{#if aiChat.length === 0}
						<!-- Empty state -->
						<div class="flex h-full flex-col items-center justify-center gap-4 opacity-60">
							<div class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-indigo-500/20 to-purple-500/20 ring-1 ring-white/10">
								<Sparkles size={22} class="text-indigo-400" />
							</div>
							<div class="flex flex-col items-center gap-1">
								<p class="text-sm font-medium text-neutral-300">Start a conversation</p>
								<p class="text-xs text-neutral-500 text-center leading-relaxed">
									Ask about your data, generate queries,<br/>or explore your schema.
								</p>
							</div>
							<p class="text-muted-foreground text-xs mt-2">
								Use
								<Kbd.Group>
									<Kbd.Root>+</Kbd.Root>
								</Kbd.Group>
								to add context
							</p>
						</div>
					{:else}
						<div class="flex flex-col gap-4">
							{#each aiChat as message (message.ID)}
								{#if message.Role === 'user'}
									<!-- User message -->
									<div class="chat-message-in flex items-end justify-end gap-2">
										<div class="flex flex-col items-end gap-1 max-w-[85%]">
											<div class="rounded-2xl rounded-br-md bg-gradient-to-br from-indigo-600 to-indigo-700 px-3.5 py-2.5 text-sm text-white shadow-lg shadow-indigo-500/10">
												<p class="leading-relaxed whitespace-pre-wrap">{message.Content}</p>
											</div>
											<span class="text-[10px] text-neutral-500 px-1">{formatTime(new Date(message.CreatedAt))}</span>
										</div>
										<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-neutral-700 ring-1 ring-white/10">
											<User size={14} class="text-neutral-300" />
										</div>
									</div>
								{:else}
									<!-- AI message -->
									<div class="chat-message-in flex items-end gap-2">
										<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500/30 to-purple-500/30 ring-1 ring-white/10">
											<Sparkles size={14} class="text-indigo-400" />
										</div>
										<div class="flex flex-col gap-1 max-w-[85%]">
											<div class="rounded-2xl rounded-bl-md bg-neutral-800/80 px-3.5 py-2.5 text-sm text-neutral-200 shadow-lg ring-1 ring-white/5">
												<p class="leading-relaxed whitespace-pre-wrap">{message.Content}</p>
											</div>
											<span class="text-[10px] text-neutral-500 px-1">{formatTime(new Date(message.CreatedAt))}</span>
										</div>
									</div>
								{/if}
							{/each}

							{#if isAiTyping}
								<!-- Typing indicator -->
								<div class="chat-message-in flex items-end gap-2">
									<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-indigo-500/30 to-purple-500/30 ring-1 ring-white/10">
										<Sparkles size={14} class="text-indigo-400" />
									</div>
									<div class="rounded-2xl rounded-bl-md bg-neutral-800/80 px-4 py-3 shadow-lg ring-1 ring-white/5">
										<div class="flex items-center gap-1">
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-indigo-400"></span>
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-indigo-400" style="animation-delay: 0.15s"></span>
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-indigo-400" style="animation-delay: 0.3s"></span>
										</div>
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
			<div
				class="flex flex-[1] flex-col items-center justify-center rounded-3xl bg-neutral-800 mr-2"
			>
				<div class="flex w-full flex-[5] items-center justify-center">
					<Textarea bind:value={chatInput} onkeydown={handleChatKeydown} class="max-h-48 m-1 focus-visible:ring-0 border-0" placeholder="Ask anything..." />
				</div>
				<div class="flex w-full flex-[1] items-end justify-between">
				<Tooltip.Provider>
					<Tooltip.Root>
						<Tooltip.Trigger>
							<Button variant="ghost" class="m-1"><Plus size={12} /></Button>
						</Tooltip.Trigger>
						<Tooltip.Content>
							<p>Add context</p>
						</Tooltip.Content>
					</Tooltip.Root>
				</Tooltip.Provider>
				<Select.Root type="single" name="model" bind:value>
					<Select.Trigger class="w-auto m-1">
						{triggerContent}
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
						{#each models as model (model.value)}
							<Select.Item
							value={model.value}
							label={model.label}
							>
							{model.label}
							</Select.Item>
						{/each}
						</Select.Group>
					</Select.Content>
				</Select.Root>
				<Button variant="outline" class="m-1 rounded-full dark:hover:bg-white" onclick={sendMessage}><Play size={12} fill="black" /></Button>
				</div>
			</div>
		</div>
	</Resizable.Pane>
</Resizable.PaneGroup>

<style>
	/* Typing indicator animation */
	:global(.typing-dot) {
		animation: typingBounce 1.2s ease-in-out infinite;
	}

	@keyframes typingBounce {
		0%, 60%, 100% {
			transform: translateY(0);
			opacity: 0.4;
		}
		30% {
			transform: translateY(-4px);
			opacity: 1;
		}
	}

	/* Message slide-in animation */
	:global(.chat-message-in) {
		animation: messageSlideIn 0.3s ease-out;
	}

	@keyframes messageSlideIn {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Custom scrollbar for the chat area */
	:global(.chat-scroll) {
		scrollbar-width: thin;
		scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
	}

	:global(.chat-scroll::-webkit-scrollbar) {
		width: 4px;
	}

	:global(.chat-scroll::-webkit-scrollbar-track) {
		background: transparent;
	}

	:global(.chat-scroll::-webkit-scrollbar-thumb) {
		background-color: rgba(255, 255, 255, 0.1);
		border-radius: 4px;
	}

	:global(.chat-scroll::-webkit-scrollbar-thumb:hover) {
		background-color: rgba(255, 255, 255, 0.2);
	}
</style>
