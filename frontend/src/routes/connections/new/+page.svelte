<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { buttonVariants } from '$lib/components/ui/button/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import Postgresql from '../postgresql.svelte';

	let selectedView = $state('PostgreSQL');
</script>

<div class="flex h-full flex-col items-center justify-center overflow-hidden p-12">
	<div class="flex h-full w-full flex-col self-center overflow-hidden">
		<div class="flex h-12 flex-[1] items-center justify-between p-4">
			<h1 class="font-mono text-6xl font-bold">New Connection</h1>
		</div>

		<div class="h-100 flex w-full flex-[8] flex-row overflow-hidden p-4">
			<!-- Tab Management -->
			<div class="h-100 mr-2 flex w-full flex-[1] flex-col overflow-hidden">
				<div class="mb-0.5 flex w-full flex-1 items-center justify-center">
					<Button
						variant={selectedView === 'PostgreSQL' ? 'default' : 'ghost'}
						class="flex h-full w-full flex-col items-center justify-center"
						onclick={() => (selectedView = 'PostgreSQL')}
					>
						<Avatar.Root>
							<Avatar.Image
								src="https://www.postgresql.org/media/img/about/press/elephant.png"
								alt="PostgreSQL"
							/>
							<Avatar.Fallback>PostgreSQL</Avatar.Fallback>
						</Avatar.Root>
					</Button>
				</div>
				<div class="mb-0.5 flex flex-1 items-center justify-center">
					<Button
						variant={selectedView === 'ClickHouse' ? 'default' : 'ghost'}
						class="flex h-full w-full flex-col items-center justify-center"
						onclick={() => (selectedView = 'ClickHouse')}
					>
						<Avatar.Root>
							<Avatar.Image
								src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTHqAYbPCD1YHc6q8sMJ8A6KEzW8eU8GsahZw&s"
								alt="ClickHouse"
							/>
							<Avatar.Fallback>ClickHouse</Avatar.Fallback>
						</Avatar.Root>
					</Button>
				</div>
				<div class="mt-0.5 flex flex-1 items-center justify-center">
					<Button
						variant={selectedView === 'MySQL' ? 'default' : 'ghost'}
						class="flex h-full w-full flex-col items-center justify-center"
						onclick={() => (selectedView = 'MySQL')}
					>
						<Avatar.Root>
							<Avatar.Image
								src="https://www.mysql.com/common/logos/logo-mysql-170x115.png"
								alt="MySQL"
							/>
							<Avatar.Fallback>MySQL</Avatar.Fallback>
						</Avatar.Root>
					</Button>
				</div>
				<div class="mt-0.5 flex flex-1 items-center justify-center">
					<Button
						variant={selectedView === 'SQLite3' ? 'default' : 'ghost'}
						class="flex h-full w-full flex-col items-center justify-center"
						onclick={() => (selectedView = 'SQLite3')}
					>
						<Avatar.Root>
							<Avatar.Image src="https://sqlite.org/images/sqlite370_banner.svg" alt="SQLite3" />
							<Avatar.Fallback>SQLite3</Avatar.Fallback>
						</Avatar.Root>
					</Button>
				</div>
				<div class="mt-0.5 flex flex-1 items-center justify-center">
					<Button
						variant={selectedView === 'MongoDB' ? 'default' : 'ghost'}
						class="flex h-full w-full flex-col items-center justify-center"
						onclick={() => (selectedView = 'MongoDB')}
					>
						<Avatar.Root>
							<Avatar.Image
								src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTkIyzuKbWXglpMjXFmRqCSJOaRfyswluGasA&s"
								alt="MongoDB"
							/>
							<Avatar.Fallback>MongoDB</Avatar.Fallback>
						</Avatar.Root>
					</Button>
				</div>
			</div>

			<!-- Tab Content -->
			<div class="flex h-full flex-[15] overflow-hidden">
				{#if selectedView === 'PostgreSQL'}
					<div class="mx-4 flex h-full flex-1 flex-col overflow-auto">
						<Postgresql />
					</div>
				{:else if selectedView === 'ClickHouse'}
					<div class="mx-4 flex h-full flex-1 flex-col overflow-auto">
						<Postgresql />
					</div>
				{:else if selectedView === 'MySQL'}
					<div class="mx-4 flex h-full flex-1 flex-col overflow-auto">
						<Postgresql />
					</div>
				{:else if selectedView === 'SQLite3'}
					<div class="mx-4 flex h-full flex-1 flex-col overflow-auto">
						<Postgresql />
					</div>
				{:else if selectedView === 'MongoDB'}
					<div class="mx-4 flex h-full flex-1 flex-col overflow-auto">
						<Postgresql />
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
