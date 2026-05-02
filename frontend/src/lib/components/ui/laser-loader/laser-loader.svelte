<script lang="ts">
	import { cn } from '$lib/utils.js';

	let { class: className = '' }: { class?: string } = $props();
</script>

<div class={cn('laser-loader-container', className)}>
	<div class="laser-beam"></div>
</div>

<style>
	.laser-loader-container {
		position: relative;
		width: 100%;
		height: 3px;
		overflow: hidden;
		border-radius: 2px;
	}

	.laser-beam {
		position: absolute;
		top: 0;
		left: 0;
		height: 100%;
		width: 40%;
		border-radius: 2px;
		background: linear-gradient(
			90deg,
			transparent 0%,
			rgba(59, 130, 246, 0.5) 20%,
			rgba(99, 102, 241, 0.9) 50%,
			rgba(59, 130, 246, 0.5) 80%,
			transparent 100%
		);
		box-shadow:
			0 0 8px rgba(99, 102, 241, 0.6),
			0 0 16px rgba(99, 102, 241, 0.3);
		animation: laser-sweep 1.2s cubic-bezier(0.4, 0, 0.2, 1) infinite;
	}

	@keyframes laser-sweep {
		0% {
			left: -40%;
		}
		100% {
			left: 100%;
		}
	}
</style>
