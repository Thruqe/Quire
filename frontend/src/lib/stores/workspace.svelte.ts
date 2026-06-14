import { api, type Workspace, type Page } from '$lib/api'

export const workspaceStore = $state({
    workspaces: [] as Workspace[],
    pages: [] as Page[],
    activeWorkspace: null as Workspace | null,
})

export async function loadWorkspaceData() {
    console.log("=== 📡 loadWorkspaceData START ===");
    console.log("Token in localStorage:", localStorage.getItem('token') ? "Present" : "Missing");

    try {
        console.log("📡 Fetching workspaces from /workspaces...");
        const res = await api.get('/workspaces');
        console.log("✅ Workspaces response:", res.data);
        workspaceStore.workspaces = res.data;

        if (workspaceStore.workspaces.length === 0) {
            console.log("🆕 No workspaces found, creating default workspace...");
            const created = await api.post('/workspaces', { name: 'My Workspace' });
            workspaceStore.workspaces = [created.data];
            console.log("✅ Created workspace:", created.data);
        }

        workspaceStore.activeWorkspace = workspaceStore.workspaces[0];
        console.log("🎯 Active workspace:", workspaceStore.activeWorkspace?.id, workspaceStore.activeWorkspace?.name);

        console.log(`📡 Fetching pages for workspace ${workspaceStore.activeWorkspace.id}...`);
        const pagesRes = await api.get(`/workspaces/${workspaceStore.activeWorkspace.id}/pages`);
        workspaceStore.pages = pagesRes.data;
        console.log(`✅ Loaded ${workspaceStore.pages.length} pages`);

        if (workspaceStore.pages.length > 0) {
            console.log("📄 First page:", workspaceStore.pages[0].title);
        }

        console.log("=== ✅ loadWorkspaceData END ===");
        return { success: true, pagesCount: workspaceStore.pages.length };
    } catch (error) {
        console.error("=== ❌ loadWorkspaceData ERROR ===");
        console.error("Error details:", error);
        if (error.response) {
            console.error("Response status:", error.response.status);
            console.error("Response data:", error.response.data);
        }
        throw error;
    }
}

export async function addPage(workspaceId: string, title?: string, parentId?: string) {
    console.log(`📝 Creating page in workspace ${workspaceId} with title:`, title || 'Untitled');
    try {
        const res = await api.post(`/workspaces/${workspaceId}/pages`, {
            title: title || 'Untitled',
            parent_id: parentId || null
        });
        workspaceStore.pages = [...workspaceStore.pages, res.data];
        console.log(`✅ Page created:`, res.data.id, res.data.title);
        return res.data;
    } catch (error) {
        console.error("❌ Failed to create page:", error);
        throw error;
    }
}

export async function removePage(pageId: string) {
    console.log(`🗑️ Removing page:`, pageId);
    try {
        await api.delete(`/pages/${pageId}`);
        workspaceStore.pages = workspaceStore.pages.filter(p => p.id !== pageId);
        console.log(`✅ Page removed:`, pageId);
    } catch (error) {
        console.error("❌ Failed to remove page:", error);
        throw error;
    }
}

export async function updatePage(pageId: string, data: Partial<Page>) {
    console.log(`✏️ Updating page:`, pageId, data);
    try {
        const res = await api.patch(`/pages/${pageId}`, data);
        const index = workspaceStore.pages.findIndex(p => p.id === pageId);
        if (index !== -1) {
            workspaceStore.pages[index] = res.data;
            workspaceStore.pages = [...workspaceStore.pages];
        }
        console.log(`✅ Page updated:`, pageId);
        return res.data;
    } catch (error) {
        console.error("❌ Failed to update page:", error);
        throw error;
    }
}

export function resetWorkspaceStore() {
    console.log("🔄 Resetting workspace store");
    workspaceStore.workspaces = [];
    workspaceStore.pages = [];
    workspaceStore.activeWorkspace = null;
}