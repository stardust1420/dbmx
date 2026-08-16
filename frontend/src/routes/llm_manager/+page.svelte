<script lang="ts">
	import ChevronLeft from '@lucide/svelte/icons/chevron-left';
	import Switch from '$lib/components/ui/switch/switch.svelte';
	import * as Accordion from "$lib/components/ui/accordion/index.js";
	import Input from '$lib/components/ui/input/input.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import { isLoggedIn, userIsAIEnabled, userUseDefaultKey } from '$lib/state.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { AddProviderAPIKey, DeleteProviderAPIKey, DisableStardustAI, EnableStardustAI, ListProviders, SwitchDefaultKey, UpdateProviderAPIKey } from '$lib/wailsjs/go/app/Stardust';
	import { toast } from 'svelte-sonner';
	import { dbmx } from '$lib/wailsjs/go/models';


	let isAIEnabled = $state(false);
	let useStardustModels = $state(false);
	let userProviders: dbmx.UserProvider[] = $state([])

	onMount(() => {
		isAIEnabled = $userIsAIEnabled;
		useStardustModels = $userUseDefaultKey;
		listProviders();
	});

	let listProviders = () => {
		if (useStardustModels) {
			return
		}
		ListProviders()
		.then((providers) => {
			userProviders = providers
		})
		.catch((error) => {
			toast.error('Failed to fetch user providers', {
				description: error,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
	}

	let addProviderAPIKey = (provider:string, apiKey:string) => {
		if (useStardustModels) {
			return
		}
		AddProviderAPIKey(provider, apiKey)
		.then(() => {
			listProviders();
			toast.success('Success', {
				description: "API Key added successfully",
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
		.catch((error) => {
			toast.error('Failure', {
				description: error,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
	}

	let updateProviderAPIKey = (keyID: string, provider: string, apiKey: string) => {
		if (useStardustModels) {
			return
		}
		UpdateProviderAPIKey(keyID, provider, apiKey)
		.then(() => {
			listProviders();
			toast.success('Success', {
				description: "API Key updated successfully",
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
		.catch((error) => {
			toast.error('Failure', {
				description: error,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
	}

	let deleteProviderAPIKey = (keyID: string, provider: string) => {
		if (useStardustModels) {
			return
		}
		DeleteProviderAPIKey(keyID, provider)
		.then(() => {
			listProviders();
			toast.success('Success', {
				description: "API Key updated successfully",
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
		.catch((error) => {
			toast.error('Failure', {
				description: error,
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
	}

	const providerNames: Record<string, string> = {
		anthropic: "Anthropic",
		gemini: "Google Gemini",
		groq: "Groq",
		mistral: "Mistral AI",
		ollama: "Ollama",
		openai: "OpenAI",
		openrouter: "OpenRouter"
	};

	let toggleStardustAI = (checked: boolean) => {
		if (checked) {
			EnableStardustAI()
			.then((customer) => {
				$userIsAIEnabled = true
				$userUseDefaultKey = true
				useStardustModels = true
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
					useStardustModels = false
					toast.success('Success', {
						description: "Disabled Stardust AI",
						action: {
							label: 'OK',
							onClick: () => console.info('OK')
						}
					});
				})
				.catch((error) => {
					isAIEnabled = !isAIEnabled
					toast.error('Failed to disable Stardust AI', {
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
			if (!checked) {
				listProviders();
			}
			$userUseDefaultKey = checked
			toast.success('Success', {
				description: "Switched Default Key",
				action: {
					label: 'OK',
					onClick: () => console.info('OK')
				}
			});
		})
		.catch((error) => {
			useStardustModels = !useStardustModels
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
		
										{#each userProviders as item (item.provider)}
											<Accordion.Item value={item.provider} class="border-2 border-neutral-700 rounded-2xl px-4">
												<Accordion.Trigger class="ml-3">
													{providerNames[item.provider] || item.provider}
												</Accordion.Trigger>
												<Accordion.Content>
													<div class="flex items-center gap-3">
														<Input 
															id={item.provider} 
															placeholder={`${providerNames[item.provider] || item.provider} API Key`} 
															bind:value={item.api_key} 
															class="flex-1 m-0.5" 
														/>
														{#if item.key_id && item.key_id.trim() !== ''}
															<Button variant="secondary" onclick={() => updateProviderAPIKey(item.key_id, item.provider, item.api_key)} >Update</Button>
															<Button variant="destructive" onclick={() => deleteProviderAPIKey(item.key_id, item.provider)} >Remove</Button>
														{:else}
															<Button variant="secondary" onclick={() => addProviderAPIKey(item.provider, item.api_key)} >Add</Button>
														{/if}
													</div>
												</Accordion.Content>
											</Accordion.Item>
										{/each}

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
