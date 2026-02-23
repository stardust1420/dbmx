<script lang="ts">
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import UserIcon from '@lucide/svelte/icons/user';
	import ActivityIcon from 'lucide-svelte/icons/activity';
	import CreditCardIcon from '@lucide/svelte/icons/credit-card';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import BotIcon from '@lucide/svelte/icons/bot';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SparklesIcon from '@lucide/svelte/icons/sparkles';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { GetLoggedInUser } from '$lib/wailsjs/go/app/Auth';
	import { toast } from 'svelte-sonner';
	import { Logout } from '$lib/wailsjs/go/app/Auth';

	const sidebar = useSidebar();

	let open = $state(false);

	function goToRoute(route: string) {
		goto(route);
		open = false;
	}

	
	let username = $state("No Account");
	let email = $state("No Account");
	let avatar = $state("https://api.dicebear.com/9.x/avataaars-neutral/svg?backgroundRotation=0,360");
	let isLoggedIn = $state(false);
	
	onMount(() => {
		GetLoggedInUser()
			.then((user) => {
				username = user.fullname;
				email = user.email;
				avatar = "https://api.dicebear.com/9.x/fun-emoji/svg?seed=Aneka";
				isLoggedIn = true;
			})
			.catch((error) => {
				console.log(error);
			});
	});

	let logout = () => {
		Logout()
			.then(() => {
				toast.success('Logged out successfully');
				isLoggedIn = false;
				username = "No Account";
				email = "No Account";
				avatar = "https://api.dicebear.com/9.x/avataaars-neutral/svg?backgroundRotation=0,360";
			})
			.catch((error) => {
				toast.error('Failed to log out', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			});
	};

</script>

<Sidebar.Menu>
	<Sidebar.MenuItem>
		<DropdownMenu.Root bind:open>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Sidebar.MenuButton
						size="lg"
						class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
						{...props}
					>
						<Avatar.Root class="size-8 rounded-lg">
							<Avatar.Image src={avatar} alt={username} />
							<Avatar.Fallback class="rounded-lg">CN</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-medium">{username}</span>
							<span class="truncate text-xs">{email}</span>
						</div>
						<ChevronsUpDownIcon class="ml-auto size-4" />
					</Sidebar.MenuButton>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content
				class="w-(--bits-dropdown-menu-anchor-width) min-w-56 rounded-lg"
				side={sidebar.isMobile ? 'bottom' : 'right'}
				align="end"
				sideOffset={4}
			>
				<DropdownMenu.Label class="p-0 font-normal">
					<div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
						<Avatar.Root class="size-8 rounded-lg">
							<Avatar.Image src={avatar} alt={username} />
							<Avatar.Fallback class="rounded-lg">CN</Avatar.Fallback>
						</Avatar.Root>
						<div class="grid flex-1 text-left text-sm leading-tight">
							<span class="truncate font-medium">{username}</span>
							<span class="truncate text-xs">{email}</span>
						</div>
					</div>
				</DropdownMenu.Label>
				<DropdownMenu.Separator />
				<DropdownMenu.Group>
					<DropdownMenu.Item onSelect={() => goToRoute('/')}>
						<SparklesIcon />
						Upgrade to Pro
					</DropdownMenu.Item>
				</DropdownMenu.Group>
				<DropdownMenu.Separator />
				<DropdownMenu.Group>
					<DropdownMenu.Item onSelect={() => {
						if(isLoggedIn){
							goToRoute('/user')
						}else{
							goToRoute('/user/login')
						}
					}}>
						<UserIcon />
						User
					</DropdownMenu.Item>
					<DropdownMenu.Item onSelect={() => goToRoute('/connections')}>
						<ActivityIcon />
						Connections
					</DropdownMenu.Item>
					<DropdownMenu.Item onSelect={() => goToRoute('/llm_manager')}>
						<BotIcon />
						LLM Manager
					</DropdownMenu.Item>
					<DropdownMenu.Item onSelect={() => goToRoute('/billing')}>
						<CreditCardIcon />
						Billing
					</DropdownMenu.Item>
				</DropdownMenu.Group>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</Sidebar.MenuItem>
</Sidebar.Menu>
