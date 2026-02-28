<script lang="ts">
import { onMount } from 'svelte';
import { GetLoggedInUser, Logout } from '$lib/wailsjs/go/app/Auth';
import { goto } from '$app/navigation';
import Label from '$lib/components/ui/label/label.svelte';
import { Button } from '$lib/components/ui/button/index.js';
import { toast } from 'svelte-sonner';
import { ChevronLeft } from 'lucide-svelte';

let fullname = $state("No Account");
let email = $state("No Account");
let avatar = $state("https://api.dicebear.com/9.x/avataaars-neutral/svg?backgroundRotation=0,360");
let isLoggedIn = $state(false);
	
onMount(() => {
    GetLoggedInUser()
        .then((user) => {
            fullname = user.fullname;
            email = user.email;
            avatar = "https://api.dicebear.com/9.x/fun-emoji/svg?seed=Aneka";
            isLoggedIn = true;
        })
        .catch((error) => {
            console.log(error);
        });
});

let logout = () => {
    Logout()
        .then(() => {
            toast.success('Logged out successfully');
            goto('/');
        })
        .catch((error) => {
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
        <Label>{email}</Label>
        <Label>{fullname}</Label>
        {#if isLoggedIn}
            <Button onclick={logout}>Logout</Button>
        {:else}
            <Button onclick={() => goto('/user/login')}>Login</Button>
        {/if}
    </div>
</div>
