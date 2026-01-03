<script lang="ts">
	import * as Drawer from '$lib/components/ui/drawer/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { IsMobile } from '$lib/hooks/is-mobile.svelte.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';

	let {
		connectionID,
		newConnection,
		title,
		description,
	}: {
		connectionID?: number;
		newConnection: boolean;
		title: string | '';
		description: string;
	} = $props();
	
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

	let name = $derived(title);
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

		
	import { AddPostgresConnection, TestConnectPostgres, GetConnection, UpdateConnection } from '$lib/wailsjs/go/app/Connections';
	import { toast } from 'svelte-sonner';
	import { model } from '$lib/wailsjs/go/models';
	import { onMount } from 'svelte';



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

	let disabled = $derived(!newConnection);
	let checked = $derived(!disabled);

	onMount(async () => {
		console.log(connectionID);
		console.log(newConnection);
		if (!newConnection && connectionID) {
			try {
				const connection = await GetConnection(connectionID);
				name = connection.Name;
				env = connection.Env;
				host = connection.Host;
				port = connection.Port;
				username = connection.Username;
				password = connection.Password;
				database = connection.Database;
				color = connection.Color;
				isAdvanced = connection.IsAdvanced;
				sslMode = connection.SSLMode;
				overSSH = connection.OverSSH;
				sshHost = connection.SSHHost;
				sshPort = connection.SSHPort;
				sshUsername = connection.SSHUsername;
				sshPassword = connection.SSHPassword;
				useSSHKey = connection.UseSSHKey;
			}
			catch (error) {
				toast.error('Connection Failed', {
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			}
		}
	});

	let updateConnection = async () => {
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

		const updateConnectionData = new model.Connection({
			ID: connectionID,
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
			// Await the result of the UpdateConnection call
			await UpdateConnection(updateConnectionData);

			// If successful, show success message
			toast.success('Success', {
				description: 'The connection was updated successfully',
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
			// Handle errors from the UpdateConnection call
			toast.error('Connection Failed', {
				description: errorMessage,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		}
	};
</script>

<Drawer.Content>
	<Drawer.Header class="gap-1 font-mono">
		<Drawer.Title class="text-3xl">{name}</Drawer.Title>
		<Drawer.Description>{description}</Drawer.Description>
		{#if !newConnection}
			<div class="flex flex-row items-center gap-4 mt-2">
				<span class="text-sm font-medium">Update</span>
				<Switch id="isEditable" {checked}
					onCheckedChange={
						() => {
							disabled = !disabled
							if (disabled) {
								description = 'View Connection Details'
							} else {
								description = 'Update Connection Details'
							}
						}
					}
				/>
			</div>
		{/if}
	</Drawer.Header>
	<div class="flex flex-col gap-2 overflow-y-auto px-4 text-sm font-mono">

		<form class="flex flex-col gap-2">
			<h1 class="mb-2 text-2xl font-bold">Basic</h1>

			<div class="flex flex-col gap-2">
				<Label for="host">Host</Label>
				<Input id="host" placeholder="localhost" bind:value={host} {disabled}/>
			</div>
			<div class="grid grid-cols-2 gap-2">
				<div class="flex flex-col gap-2">
					<Label for="port">Port</Label>
					<Input id="port" placeholder="5432" bind:value={port} {disabled}/>
				</div>
				<div class="flex flex-col gap-2">
					<Label for="database">Database</Label>
					<Input id="database" placeholder="postgres" bind:value={database} {disabled}/>
				</div>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="username">Username</Label>
				<Input id="username" placeholder="username" bind:value={username} {disabled}/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="password">Password</Label>
				<Input id="password" placeholder="password" bind:value={password} {disabled}/>
			</div>
			<div class="flex flex-col gap-2">
				<Label for="saveAs">Save As</Label>
				<Input id="saveAs" placeholder="Name your connection" bind:value={name} {disabled}/>
			</div>
			<div class="grid grid-cols-2 gap-2">
				<div class="flex flex-col gap-2 pb-1">
					<Label for="env">Env</Label>
					<Select.Root type="single" name="environment" bind:value={env} {disabled}>
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
					<Select.Root type="single" name="color" bind:value={color} {disabled}>
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
				<Switch id="isAdvanced" class="self-center" bind:checked={isAdvanced} {disabled}/>
			</div>

			{#if isAdvanced}
				<div class="grid grid-cols-2 gap-2">
					<div class="flex flex-col gap-2 pb-1">
						<Label for="sslMode">SSL Mode</Label>
						<Select.Root type="single" name="sslMode" bind:value={sslMode} {disabled}>
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
							{disabled}
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
							{disabled}
						/>
					</div>
					<div class="flex flex-col gap-2">
						<Label for="rootCACert">Root CA Cert</Label>
						<Input id="rootCACert" type="file"
							onchange={(e) => {
								const input = e.currentTarget as HTMLInputElement
								rootCACertFile = input.files?.[0] ?? null
							}}
							{disabled}
						/>
					</div>
				</div>

				<div class="flex flex-row items-center gap-4 my-2">
					<Label for="overSSH">Over SSH</Label>
					<Switch id="overSSH" bind:checked={overSSH} {disabled}/>
				</div>

				{#if overSSH}
					<Label for="sshHost">SSH Host</Label>
					<Input id="sshHost" placeholder="127.0.0.1" bind:value={sshHost} {disabled}/>
					<Label for="sshPort">SSH Port</Label>
					<Input id="sshPort" placeholder="22" bind:value={sshPort} {disabled}/>
					<Label for="sshUsername">SSH Username</Label>
					<Input id="sshUsername" placeholder="username" bind:value={sshUsername} {disabled}/>
					<Label for="sshPassword">SSH Password</Label>
					<Input id="sshPassword" placeholder="password" bind:value={sshPassword} {disabled}/>

					<div class="flex flex-row items-center gap-4 my-2">
						<Label for="useSSHKey">Use SSH Key</Label>
						<Switch id="useSSHKey" bind:checked={useSSHKey} {disabled}/>
					</div>

					{#if useSSHKey}
						<Label for="sshKey">SSH Key</Label>
						<Input id="sshKey" type="file"
							onchange={(e) => {
								const input = e.currentTarget as HTMLInputElement
								sshKeyFile = input.files?.[0] ?? null
							}}
							{disabled}
						/>
					{/if}
				{/if}
			{/if}
		</form>
	</div>
	<Drawer.Footer>
		{#if newConnection}
			<div class="flex flex-row gap-2">
				<Button class="w-96 self-center bg-blue-700 hover:bg-blue-800" variant="outline"
					onclick={testConnectPostgres}
				>Test</Button>
				<Button class="w-96 self-center bg-green-700 hover:bg-green-800" variant="outline"
					onclick={connectPostgres}
				>Save</Button>
			</div>
		{:else if checked}
			<div class="flex flex-row gap-2">
				<Button class="w-96 self-center bg-blue-700 hover:bg-blue-800" variant="outline"
						onclick={testConnectPostgres}
				>Test</Button>
				<Button class="w-96 self-center bg-orange-700 hover:bg-orange-800" variant="outline"
					onclick={updateConnection}
				>Update</Button>
			</div>

		{/if}
	</Drawer.Footer>
</Drawer.Content>
