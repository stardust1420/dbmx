<script lang="ts">
	import { Tabs as TabsPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		value,
		...restProps
	}: TabsPrimitive.TriggerProps = $props();
</script>

<TabsPrimitive.Trigger
	bind:ref
	class={cn(
		'ring-offset-background h-10 hover:bg-accent focus-visible:ring-ring font-small w-30 items-center justify-center whitespace-nowrap rounded-l-md bg-secondary px-3 py-1 text-sm text-secondary-foreground transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:border-primary data-[state=active]:text-foreground data-[state=active]:shadow',
		className
	)}
	{value}
	{...restProps}
/>
