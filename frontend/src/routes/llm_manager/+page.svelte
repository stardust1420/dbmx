<script lang="ts">
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import Switch from '$lib/components/ui/switch/switch.svelte';
	import * as Accordion from "$lib/components/ui/accordion/index.js";
	import Input from '$lib/components/ui/input/input.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import { isLoggedIn, userIsAIEnabled, userUseDefaultKey } from '$lib/state.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { DisableStardustAI, EnableStardustAI, SwitchDefaultKey } from '$lib/wailsjs/go/app/Stardust';
	import { toast } from 'svelte-sonner';


	let isAIEnabled = $state(false);
	let useStardustModels = $state(true);

	onMount(() => {
		isAIEnabled = $userIsAIEnabled;
		useStardustModels = $userUseDefaultKey;
	});

	let toggleStardustAI = (checked: boolean) => {
		if (checked) {
			EnableStardustAI()
			.then((customer) => {
				$userIsAIEnabled = true
				$userUseDefaultKey = true
				toast.success('Success', {
					description: "Enabled Stardust AI",
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			})
			.catch((error) => {
				isAIEnabled = !checked
				toast.error('Failed to enable Stardust AI', {
					description: error,
					action: {
						label: 'OK',
						onClick: () => console.info('OK')
					}
				});
			})
		} else {
			DisableStardustAI()
				.then((success) => {
					$userIsAIEnabled = false
					$userUseDefaultKey = false
					toast.success('Success', {
						description: "Switched Default Key",
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
				})
				.catch((error) => {
					toast.error('Failed to enable Stardust AI', {
						description: error,
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
				})
		}
	}

	let toggleDefaultKey = (checked: boolean) => {
		SwitchDefaultKey(checked)
		.then((success) => {
			toast.success('Success', {
				description: "Switched Default Key",
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
		.catch((error) => {
			toast.error('Failed to enable Stardust AI', {
				description: error,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
	}


</script>

<div class="flex h-full flex-col items-center justify-center">
	<div class="flex h-full w-full flex-col self-center">
		<div class="flex flex-[1] items-center justify-center pt-4">
			<div class="flex items-center justify-center flex-row w-full">
				<a class="flex flex-[1] items-center justify-center" href="/">
					<ChevronLeft size={32} />
				</a>
				<h1 class="flex flex-[9] items-center justify-center font-mono text-6xl font-bold">LLM Manager</h1>
				<div class="flex-[1] items-center justify-center">
					<!-- Empty box -->
				</div>
			</div>
		</div>

		{#if $isLoggedIn}

			<div class="flex flex-[6] items-center justify-center">
				<div class="flex flex-[1] h-full w-full mx-48 p-2 flex-col gap-4">
					<div class="flex h-16 px-4">
						<div class="flex flex-[8] p-2 flex-col">
							<h1 class="text-lg">Enable Stardust AI</h1>
							<p>Enables Stardust AI assistant that helps with Text-to-SQL generation</p>
						</div>
						<div class="flex flex-[2] items-center justify-center rounded-3xl border border-blue-600">
							<Switch id="isAIEnabled" bind:checked={isAIEnabled} onCheckedChange={toggleStardustAI}/>
						</div>
					</div>
					{#if isAIEnabled}
					<div class="flex ml-8 h-full flex-col gap-4">
						<div class="flex h-16 px-4">
							<div class="flex flex-[8] p-2 flex-col">
								<h1 class="text-lg">Use Stardust models</h1>
								<p>Use default models provided by DBMX. <span class="text-green-500 font-bold">Disable this to Bring Your Own Keys (BYOK)</span></p>
							</div>
							<div class="flex flex-[2] items-center justify-center rounded-3xl border border-green-600">
								<Switch id="useStardustModels" bind:checked={useStardustModels} onCheckedChange={toggleDefaultKey}/>
							</div>
						</div>
						{#if !useStardustModels}
							<div class="flex px-4 flex-col gap-4">
								<div class="flex p-2 flex-col">
									<h1 class="text-lg">Bring Your Own Key (BYOK)</h1>
									<p>Add your API keys to enable AI models from different providers.</p>
								</div>
								<div class="flex flex-col gap-3 mr-24">
									<Accordion.Root type="single" collapsible class="flex flex-col gap-3">
		
										<Accordion.Item value="openai" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">OpenAI</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="OpenAI" placeholder="OpenAI API Key" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>
												</div>
											</Accordion.Content>
										</Accordion.Item>

										<Accordion.Item value="anthropic" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">Anthropic</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="Anthropic" placeholder="Anthropic API Key" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>

												</div>
											</Accordion.Content>
										</Accordion.Item>

										<Accordion.Item value="groq" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">Groq</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="Groq" placeholder="Groq API Key" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>

												</div>
											</Accordion.Content>
										</Accordion.Item>

										<Accordion.Item value="google" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">Google</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="Google" placeholder="Google API Key" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>

												</div>
											</Accordion.Content>
										</Accordion.Item>

										<Accordion.Item value="openrouter" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">OpenRouter</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="OpenRouter" placeholder="OpenRouter API Key" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>

												</div>
											</Accordion.Content>
										</Accordion.Item>

										<Accordion.Item value="ollama" class="border-2 border-neutral-700 rounded-2xl px-4">
											<Accordion.Trigger class="ml-3">Ollama</Accordion.Trigger>
											<Accordion.Content>
												<div class="flex items-center gap-3">
													<Input id="Ollama" placeholder="Ollama URL" class="flex-1 m-0.5" />
													<Button variant="secondary">Update</Button>
													<Button variant="destructive">Remove</Button>

												</div>
											</Accordion.Content>
										</Accordion.Item>

									</Accordion.Root>
								</div>
							</div>
						{/if}
					</div>
					{/if}
				</div>
			</div>

		{:else}
			<div class="flex flex-[6] items-center justify-center">
				<Button onclick={() => {goto('/user/login');}}>Go to Login</Button>
			</div>

		{/if}
	</div>
</div>
