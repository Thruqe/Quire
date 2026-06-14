<script lang="ts">
  import { login, register } from '$lib/stores/auth.svelte'
  import { goto } from '$app/navigation'

  let email = $state('')
  let password = $state('')
  let error = $state('')
  let isLogin = $state(true)

  async function submit() {
    error = ''
    try {
      if (isLogin) await login(email, password)
      else await register(email, password)
      goto('/')
    } catch (e: any) {
      error = e.response?.data?.error ?? 'Something went wrong'
    }
  }
</script>

<div class="min-h-screen flex items-center justify-center bg-gray-50">
  <div class="w-full max-w-sm p-8 bg-white rounded-xl shadow-sm border border-gray-200">
    <h1 class="text-xl font-semibold mb-1">Quire</h1>
    <p class="text-sm text-gray-500 mb-6">Your workspace, no noise.</p>

    <form onsubmit={(e) => { e.preventDefault(); submit() }} class="space-y-4">
      <input
        type="email"
        placeholder="Email"
        bind:value={email}
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
      />
      <input
        type="password"
        placeholder="Password"
        bind:value={password}
        class="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-gray-900"
      />
      {#if error}
        <p class="text-red-500 text-xs">{error}</p>
      {/if}
      <button
        type="submit"
        class="w-full py-2 bg-gray-900 text-white rounded-lg text-sm hover:bg-gray-700 transition-colors"
      >
        {isLogin ? 'Sign in' : 'Create account'}
      </button>
    </form>

    <p class="text-center text-xs text-gray-500 mt-4">
      {isLogin ? "Don't have an account?" : 'Already have an account?'}
      <button onclick={() => isLogin = !isLogin} class="underline ml-1">
        {isLogin ? 'Sign up' : 'Sign in'}
      </button>
    </p>
  </div>
</div>