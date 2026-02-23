<script lang="ts">
	import LoginForm from '$lib/components/app/user/login-form.svelte';
	import DatabaseIcon from '@lucide/svelte/icons/between-horizontal-start';
</script>

<div class="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="##" class="flex items-center gap-2 self-center font-medium">
			<div
				class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md"
			>
				<DatabaseIcon class="size-4" />
			</div>
			DBMX
		</a>
		<LoginForm />
	</div>
</div>
