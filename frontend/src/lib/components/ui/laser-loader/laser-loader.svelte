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
		height: 2px;
		overflow: hidden;
		border-radius: 1px;
	}

	.laser-beam {
		position: absolute;
		top: 0;
		left: 0;
		height: 100%;
		width: 30%;
		border-radius: 1px;
		background: linear-gradient(
			90deg,
			transparent 0%,
			rgba(0, 255, 220, 0.4) 10%,
			rgba(0, 255, 220, 0.8) 30%,
			#00ffdc 50%,
			rgba(0, 255, 220, 0.8) 70%,
			rgba(0, 255, 220, 0.4) 90%,
			transparent 100%
		);
		box-shadow:
			0 0 6px rgba(0, 255, 220, 0.9),
			0 0 12px rgba(0, 255, 220, 0.7),
			0 0 24px rgba(0, 255, 220, 0.5),
			0 0 48px rgba(0, 255, 220, 0.3);
		animation: laser-sweep 2.4s cubic-bezier(0.4, 0, 0.2, 1) infinite;
	}

	@keyframes laser-sweep {
		0% {
			left: -30%;
		}
		100% {
			left: 100%;
		}
	}
</style>
