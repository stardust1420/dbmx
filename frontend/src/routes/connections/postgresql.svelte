<script lang="ts">
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';

	import * as Select from '$lib/components/ui/select/index.js';
	import Button from '$lib/components/ui/button/button.svelte';

	const environments = [
		{ envValue: 'local', label: 'Local' },
		{ envValue: 'dev', label: 'Development' },
		{ envValue: 'test', label: 'Test' },
		{ envValue: 'staging', label: 'Staging' },
		{ envValue: 'production', label: 'Production' }
	];
	let envValue = $state('');
	const triggerEnvSelect = $derived(
		environments.find((e) => e.envValue === envValue)?.label ?? 'Select Environment'
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
		colors.find((e) => e.colorValue === colorValue)?.label ?? 'Select Environment'
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
	let env = $state('');
	let host = $state('');
	let port = $state('');
	let username = $state('');
	let password = $state('');
	let database = $state('');
	let colour = $state('');

	let overSSH = $state(false);
	let useSSHKey = $state(false);
</script>

<div class="h-100 flex flex-col gap-2 p-4">
	<div class="flex flex-1 flex-row overflow-hidden">
		<div class="mb-2 flex w-full flex-col gap-2 p-4">
			<h1 class="mb-2 font-mono text-2xl font-bold">Basic</h1>

			<Label for="host">Host</Label>
			<Input bind:value={host} id="host" placeholder="127.0.0.1" />
			<Label for="port">Port</Label>
			<Input bind:value={port} id="port" placeholder="5432" />
			<Label for="username">Username</Label>
			<Input bind:value={username} id="username" placeholder="username" />
			<Label for="password">Password</Label>
			<Input bind:value={password} id="password" placeholder="password" />
			<Label for="database">Database</Label>
			<Input
				bind:value={database}
				id="database"
				placeholder="database name (defaults to 'postgres')"
			/>

			<h1 class="mb-2 mt-4 font-mono text-2xl font-bold">Save As</h1>

			<Label for="name">Name</Label>
			<Input bind:value={name} id="name" placeholder="name" />
			<Label for="environment">Environment</Label>
			<Select.Root type="single" name="environment" bind:value={envValue}>
				<Select.Trigger class="w-[180px]">
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

			<Label for="color">Color</Label>
			<Select.Root type="single" name="color" bind:value={colorValue}>
				<Select.Trigger class={colorValue + ' w-[180px] bg-opacity-60'}>
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

		<div class="mb-2 flex w-full flex-col gap-2 overflow-hidden p-4">
			<h1 class="mb-2 font-mono text-2xl font-bold">Advanced</h1>

			<Label for="sslMode">SSL Mode</Label>
			<Select.Root type="single" name="sslMode" bind:value={sslModeValue}>
				<Select.Trigger class={sslModeValue + ' w-[180px] bg-opacity-60'}>
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

			<Label for="clientKey">Client Key</Label>
			<Input id="clientKey" type="file" />
			<Label for="serverCert">Server Cert</Label>
			<Input id="serverCert" type="file" />
			<Label for="rootCACert">Root CA Cert</Label>
			<Input id="rootCACert" type="file" />

			<div class="flex flex-row items-center gap-4">
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

				<div class="flex flex-row items-center gap-4">
					<Label for="useSSHKey">Use SSH Key</Label>
					<Switch id="useSSHKey" bind:checked={useSSHKey} />
				</div>

				{#if useSSHKey}
					<Label for="sshKey">SSH Key</Label>
					<Input id="sshKey" type="file" />
				{/if}
			{/if}
		</div>
	</div>

	<div class="flex flex-row gap-10">
		<Button class="w-96 self-center" variant="secondary">Save</Button>
		<Button class="w-96 self-center" variant="default">Test</Button>
		<Button class="w-96 self-center" variant="destructive">Connect</Button>
	</div>
</div>
