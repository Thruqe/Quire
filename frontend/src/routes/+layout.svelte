<script lang="ts">
	import "../app.css";
	import { auth, logout } from "$lib/stores/auth.svelte";
	import {
		workspaceStore,
		loadWorkspaceData,
		addPage,
		removePage,
		resetWorkspaceStore,
	} from "$lib/stores/workspace.svelte";
	import { goto } from "$app/navigation";
	import { page } from "$app/state";
	import { api, type Page } from "$lib/api";
	import { onMount } from "svelte";
	import {
		FileText,
		Plus,
		Search,
		Upload,
		RefreshCw,
		MoreHorizontal,
		Trash2,
		Copy,
		PenLine,
		ChevronLeft,
		ChevronRight,
	} from "lucide-svelte";

	let { children } = $props();

	let sidebarCollapsed = $state(false);
	let activePageMenu = $state<string | null>(null);
	let searchQuery = $state("");
	let showSearch = $state(false);

	const isAuthPage = $derived(page.url.pathname === "/login");

	const filteredPages = $derived(
		searchQuery
			? workspaceStore.pages.filter((p) =>
					p.title.toLowerCase().includes(searchQuery.toLowerCase()),
				)
			: workspaceStore.pages,
	);

	// DEBUG: Log auth state
	$effect(() => {
		console.log("🔐 Auth state:", {
			token: auth.token ? `${auth.token.substring(0, 20)}...` : null,
			email: auth.email,
			isAuthPage,
			pathname: page.url.pathname
		});
	});

	// DEBUG: Log workspace store state
	$effect(() => {
		console.log("📦 Workspace store:", {
			workspaces: workspaceStore.workspaces.length,
			pages: workspaceStore.pages.length,
			activeWorkspace: workspaceStore.activeWorkspace?.id,
			firstPage: workspaceStore.pages[0]?.title
		});
	});

	$effect(() => {
		console.log("🔄 Checking auth for redirect:", { token: !!auth.token, isAuthPage });
		if (!auth.token && !isAuthPage) {
			console.log("🚫 No token, redirecting to login");
			goto("/login");
		}
	});

	onMount(() => {
		console.log("📱 Component mounted, token:", !!auth.token, "isAuthPage:", isAuthPage);
		if (auth.token && !isAuthPage) {
			console.log("🔄 Calling loadWorkspaceData from onMount");
			loadWorkspaceData().then(() => {
				console.log("✅ loadWorkspaceData completed successfully");
			}).catch(err => {
				console.error("❌ Error loading workspace data:", err);
			});
		} else {
			console.log("⏭️ Skipping loadWorkspaceData - no token or on auth page");
		}
	});

	async function newPage() {
		if (!workspaceStore.activeWorkspace) return;
		console.log("📝 Creating new page in workspace:", workspaceStore.activeWorkspace.id);
		const p = await addPage(workspaceStore.activeWorkspace.id);
		console.log("✅ New page created:", p.id);
		goto(`/pages/${p.id}`);
	}

	async function deletePage(id: string) {
		console.log("🗑️ Deleting page:", id);
		await removePage(id);
		if (page.url.pathname === `/pages/${id}`) goto("/");
		activePageMenu = null;
		console.log("✅ Page deleted");
	}

	async function duplicatePage(p: Page) {
		if (!workspaceStore.activeWorkspace) return;
		console.log("📋 Duplicating page:", p.id);
		await addPage(
			workspaceStore.activeWorkspace.id,
			`${p.title} (copy)`,
			p.parent_id ?? undefined,
		);
		activePageMenu = null;
		console.log("✅ Page duplicated");
	}

	async function refresh() {
		console.log("🔄 Manual refresh triggered");
		resetWorkspaceStore();
		await loadWorkspaceData();
		console.log("✅ Refresh completed");
	}

	function handleClickOutside() {
		activePageMenu = null;
	}
</script>

<svelte:window onclick={handleClickOutside} />

{#if isAuthPage || !auth.token}
	{@render children()}
{:else}
	<div class="flex h-screen bg-white text-gray-900 overflow-hidden">
		<!-- Sidebar -->
		<aside
			class="relative flex flex-col border-r border-gray-200 transition-all duration-200 shrink-0 {sidebarCollapsed
				? 'w-0 border-r-0 overflow-hidden'
				: 'w-60'}"
		>
			<!-- Header -->
			<div
				class="flex items-center justify-between px-4 py-3 border-b border-gray-200 shrink-0"
			>
				<span class="font-semibold text-sm truncate">Quire</span>
			</div>

			<!-- Actions row -->
			<div
				class="flex items-center gap-1 px-2 py-2 border-b border-gray-200 shrink-0"
			>
				<button
					onclick={(e) => {
						e.stopPropagation();
						showSearch = !showSearch;
					}}
					class="sidebar-action-btn"
					title="Search pages"
				>
					<Search size={15} />
				</button>
				<button
					onclick={(e) => {
						e.stopPropagation();
						newPage();
					}}
					class="sidebar-action-btn"
					title="New page"
				>
					<Plus size={15} />
				</button>
				<button class="sidebar-action-btn" title="Upload file">
					<Upload size={15} />
				</button>
				<button onclick={refresh} class="sidebar-action-btn" title="Refresh">
					<RefreshCw size={15} />
				</button>
			</div>

			<!-- Search box -->
			{#if showSearch}
				<div class="px-2 py-2 border-b border-gray-200 shrink-0">
					<input
						type="text"
						placeholder="Search pages..."
						bind:value={searchQuery}
						class="w-full px-2 py-1.5 text-xs border border-gray-200 rounded-md outline-none focus:ring-1 focus:ring-gray-900"
					/>
				</div>
			{/if}

			<!-- Pages list -->
			<nav class="flex-1 overflow-y-auto p-2">
				<div class="mb-1 px-2">
					<span
						class="text-xs font-medium text-gray-400 uppercase tracking-wide"
						>Pages</span
					>
				</div>
				{#each filteredPages as p}
					<div
						class="group relative flex items-center rounded-md hover:bg-gray-100 {page
							.url.pathname === `/pages/${p.id}`
							? 'bg-gray-100'
							: ''}"
					>
						<a
							href="/pages/{p.id}"
							class="flex items-center gap-2 px-2 py-1.5 text-sm text-gray-700 truncate flex-1 min-w-0"
						>
							<FileText size={14} class="shrink-0 text-gray-400" />
							<span class="truncate">{p.title || "Untitled"}</span>
						</a>

						<!-- Three dot menu trigger -->
						<button
							onclick={(e) => {
								e.stopPropagation();
								activePageMenu = activePageMenu === p.id ? null : p.id;
							}}
							class="shrink-0 mr-1 p-0.5 rounded opacity-0 group-hover:opacity-100 hover:bg-gray-200 text-gray-500"
							title="Page options"
						>
							<MoreHorizontal size={14} />
						</button>

						<!-- Dropdown -->
						{#if activePageMenu === p.id}
							<div
								role="menu"
								tabindex="-1"
								class="absolute right-0 top-8 z-50 w-44 bg-white border border-gray-200 rounded-lg shadow-lg py-1"
								onclick={(e) => e.stopPropagation()}
								onkeydown={(e) => {
									if (e.key === "Escape") {
										activePageMenu = null;
									}
								}}
							>
								<button
									onclick={() => {
										goto(`/pages/${p.id}`);
										activePageMenu = null;
									}}
									class="page-menu-item"
									role="menuitem"
								>
									<PenLine size={13} /> Open
								</button>
								<button
									onclick={() => duplicatePage(p)}
									class="page-menu-item"
									role="menuitem"
								>
									<Copy size={13} /> Duplicate
								</button>
								<div class="my-1 border-t border-gray-100"></div>
								<button
									onclick={() => deletePage(p.id)}
									class="page-menu-item text-red-500 hover:bg-red-50"
									role="menuitem"
								>
									<Trash2 size={13} /> Delete
								</button>
							</div>
						{/if}
					</div>
				{/each}

				{#if filteredPages.length === 0}
					<p class="text-xs text-gray-400 px-2 py-2">
						{searchQuery ? "No pages match your search" : "No pages yet"}
					</p>
				{/if}
			</nav>

			<!-- Footer -->
			<div
				class="p-3 border-t border-gray-200 text-xs text-gray-400 flex justify-between items-center shrink-0"
			>
				<span class="truncate">{auth.email}</span>
				<button
					onclick={() => {
						logout();
						resetWorkspaceStore();
						goto("/login");
					}}
					class="hover:text-gray-700 shrink-0 ml-2"
				>
					Sign out
				</button>
			</div>
		</aside>

		<!-- Collapse toggle -->
		<button
			onclick={() => (sidebarCollapsed = !sidebarCollapsed)}
			class="absolute top-1/2 -translate-y-1/2 z-20 flex items-center justify-center w-5 h-10 bg-white border border-gray-200 rounded-r-md shadow-sm hover:bg-gray-50 text-gray-400 hover:text-gray-700 transition-all duration-200 {sidebarCollapsed
				? 'left-0'
				: 'left-60'}"
		>
			{#if sidebarCollapsed}
				<ChevronRight size={12} />
			{:else}
				<ChevronLeft size={12} />
			{/if}
		</button>

		<!-- Main content -->
		<main class="flex-1 overflow-y-auto min-w-0">
			{@render children()}
		</main>
	</div>
{/if}

<style>
	:global(.sidebar-action-btn) {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border-radius: 6px;
		color: #6b7280;
		cursor: pointer;
		border: none;
		background: transparent;
		transition:
			background 0.1s,
			color 0.1s;
	}

	:global(.sidebar-action-btn:hover) {
		background: #f3f4f6;
		color: #111827;
	}

	:global(.page-menu-item) {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 6px 12px;
		font-size: 13px;
		color: #374151;
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
	}

	:global(.page-menu-item:hover) {
		background: #f9fafb;
	}
</style>