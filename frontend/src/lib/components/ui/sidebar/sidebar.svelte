<script lang="ts">
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { cn } from '$lib/utils.js';
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import { SIDEBAR_WIDTH_MOBILE } from './constants.js';
	import { useSidebar } from './context.svelte.js';

	let {
		ref = $bindable(null),
		side = 'left',
		variant = 'sidebar',
		collapsible = 'offcanvas',
		class: className,
		children,
		tabID = $bindable(0),
		tabName = $bindable(''),
		tabTableDBPoolID = $bindable(''),
		tabConnID = $bindable(0),
		tabDBName = $bindable(''),
		...restProps
	}: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		side?: 'left' | 'right';
		variant?: 'sidebar' | 'floating' | 'inset';
		collapsible?: 'offcanvas' | 'icon' | 'none';
		tabID?: number;
		tabName?: string;
		tabTableDBPoolID?: string;
		tabConnID?: number;
		tabDBName?: string;
	} = $props();

	const sidebar = useSidebar();

	const MIN_WIDTH = 180;
	const MAX_WIDTH = 480;

	function startResize(e: MouseEvent) {
		e.preventDefault();
		const startX = e.clientX;
		const startWidth = parseInt(sidebar.sidebarWidth, 10);

		function onMouseMove(e: MouseEvent) {
			const delta = side === 'left' ? e.clientX - startX : startX - e.clientX;
			const newWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, startWidth + delta));
			sidebar.sidebarWidth = `${newWidth}px`;
		}

		function onMouseUp() {
			document.removeEventListener('mousemove', onMouseMove);
			document.removeEventListener('mouseup', onMouseUp);
			document.body.style.cursor = '';
			document.body.style.userSelect = '';
		}

		document.body.style.cursor = 'col-resize';
		document.body.style.userSelect = 'none';
		document.addEventListener('mousemove', onMouseMove);
		document.addEventListener('mouseup', onMouseUp);
	}
</script>

{#if collapsible === 'none'}
	<div
		class={cn(
			'bg-sidebar text-sidebar-foreground flex h-full w-[--sidebar-width] flex-col',
			className
		)}
		bind:this={ref}
		{...restProps}
	>
		{@render children?.()}
	</div>
{:else if sidebar.isMobile}
	<Sheet.Root bind:open={() => sidebar.openMobile, (v) => sidebar.setOpenMobile(v)} {...restProps}>
		<Sheet.Content
			data-sidebar="sidebar"
			data-mobile="true"
			class="bg-sidebar text-sidebar-foreground w-[--sidebar-width] p-0 [&>button]:hidden"
			style="--sidebar-width: {SIDEBAR_WIDTH_MOBILE};"
			{side}
		>
			<div class="flex h-full w-full flex-col">
				{@render children?.()}
			</div>
		</Sheet.Content>
	</Sheet.Root>
{:else}
	<div
		bind:this={ref}
		class="text-sidebar-foreground group peer hidden md:block h-screen"
		data-state={sidebar.state}
		data-collapsible={sidebar.state === 'collapsed' ? collapsible : ''}
		data-variant={variant}
		data-side={side}
	>
		<!-- This is what handles the sidebar gap on desktop -->
		<div
			class={cn(
				'relative h-svh w-[--sidebar-width] bg-transparent',
				'group-data-[collapsible=offcanvas]:w-0',
				'group-data-[side=right]:rotate-180',
				variant === 'floating' || variant === 'inset'
					? 'group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)_+_theme(spacing.4))]'
					: 'group-data-[collapsible=icon]:w-[--sidebar-width-icon]'
			)}
		></div>
		<div
			class={cn(
				'fixed inset-y-0 z-10 hidden h-svh w-[--sidebar-width] md:flex',
				side === 'left'
					? 'left-0 group-data-[collapsible=offcanvas]:left-[calc(var(--sidebar-width)*-1)]'
					: 'right-0 group-data-[collapsible=offcanvas]:right-[calc(var(--sidebar-width)*-1)]',
				// Adjust the padding for floating and inset variants.
				variant === 'floating' || variant === 'inset'
					? 'group-data-[collapsible=icon]:w-[calc(var(--sidebar-width-icon)_+_theme(spacing.4)_+2px)]'
					: 'group-data-[collapsible=icon]:w-[--sidebar-width-icon] group-data-[side=left]:border-r group-data-[side=right]:border-l',
				className
			)}
			{...restProps}
		>
			<div
				data-sidebar="sidebar"
				class="bg-sidebar flex h-full w-full flex-col group-data-[variant=floating]:rounded-lg group-data-[variant=floating]:shadow"
			>
				{@render children?.()}
			</div>
			<!-- Resize handle -->
			<div
				role="separator"
				aria-label="Resize sidebar"
				tabindex="-1"
				class="absolute top-0 {side === 'left' ? 'right-0' : 'left-0'} z-20 h-full w-1 cursor-col-resize select-none hover:bg-border active:bg-border transition-colors"
				onmousedown={startResize}
			></div>
		</div>
	</div>
{/if}
