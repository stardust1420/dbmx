<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Chart from '$lib/components/ui/chart/index.js';
	import { IsMobile } from '$lib/hooks/is-mobile.svelte.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import Plus from '@lucide/svelte/icons/plus';
	import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";

	const isMobile = new IsMobile();
	let { scheme }: { scheme: string } = $props();
	
	const environments = [
		{ envValue: 'local', label: 'Local' },
		{ envValue: 'feature', label: 'Feature' },
		{ envValue: 'staging', label: 'Staging' },
		{ envValue: 'production', label: 'Production' }
	];
	let envValue = $state('');
	const triggerEnvSelect = $derived(
		environments.find((e) => e.envValue === envValue)?.label ?? 'Select Env'
	);

	const colors = [
		{ colorValue: 'bg-red-500', label: 'Red' },
		{ colorValue: 'bg-purple-500', label: 'Purple' },
		{ colorValue: 'bg-blue-500', label: 'Blue' },
		{ colorValue: 'bg-orange-500', label: 'Orange' },
		{ colorValue: 'bg-green-500', label: 'Green' }
	];
	let colorValue = $state('');
	const triggerColorSelect = $derived(
		colors.find((e) => e.colorValue === colorValue)?.label ?? 'Select Color'
	);

	const sslModes = [
		{ sslModeValue: 'PREFERRED', label: 'PREFERRED' },
		{ sslModeValue: 'DISABLED', label: 'DISABLED' },
		{ sslModeValue: 'ALLOW', label: 'ALLOW' },
		{ sslModeValue: 'REQUIRED', label: 'REQUIRED' },
		{ sslModeValue: 'VERIFY-CA', label: 'VERIFY-CA' },
		{ sslModeValue: 'VERIFY-FULL', label: 'VERIFY-FULL' }
	];
	let sslModeValue = $state('');
	const triggerSslModeSelect = $derived(
		sslModes.find((e) => e.sslModeValue === sslModeValue)?.label ?? 'PREFERRED'
	);

	let name = $state('');
	let host = $state('');
	let port = $state('');
	let username = $state('');
	let password = $state('');
	let database = $state('');

	let overSSH = $state(false);
	let useSSHKey = $state(false);
</script>

<Drawer.Root direction={isMobile.current ? 'bottom' : 'right'}>

	<DropdownMenu.Root>
		<DropdownMenu.Trigger class="mr-6">
			<Button variant="outline">
				<Plus class="size-4" />
				New Connection
			</Button>
		</DropdownMenu.Trigger>
		<DropdownMenu.Content class="w-40" align="start">
			<DropdownMenu.Label>Schemes</DropdownMenu.Label>
				<Drawer.Trigger class="flex w-full">

					<DropdownMenu.Item class="flex w-full">
						PostgreSQL
					</DropdownMenu.Item>
				</Drawer.Trigger>
			<DropdownMenu.Item>
				MySQL
			</DropdownMenu.Item>
			<DropdownMenu.Item>
				ClickHouse
			</DropdownMenu.Item>
			<DropdownMenu.Item>
				SQLite
			</DropdownMenu.Item>
		</DropdownMenu.Content>
	</DropdownMenu.Root>

	<Drawer.Content>
		<Drawer.Header class="gap-1 font-mono">
			<Drawer.Title class="text-3xl">PostgreSQL</Drawer.Title>
			<Drawer.Description>New Connection</Drawer.Description>
		</Drawer.Header>
		<div class="flex flex-col gap-2 overflow-y-auto px-4 text-sm font-mono">

			<form class="flex flex-col gap-2">
				<h1 class="mb-2 text-2xl font-bold">Basic</h1>

				<div class="flex flex-col gap-2">
					<Label for="host">Host</Label>
					<Input id="host" placeholder="localhost" value={host}/>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<div class="flex flex-col gap-2">
						<Label for="port">Port</Label>
						<Input id="port" placeholder="5432" value={port}/>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="database">Database</Label>
						<Input id="database" placeholder="postgres" value={database}/>
					</div>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="username">Username</Label>
					<Input id="username" placeholder="username" value={username}/>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="password">Password</Label>
					<Input id="password" placeholder="password" value={password}/>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="saveAs">Save As</Label>
					<Input id="saveAs" placeholder="Name your connection" value={name}/>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<div class="flex flex-col gap-2 pb-1">
						<Label for="env">Env</Label>
						<Select.Root type="single" name="environment" bind:value={envValue}>
							<Select.Trigger class="w-full">
								{triggerEnvSelect}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each environments as e (e.envValue)}
										<Select.Item value={e.envValue} label={e.label} class={e.envValue}>
											{e.label}
										</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex flex-col gap-2 pb-1">
						<Label for="color">Color</Label>
						<Select.Root type="single" name="color" bind:value={colorValue}>
							<Select.Trigger class={colorValue + ' w-full bg-opacity-60 '}>
									{triggerColorSelect}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each colors as c}
										<Select.Item
											value={c.colorValue}
											label={c.label}
											class={c.colorValue + ' bg-opacity-60'}
										>
											{c.label}
										</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>
				</div>

				<h1 class="mb-2 text-2xl font-bold">Advanced</h1>

				<div class="grid grid-cols-2 gap-2">
					<div class="flex flex-col gap-2 pb-1">
						<Label for="sslMode">SSL Mode</Label>
						<Select.Root type="single" name="sslMode" bind:value={sslModeValue}>
							<Select.Trigger class={sslModeValue + ' w-full bg-opacity-60'}>
								{triggerSslModeSelect}
							</Select.Trigger>
							<Select.Content>
								<Select.Group>
									{#each sslModes as c}
										<Select.Item
											value={c.sslModeValue}
											label={c.label}
											class={c.sslModeValue + ' bg-opacity-60'}
										>
											{c.label}
										</Select.Item>
									{/each}
								</Select.Group>
							</Select.Content>
						</Select.Root>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="clientKey">Client Key</Label>
						<Input id="clientKey" type="file" />
					</div>
				</div>
				<div class="grid grid-cols-2 gap-2">
					<div class="flex flex-col gap-2">
						<Label for="serverCert">Server Cert</Label>
						<Input id="serverCert" type="file" />
					</div>
					<div class="flex flex-col gap-2">
						<Label for="rootCACert">Root CA Cert</Label>
						<Input id="rootCACert" type="file" />
					</div>
				</div>

				<div class="flex flex-row items-center gap-4 my-2">
					<Label for="overSSH">Over SSH</Label>
					<Switch id="overSSH" bind:checked={overSSH} />
				</div>

				{#if overSSH}
					<Label for="sshHost">SSH Host</Label>
					<Input id="sshHost" placeholder="127.0.0.1" />
					<Label for="sshPort">SSH Port</Label>
					<Input id="sshPort" placeholder="22" />
					<Label for="sshUsername">SSH Username</Label>
					<Input id="sshUsername" placeholder="username" />
					<Label for="sshPassword">SSH Password</Label>
					<Input id="sshPassword" placeholder="password" />

					<div class="flex flex-row items-center gap-4 my-2">
						<Label for="useSSHKey">Use SSH Key</Label>
						<Switch id="useSSHKey" bind:checked={useSSHKey} />
					</div>

					{#if useSSHKey}
						<Label for="sshKey">SSH Key</Label>
						<Input id="sshKey" type="file" />
					{/if}
				{/if}

			</form>
		</div>
		<Drawer.Footer>
			<div class="flex flex-row gap-2">
				<Button class="w-96 self-center" variant="secondary">Save</Button>
				<Button class="w-96 self-center" variant="default">Test</Button>
				<Button class="w-96 self-center" variant="destructive">Connect</Button>
			</div>
		</Drawer.Footer>
	</Drawer.Content>
</Drawer.Root>
