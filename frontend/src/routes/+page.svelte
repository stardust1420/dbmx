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
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
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
	import { onMount, tick } from 'svelte';
	import { dbmx, model } from '$lib/wailsjs/go/models';
	import { tabsMap, modelID, availableModels, userIsAIEnabled } from '$lib/state.svelte';
	import { SaveNewChatMessage } from '$lib/wailsjs/go/app/Tabs';
	import { Chat, ListAvailableModels } from '$lib/wailsjs/go/app/Stardust';
	import { mode } from 'mode-watcher';
	import Spinner from '$lib/components/ui/spinner/spinner.svelte';
	import { Marked } from 'marked';
	import { markedHighlight } from 'marked-highlight';
    import hljs from 'highlight.js';

	// Import a theme for code block styling
    import 'highlight.js/styles/github-dark.css';

	// Create a marked instance configured with highlight.js
    const marked = new Marked(
        markedHighlight({
            emptyLangClass: 'hljs',
            langPrefix: 'hljs language-',
            highlight(code, lang) {
                const language = hljs.getLanguage(lang) ? lang : 'plaintext';
                return hljs.highlight(code, { language }).value;
            }
        })
	);

	function renderMarkdown(content: string): string {
		let html = marked.parse(content) as string;
		// Wrap <table> elements in a scrollable container
		html = html.replace(/<table>/g, '<div class="table-wrapper"><table>').replace(/<\/table>/g, '</table></div>');
		// Add copy button to code blocks
		html = html.replace(/<pre><code(.*?)>/g, '<pre class="code-block-wrapper"><button class="copy-btn" onclick="(function(btn){const code=btn.parentElement.querySelector(\'code\').innerText;navigator.clipboard.writeText(code);btn.textContent=\'Copied!\';setTimeout(()=>btn.textContent=\'Copy\',1500)})(this)">Copy</button><code$1>');
		return html;
	}

	function splitThinking(content: string): { thinking: string | null; response: string } {
		const thinkMatch = content.match(/<think>([\s\S]*?)<\/think>/);
		if (thinkMatch) {
			const thinking = thinkMatch[1].trim();
			const response = content.replace(/<think>[\s\S]*?<\/think>/, '').trim();
			return { thinking, response };
		}
		return { thinking: null, response: content };
	}

	let expandedThinking: Record<string, boolean> = $state({});

	function toggleThinking(id: string) {
		expandedThinking[id] = !expandedThinking[id];
	}

	let availableModelsLoading = $state(false)

	let listAvailableModels = () => {
		availableModelsLoading = true
		ListAvailableModels()
		.then((models) => {
			if (models.length > 0) {
				$modelID = models[0].id
			}
			$availableModels = models
			availableModelsLoading = false
		})
		.catch((error) => {
			toast.error('Failed to fetch available models', {
				description: String(error),
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			availableModelsLoading = false
		})
	}


	onMount(() => {
		listAvailableModels();
	});
 
	
	const triggerContent = $derived(
		$availableModels.find((f) => f.id === $modelID)?.normalized_name ?? ""
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
		if (!$userIsAIEnabled) {
			toast.error('Not Enabled', {
					description: "Stardust AI is not enabled. Please enable it in the LLM Manager",
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			return;
		}
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

		Chat($modelID, aiChat)
		.then((chatRes) => {
			isAiTyping = false;
			const aiMsg = new model.AIMsg();
			aiMsg.ID = chatRes.ID;
			aiMsg.Role = chatRes.Role;
			aiMsg.Content = chatRes.Content;
			aiMsg.CreatedAt = new Date().toISOString();

			saveMsg(currentTabId, aiMsg);

			scrollToBottom();
		})
		.catch((error) => {
			isAiTyping = false;
			toast.error('Failed to chat', {
				description: String(error),
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})

		// // Simulate AI response (replace with real API call)
		// setTimeout(async () => {
		// 	isAiTyping = false;
		// 	const aiMsg = new model.AIMsg();
		// 	aiMsg.ID = crypto.randomUUID().toString();
		// 	aiMsg.Role = 'assistant';
		// 	aiMsg.Content = 'I can help you with that! Let me analyze your database schema and get back to you.';
		// 	aiMsg.CreatedAt = new Date().toISOString();

		// 	saveMsg(currentTabId, aiMsg);

		// 	await scrollToBottom();
		// }, 1500);
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
						<div class="flex flex-col gap-2.5 py-2">
							{#each aiChat as message (message.ID)}
								{#if message.Role === 'user'}
									<!-- User message -->
									<div class="chat-message-in flex items-start justify-end gap-2">
										<div class="flex flex-col items-end gap-0.5 max-w-[80%]">
											<div class="rounded-2xl rounded-br-sm bg-neutral-700 px-3 py-1.5 text-base text-neutral-100">
												<p class="leading-relaxed whitespace-pre-wrap">{message.Content}</p>
											</div>
											<span class="text-[10px] text-neutral-500 px-1">{formatTime(new Date(message.CreatedAt))}</span>
										</div>
										<div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-neutral-700/80 mt-0.5">
											<User size={12} class="text-neutral-400" />
										</div>
									</div>
								{:else}
									{@const parts = splitThinking(message.Content)}
									<!-- AI message -->
									<div class="chat-message-in flex items-start gap-2">
										<div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-neutral-700/50 mt-0.5">
											<Sparkles size={12} class="text-neutral-400" />
										</div>
										<div class="flex flex-col gap-1 max-w-[95%] min-w-0">
											{#if parts.thinking}
												<button
													class="thinking-toggle flex items-center gap-1 text-[11px] text-neutral-500 hover:text-neutral-300 transition-colors py-0.5"
													onclick={() => toggleThinking(message.ID)}
												>
													<ChevronRight
														size={11}
														class="transition-transform duration-500 {expandedThinking[message.ID] ? 'rotate-90' : ''}"
													/>
													Thinking
												</button>
												{#if expandedThinking[message.ID]}
													<div class="ai-prose thinking-content rounded-lg bg-neutral-900/50 px-3 py-2 text-xs text-neutral-500 border border-neutral-800">
														{@html renderMarkdown(parts.thinking)}
													</div>
												{/if}
											{/if}
											<div class="ai-prose rounded-2xl rounded-tl-sm bg-neutral-800/60 px-3.5 py-2.5 text-base text-neutral-200 border border-neutral-700/40">
												{@html renderMarkdown(parts.response)}
											</div>
											<span class="text-[10px] text-neutral-500 px-1">{formatTime(new Date(message.CreatedAt))}</span>
										</div>
									</div>
								{/if}
							{/each}

							{#if isAiTyping}
								<!-- Typing indicator -->
								<div class="chat-message-in flex items-start gap-2">
									<div class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-neutral-700/50 mt-0.5">
										<Sparkles size={12} class="text-neutral-400" />
									</div>
									<div class="rounded-2xl rounded-tl-sm bg-neutral-800/60 px-4 py-3 border border-neutral-700/40">
										<div class="flex items-center gap-1.5">
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-neutral-400"></span>
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-neutral-400" style="animation-delay: 0.15s"></span>
											<span class="typing-dot h-1.5 w-1.5 rounded-full bg-neutral-400" style="animation-delay: 0.3s"></span>
										</div>
									</div>
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
			<div
				class="flex flex-[1] flex-col items-center justify-center rounded-3xl bg-neutral-800 mr-2 mt-1"
			>
			{#if availableModelsLoading}
				<Spinner class="size-6 text-yellow-500"/>
				<span>Loading models...</span>
			{:else}
				{#if $availableModels.length == 0}
					<span class="p-4">No models found. Please configure them in LLM Manager</span>
				{:else}
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
						<Select.Root type="single" name="model" bind:value={$modelID}>
							<Select.Trigger class="w-auto m-1">
								{triggerContent}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
								{#each $availableModels as model (model.id)}
									<Select.Item
									value={model.id}
									label={model.normalized_name}
									>
									{model.normalized_name}
									</Select.Item>
								{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					<Button variant="outline" class="m-1 rounded-full dark:hover:bg-white" onclick={sendMessage}><Play size={12} fill="black" /></Button>
					</div>
				{/if}
			{/if}
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

	/* Markdown prose styling for AI messages */
	:global(.ai-prose) {
		line-height: 1.6;
		word-wrap: break-word;
		overflow-wrap: break-word;
	}

	:global(.ai-prose p) {
		margin: 0.5em 0;
	}

	:global(.ai-prose p:first-child) {
		margin-top: 0;
	}

	:global(.ai-prose p:last-child) {
		margin-bottom: 0;
	}

	:global(.ai-prose h1),
	:global(.ai-prose h2),
	:global(.ai-prose h3),
	:global(.ai-prose h4) {
		font-weight: 600;
		margin: 1em 0 0.5em;
		line-height: 1.3;
	}

	:global(.ai-prose h1) { font-size: 1.25em; }
	:global(.ai-prose h2) { font-size: 1.15em; }
	:global(.ai-prose h3) { font-size: 1.05em; }

	:global(.ai-prose ul),
	:global(.ai-prose ol) {
		margin: 0.5em 0;
		padding-left: 1.5em;
	}

	:global(.ai-prose li) {
		margin: 0.25em 0;
	}

	:global(.ai-prose code) {
		font-family: 'JetBrains Mono', 'Fira Code', ui-monospace, monospace;
		font-size: 0.85em;
		background: rgba(255, 255, 255, 0.08);
		padding: 0.15em 0.4em;
		border-radius: 4px;
	}

	:global(.ai-prose pre) {
		position: relative;
		margin: 0.75em 0;
		border-radius: 8px;
		overflow-x: auto;
		background: #1a1b26 !important;
		border: 1px solid rgba(255, 255, 255, 0.08);
	}

	:global(.ai-prose pre .copy-btn) {
		position: absolute;
		top: 6px;
		right: 6px;
		padding: 2px 8px;
		font-size: 0.7em;
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.15);
		border-radius: 4px;
		color: #a0a0b0;
		cursor: pointer;
		opacity: 0;
		transition: opacity 0.2s;
	}

	:global(.ai-prose pre:hover .copy-btn) {
		opacity: 1;
	}

	:global(.ai-prose pre .copy-btn:hover) {
		background: rgba(255, 255, 255, 0.2);
		color: #e0e0e0;
	}

	:global(.ai-prose pre code) {
		display: block;
		padding: 0.85em 1em;
		background: transparent !important;
		font-size: 0.8em;
		line-height: 1.5;
		white-space: pre;
		overflow-x: auto;
	}

	:global(.ai-prose blockquote) {
		border-left: 3px solid rgba(99, 102, 241, 0.5);
		margin: 0.75em 0;
		padding: 0.25em 0.75em;
		color: rgba(255, 255, 255, 0.7);
	}

	:global(.ai-prose a) {
		color: #818cf8;
		text-decoration: underline;
		text-underline-offset: 2px;
	}

	:global(.ai-prose strong) {
		font-weight: 600;
		color: #e2e8f0;
	}

	:global(.ai-prose .table-wrapper) {
		overflow-x: auto;
		overflow-y: hidden;
		margin: 0.75em 0;
		border-radius: 6px;
		max-width: 100%;
		scrollbar-width: thin;
		scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
	}

	:global(.ai-prose .table-wrapper table) {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85em;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	:global(.ai-prose table) {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.85em;
	}

	:global(.ai-prose th),
	:global(.ai-prose td) {
		border: 1px solid rgba(255, 255, 255, 0.1);
		padding: 0.4em 0.75em;
		text-align: left;
		white-space: normal !important;
		overflow: visible !important;
		text-overflow: unset !important;
		max-width: none !important;
		min-width: unset !important;
		height: auto !important;
	}

	:global(.ai-prose th) {
		background: rgba(255, 255, 255, 0.05);
		font-weight: 600;
	}

	:global(.ai-prose hr) {
		border: none;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
		margin: 1em 0;
	}

	/* Thinking accordion */
	.thinking-toggle {
		cursor: pointer;
		background: none;
		border: none;
		font-family: inherit;
	}

	:global(.thinking-content) {
		animation: thinkingFadeIn 0.5s ease-out;
	}

	@keyframes thinkingFadeIn {
		from {
			opacity: 0;
			transform: translateY(-4px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	:global(.thinking-content.ai-prose) {
		max-height: 200px;
		overflow-y: auto;
		scrollbar-width: thin;
		scrollbar-color: rgba(255, 255, 255, 0.1) transparent;
	}
</style>
