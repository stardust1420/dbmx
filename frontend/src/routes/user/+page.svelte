<script lang="ts">
import { onMount } from 'svelte';
import { GetLoggedInUser, Logout } from '$lib/wailsjs/go/app/Auth';
import { goto } from '$app/navigation';
import Label from '$lib/components/ui/label/label.svelte';
import { Button } from '$lib/components/ui/button/index.js';
import { toast } from 'svelte-sonner';
import { ChevronLeft } from 'lucide-svelte';
import { Spinner } from "$lib/components/ui/spinner/index.js";
import { userAvatar, userEmail, userFullName, isLoggedIn, userUseDefaultKey, userIsAIEnabled } from '$lib/state.svelte';

let logoutLoading = $state(false);

let logout = () => {
    logoutLoading = true
    Logout()
        .then(() => {
            $isLoggedIn = false
            $userFullName = 'No Account'
            $userEmail= 'No Account'
            $userAvatar = 'https://api.dicebear.com/9.x/avataaars-neutral/svg?backgroundRotation=0,360'
            $userIsAIEnabled = false
            $userUseDefaultKey = false
            logoutLoading = false
            toast.success('Logged out successfully');
            goto('/');
        })
        .catch((error) => {
            logoutLoading = false
            toast.error('Failed to log out', {
                description: error,
                action: {
                    label: 'OK',
                    onClick: () => console.info('OK')
                }
            });
        });
}

</script>

<div class="flex h-full w-full items-center justify-center">
    <div class="flex flex-col gap-4 w-96">
        <a class="" href="/">
            <ChevronLeft size={32} />
        </a>
        <Label>{$userEmail}</Label>
        <Label>{$userFullName}</Label>
        {#if $isLoggedIn}
			{#if logoutLoading}
                <Button variant="outline" disabled>
                    <Spinner />
                </Button>
			{:else}
                <Button onclick={logout}>Logout</Button>
			{/if}
        {:else}
            <Button onclick={() => goto('/user/login')}>Login</Button>
        {/if}
    </div>
</div>
