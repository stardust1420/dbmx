<script lang="ts">
	import { ModeWatcher } from 'mode-watcher';
	import '../app.css';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	let { children } = $props();

	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import ActivityIcon from 'lucide-svelte/icons/activity';
	import TerminalIcon from '@lucide/svelte/icons/terminal';
	import UserIcon from '@lucide/svelte/icons/user';
	import BotIcon from '@lucide/svelte/icons/bot';
	import * as Command from '$lib/components/ui/command/index.js';

	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { GetPostgresConnections } from '$lib/wailsjs/go/app/Connections';
	import { GetAllTabs } from '$lib/wailsjs/go/app/Tabs';

	import { postgresConnectionsMap } from '$lib/state.svelte';

	// Fetch all connections when the root layout is mounted
	// This is mounted only for the root layout, so it will only run once
	// This will also run when the app is reloaded
	onMount(() => {
		GetPostgresConnections()
			.then((connections) => {
				for (let connection of connections) {
					postgresConnectionsMap.set(connection.ID, connection);
				}
			})
			.catch((error) => {
				// Handle errors from the EstablishPostgresDatabaseConnection call
				toast.error('Connection Failed', {
					description: error.message,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	});

	function goToRoute(route: string) {
		goto(route);
		open = false;
	}

	let open = $state(false);

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'p' && (e.metaKey || e.ctrlKey)) {
			e.preventDefault();
			open = !open;
		}
	}
</script>

<svelte:document onkeydown={handleKeydown} />
<Command.Dialog bind:open>
	<Command.Input placeholder="Type a command or search..." />
	<Command.List>
		<Command.Empty>No results found.</Command.Empty>
		<Command.Group heading="Menu">
			<Command.Item onSelect={() => goToRoute('/')}>
				<TerminalIcon class="mr-2 size-4" />
				<span>DBMX</span>
				<Command.Shortcut>⌘P</Command.Shortcut>
			</Command.Item>
			<Command.Item onSelect={() => goToRoute('/connections')}>
				<ActivityIcon class="mr-2 size-4" />
				<span>Connections</span>
				<Command.Shortcut>⌘C</Command.Shortcut>
			</Command.Item>
			<Command.Item onSelect={() => goToRoute('/user')}>
				<UserIcon class="mr-2 size-4" />
				<span>User</span>
				<Command.Shortcut>⌘P</Command.Shortcut>
			</Command.Item>
			<Command.Item onSelect={() => goToRoute('/billing')}>
				<CreditCardIcon class="mr-2 size-4" />
				<span>Billing</span>
				<Command.Shortcut>⌘B</Command.Shortcut>
			</Command.Item>
			<Command.Item onSelect={() => goToRoute('/settings')}>
				<BotIcon class="mr-2 size-4" />
				<span>LLM Manager</span>
				<Command.Shortcut>⌘S</Command.Shortcut>
			</Command.Item>
		</Command.Group>
	</Command.List>
</Command.Dialog>

<Toaster />

<ModeWatcher />
{@render children()}
