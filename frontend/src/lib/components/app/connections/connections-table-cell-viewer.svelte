<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
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
	let env = $state('');
	const triggerEnvSelect = $derived(
		environments.find((e) => e.envValue === env)?.label ?? 'Select Env'
	);

	const colors = [
		{ colorValue: 'bg-red-500', label: 'Red' },
		{ colorValue: 'bg-purple-500', label: 'Purple' },
		{ colorValue: 'bg-blue-500', label: 'Blue' },
		{ colorValue: 'bg-orange-500', label: 'Orange' },
		{ colorValue: 'bg-green-500', label: 'Green' }
	];
	let color = $state('');
	const triggerColorSelect = $derived(
		colors.find((e) => e.colorValue === color)?.label ?? 'Select Color'
	);

	const sslModes = [
		{ sslModeValue: 'PREFERRED', label: 'PREFERRED' },
		{ sslModeValue: 'DISABLED', label: 'DISABLED' },
		{ sslModeValue: 'ALLOW', label: 'ALLOW' },
		{ sslModeValue: 'REQUIRED', label: 'REQUIRED' },
		{ sslModeValue: 'VERIFY-CA', label: 'VERIFY-CA' },
		{ sslModeValue: 'VERIFY-FULL', label: 'VERIFY-FULL' }
	];
	let sslMode = $state('');
	const triggerSSlModeSelect = $derived(
		sslModes.find((e) => e.sslModeValue === sslMode)?.label ?? 'Select SSL Mode'
	);

	let name = $state('');
	let host = $state('');
	let port = $state('');
	let username = $state('');
	let password = $state('');
	let database = $state('');

	let overSSH = $state(false);
	let useSSHKey = $state(false);
	let isAdvanced = $state(false);

	let clientKeyFile = $state<File | null>(null)
	let clientCertFile = $state<File | null>(null)
	let rootCACertFile = $state<File | null>(null)

	let sshHost = $state('');
	let sshPort = $state('');
	let sshUsername = $state('');
	let sshPassword = $state('');

	let sshKeyFile = $state<File | null>(null)

		
	import { AddPostgresConnection, TestConnectPostgres } from '$lib/wailsjs/go/app/Connections';
	import { toast } from 'svelte-sonner';
	import { model } from '$lib/wailsjs/go/models';



	let connectPostgres = async () => {
		let missingFields = [];
		if (name.trim() === '') {
			missingFields.push('name');
		}
		if (env.trim() in ['', 'Select Env']) {
			missingFields.push('env');
		}
		if (host.trim() === '') {
			missingFields.push('host');
		}
		if (port.trim() === '') {
			missingFields.push('port');
		}
		if (username.trim() === '') {
			missingFields.push('username');
		}
		if (password.trim() === '') {
			missingFields.push('password');
		}
		if (color.trim() in ['', 'Select Color'] ) {
			missingFields.push('color');
		}

		if (missingFields.length > 0) {
			toast.error('Missing required field(s)', {
				description: `${missingFields.join(', ')}`,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		const clientKeyBytes = clientKeyFile
		? new Uint8Array(await clientKeyFile.arrayBuffer())
		: null

		const clientCertBytes = clientCertFile
		? new Uint8Array(await clientCertFile.arrayBuffer())
		: null

		const rootCACertBytes = rootCACertFile
		? new Uint8Array(await rootCACertFile.arrayBuffer())
		: null

		const sshKeyBytes = sshKeyFile
		? new Uint8Array(await sshKeyFile.arrayBuffer())
		: null

		const connectPostgresData = new model.Connection({
			Engine: 'postgresql',
			Host: host,
			Port: port,
			Username: username,
			Password: password,
			Database: database,
			Name: name,
			Env: env,
			Color: color,
			IsAdvanced: isAdvanced,
			SSLMode: sslMode,
			ClientKey: clientKeyBytes,
			ClientCert: clientCertBytes,
			RootCACert: rootCACertBytes,
			OverSSH: overSSH,
			SSHHost: sshHost,
			SSHPort: sshPort,
			SSHUsername: sshUsername,
			SSHPassword: sshPassword,
			UseSSHKey: useSSHKey,
			SSHKey: sshKeyBytes
		});
		try {
			// Await the result of the TestConnectPostgres call
			await AddPostgresConnection(connectPostgresData);

			// If successful, show success message
			toast.success('Success', {
				description: 'The connection was saved successfully',
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		} catch (error) {
			let errorMessage = 'An error occurred while trying to connect.';

			// Check if error is an instance of Error
			if (error instanceof Error) {
				errorMessage = error.message; // Access message property safely
			} else if (typeof error === 'string') {
				errorMessage = error; // If it's a string, use it directly
			}
			// Handle errors from the TestConnectPostgres call
			toast.error('Connection Failed', {
				description: errorMessage,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		}
	};

	let testConnectPostgres = async () => {
		let missingFields = [];
		if (name.trim() === '') {
			missingFields.push('name');
		}
		if (env.trim() in ['', 'Select Env']) {
			missingFields.push('env');
		}
		if (host.trim() === '') {
			missingFields.push('host');
		}
		if (port.trim() === '') {
			missingFields.push('port');
		}
		if (username.trim() === '') {
			missingFields.push('username');
		}
		if (password.trim() === '') {
			missingFields.push('password');
		}
		if (color.trim() in ['', 'Select Color'] ) {
			missingFields.push('color');
		}

		if (missingFields.length > 0) {
			toast.error('Missing required field(s)', {
				description: `${missingFields.join(', ')}`,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
			return;
		}

		const clientKeyBytes = clientKeyFile
		? new Uint8Array(await clientKeyFile.arrayBuffer())
		: null

		const clientCertBytes = clientCertFile
		? new Uint8Array(await clientCertFile.arrayBuffer())
		: null

		const rootCACertBytes = rootCACertFile
		? new Uint8Array(await rootCACertFile.arrayBuffer())
		: null

		const sshKeyBytes = sshKeyFile
		? new Uint8Array(await sshKeyFile.arrayBuffer())
		: null

		const testConnectPostgresData = new model.Connection({
			Engine: 'postgresql',
			Host: host,
			Port: port,
			Username: username,
			Password: password,
			Database: database,
			Name: name,
			Env: env,
			Color: color,
			IsAdvanced: isAdvanced,
			SSLMode: sslMode,
			ClientKey: clientKeyBytes,
			ClientCert: clientCertBytes,
			RootCACert: rootCACertBytes,
			OverSSH: overSSH,
			SSHHost: sshHost,
			SSHPort: sshPort,
			SSHUsername: sshUsername,
			SSHPassword: sshPassword,
			UseSSHKey: useSSHKey,
			SSHKey: sshKeyBytes
		});
		try {
			// Await the result of the TestConnectPostgres call
			await TestConnectPostgres(testConnectPostgresData);

			// If successful, show success message
			toast.success('Test Successful', {
				description: 'You can successfully connect using the given credentials',
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		} catch (error) {
			let errorMessage = 'An error occurred while trying to connect.';

			// Check if error is an instance of Error
			if (error instanceof Error) {
				errorMessage = error.message; // Access message property safely
			} else if (typeof error === 'string') {
				errorMessage = error; // If it's a string, use it directly
			}
			// Handle errors from the TestConnectPostgres call
			toast.error('Connection Failed', {
				description: errorMessage,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		}

		const result = await TestConnectPostgres(testConnectPostgresData);
	};
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
			<DropdownMenu.Label>Engines</DropdownMenu.Label>
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
						<Select.Root type="single" name="environment" bind:value={env}>
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
						<Select.Root type="single" name="color" bind:value={color}>
							<Select.Trigger class={color + ' w-full bg-opacity-60 '}>
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

				<div class="flex flex-row items-center gap-4 my-2">
					<h1 class="mb-2 text-2xl font-bold">Advanced</h1>
					<Switch id="isAdvanced" class="self-center" bind:checked={isAdvanced} />
				</div>

				{#if isAdvanced}
					<div class="grid grid-cols-2 gap-2">
						<div class="flex flex-col gap-2 pb-1">
							<Label for="sslMode">SSL Mode</Label>
							<Select.Root type="single" name="sslMode" bind:value={sslMode}>
								<Select.Trigger class={sslMode + ' w-full bg-opacity-60'}>
									{triggerSSlModeSelect}
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
								<Input id="clientKey" type="file"
								onchange={(e) => {
									const input = e.currentTarget as HTMLInputElement
									clientKeyFile = input.files?.[0] ?? null
								}}
							/>
						</div>
					</div>
					<div class="grid grid-cols-2 gap-2">
						<div class="flex flex-col gap-2">
							<Label for="clientCert">Client Cert</Label>
							<Input id="clientCert" type="file"
								onchange={(e) => {
									const input = e.currentTarget as HTMLInputElement
									clientCertFile = input.files?.[0] ?? null
								}}
							/>
						</div>
						<div class="flex flex-col gap-2">
							<Label for="rootCACert">Root CA Cert</Label>
							<Input id="rootCACert" type="file"
								onchange={(e) => {
									const input = e.currentTarget as HTMLInputElement
									rootCACertFile = input.files?.[0] ?? null
								}}
							/>
						</div>
					</div>

					<div class="flex flex-row items-center gap-4 my-2">
						<Label for="overSSH">Over SSH</Label>
						<Switch id="overSSH" bind:checked={overSSH} />
					</div>

					{#if overSSH}
						<Label for="sshHost">SSH Host</Label>
						<Input id="sshHost" placeholder="127.0.0.1" bind:value={sshHost}/>
						<Label for="sshPort">SSH Port</Label>
						<Input id="sshPort" placeholder="22" bind:value={sshPort}/>
						<Label for="sshUsername">SSH Username</Label>
						<Input id="sshUsername" placeholder="username" bind:value={sshUsername}/>
						<Label for="sshPassword">SSH Password</Label>
						<Input id="sshPassword" placeholder="password" bind:value={sshPassword}/>

						<div class="flex flex-row items-center gap-4 my-2">
							<Label for="useSSHKey">Use SSH Key</Label>
							<Switch id="useSSHKey" bind:checked={useSSHKey} />
						</div>

						{#if useSSHKey}
							<Label for="sshKey">SSH Key</Label>
							<Input id="sshKey" type="file"
								onchange={(e) => {
									const input = e.currentTarget as HTMLInputElement
									sshKeyFile = input.files?.[0] ?? null
								}}
							/>
						{/if}
					{/if}
				{/if}
			</form>
		</div>
		<Drawer.Footer>
			<div class="flex flex-row gap-2">
				<Button class="w-96 self-center" variant="secondary"
					onclick={connectPostgres}
				>Save</Button>
				<Button class="w-96 self-center" variant="default"
					onclick={testConnectPostgres}
				>Test</Button>
			</div>
		</Drawer.Footer>
	</Drawer.Content>
</Drawer.Root>
