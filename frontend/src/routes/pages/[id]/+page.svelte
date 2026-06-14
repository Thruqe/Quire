<script lang="ts">
	import { api, type Page } from "$lib/api";
	import { page } from "$app/state";
	import { onDestroy } from "svelte";
	import { Editor } from "@tiptap/core";
	import StarterKit from "@tiptap/starter-kit";
	import Placeholder from "@tiptap/extension-placeholder";
	import Typography from "@tiptap/extension-typography";
	import Underline from "@tiptap/extension-underline";
	import TextAlign from "@tiptap/extension-text-align";
	import Highlight from "@tiptap/extension-highlight";
	import { Table } from "@tiptap/extension-table";
	import { TableRow } from "@tiptap/extension-table-row";
	import { TableCell } from "@tiptap/extension-table-cell";
	import { TableHeader } from "@tiptap/extension-table-header";
	import { marked } from "marked";
	import {
		Bold,
		Italic,
		Underline as UnderlineIcon,
		Strikethrough,
		Highlighter,
		AlignLeft,
		AlignCenter,
		AlignRight,
		List,
		ListOrdered,
		Quote,
		Code,
		Code2,
		Heading1,
		Heading2,
		Heading3,
		Undo,
		Redo,
		Table as TableIcon,
	} from "lucide-svelte";
	import {
		workspaceStore,
		loadWorkspaceData,
	} from "$lib/stores/workspace.svelte";

	marked.setOptions({
		breaks: true,
		gfm: true,
		pedantic: false,
	});

	let pageData = $state<Page | null>(null);
	let loading = $state(true);
	let error = $state("");
	let editorEl = $state<HTMLElement | null>(null);
	let editor = $state<Editor | null>(null);
	let saveTimeout: ReturnType<typeof setTimeout> | null = null;
	let title = $state("");
	let saving = $state(false);
	let isLoading = $state(false);
	let previousPageId = $state<string | null>(null);

	// Track active menu state
	let activeMenus = $state({
		format: false,
		align: false,
		list: false,
		table: false,
	});

	// Reactive state for toolbar button active states
	let isBold = $state(false);
	let isItalic = $state(false);
	let isUnderline = $state(false);
	let isStrike = $state(false);
	let isHighlight = $state(false);
	let isCode = $state(false);
	let isHeading1 = $state(false);
	let isHeading2 = $state(false);
	let isHeading3 = $state(false);
	let isParagraph = $state(false);
	let isAlignLeft = $state(false);
	let isAlignCenter = $state(false);
	let isAlignRight = $state(false);
	let isBulletList = $state(false);
	let isOrderedList = $state(false);
	let isBlockquote = $state(false);
	let isCodeBlock = $state(false);

	// Close all dropdowns
	function closeAllDropdowns() {
		activeMenus.format = false;
		activeMenus.align = false;
		activeMenus.list = false;
		activeMenus.table = false;
	}

	// Toggle dropdown with click
	function toggleDropdown(dropdown: keyof typeof activeMenus) {
		// Close all other dropdowns first
		if (!activeMenus[dropdown]) {
			closeAllDropdowns();
		}
		activeMenus[dropdown] = !activeMenus[dropdown];
	}

	// Destroy editor instance
	function destroyEditor() {
		if (saveTimeout) {
			clearTimeout(saveTimeout);
			saveTimeout = null;
		}
		if (editor) {
			editor.destroy();
			editor = null;
		}
	}

	// Process pasted content
	async function processPastedContent(text: string): Promise<string> {
		try {
			const html = await marked.parse(text);
			return html;
		} catch (err) {
			console.error("Failed to parse markdown:", err);
			return `<pre><code>${escapeHtml(text)}</code></pre>`;
		}
	}

	function escapeHtml(text: string): string {
		return text
			.replace(/&/g, "&amp;")
			.replace(/</g, "&lt;")
			.replace(/>/g, "&gt;")
			.replace(/"/g, "&quot;")
			.replace(/'/g, "&#39;");
	}

	// Initialize editor
	function initEditorInstance(content: string) {
		if (!editorEl) return;

		destroyEditor();

		const newEditor = new Editor({
			element: editorEl,
			extensions: [
				StarterKit.configure({
					codeBlock: {
						HTMLAttributes: {
							class:
								"rounded-lg bg-gray-900 text-white p-4 font-mono text-sm overflow-x-auto",
						},
					},
				}),
				Table.configure({
					resizable: true,
					HTMLAttributes: {
						class: "min-w-full border-collapse border border-gray-200",
					},
				}),
				TableRow,
				TableCell,
				TableHeader,
				Placeholder.configure({
					placeholder:
						"Start writing your content here...\n\nSupports Markdown: # Headers, **bold**, *italic*, - lists, ```code blocks```, | tables |, and more!",
				}),
				Typography,
				Underline,
				TextAlign.configure({ types: ["heading", "paragraph"] }),
				Highlight,
			],
			content: content || "<p></p>",
			editorProps: {
				attributes: {
					class:
						"prose prose-gray max-w-none outline-none min-h-[400px] focus:outline-none",
					spellcheck: "false",
				},
				handlePaste: (view, event) => {
					const text = event.clipboardData?.getData("text/plain");
					const html = event.clipboardData?.getData("text/html");

					if (html && html.trim() && !text?.match(/^[\s\S]*[#*`\->[\]()|]/)) {
						return false;
					}

					if (
						text &&
						(text.includes("# ") ||
							text.includes("## ") ||
							text.includes("### ") ||
							text.includes("- ") ||
							text.includes("* ") ||
							text.includes("**") ||
							text.includes("__") ||
							text.includes("```") ||
							text.includes("|") ||
							(text.includes("[") && text.includes("](")))
					) {
						event.preventDefault();
						processPastedContent(text).then((parsedHtml) => {
							newEditor?.commands.insertContent(parsedHtml);
						});
						return true;
					}

					return false;
				},
				transformPastedHTML: (html) => {
					return html
						.replace(/<span[^>]*style="[^"]*"[^>]*>(.*?)<\/span>/gi, "$1")
						.replace(/<font[^>]*>(.*?)<\/font>/gi, "$1")
						.replace(/<meta[^>]*>/gi, "")
						.replace(/<style[^>]*>.*?<\/style>/gi, "")
						.replace(/<script[^>]*>.*?<\/script>/gi, "");
				},
			},
			onUpdate: () => {
				if (saveTimeout) clearTimeout(saveTimeout);
				saveTimeout = setTimeout(saveContent, 1000);
				updateToolbarStates();
			},
			onSelectionUpdate: () => {
				updateToolbarStates();
			},
			onTransaction: () => {
				updateToolbarStates();
			},
		});
		editor = newEditor;
		updateToolbarStates();
	}

	// Update toolbar button states based on editor selection
	function updateToolbarStates() {
		if (!editor) return;

		isBold = editor.isActive("bold");
		isItalic = editor.isActive("italic");
		isUnderline = editor.isActive("underline");
		isStrike = editor.isActive("strike");
		isHighlight = editor.isActive("highlight");
		isCode = editor.isActive("code");
		isHeading1 = editor.isActive("heading", { level: 1 });
		isHeading2 = editor.isActive("heading", { level: 2 });
		isHeading3 = editor.isActive("heading", { level: 3 });
		isParagraph = editor.isActive("paragraph");
		isAlignLeft = editor.isActive({ textAlign: "left" });
		isAlignCenter = editor.isActive({ textAlign: "center" });
		isAlignRight = editor.isActive({ textAlign: "right" });
		isBulletList = editor.isActive("bulletList");
		isOrderedList = editor.isActive("orderedList");
		isBlockquote = editor.isActive("blockquote");
		isCodeBlock = editor.isActive("codeBlock");
	}

	// Load page data
	async function loadPageData(id: string) {
		if (isLoading) return;

		isLoading = true;
		loading = true;
		error = "";

		try {
			console.log("Loading page:", id);

			if (workspaceStore.workspaces.length === 0) {
				await loadWorkspaceData();
			}

			const [pageRes, contentRes] = await Promise.all([
				api.get(`/pages/${id}`),
				api.get(`/pages/${id}/content`),
			]);

			pageData = pageRes.data;
			title = pageData!.title;
			const content = contentRes.data.content || "<p></p>";

			setTimeout(() => {
				if (editorEl) {
					initEditorInstance(content);
				}
			}, 50);
		} catch (err: any) {
			console.error("Failed to load page:", err);
			if (err.response?.status === 404) {
				error = "Page not found";
			} else if (err.message?.includes("timeout")) {
				error = "Connection timeout. Make sure backend is running.";
			} else {
				error = "Failed to load page. Please try again.";
			}
		} finally {
			loading = false;
			isLoading = false;
		}
	}

	// Watch for page ID changes
	$effect(() => {
		const currentId = page.params.id;

		if (currentId && currentId !== previousPageId && !isLoading) {
			console.log("Page ID changed from", previousPageId, "to", currentId);
			previousPageId = currentId;
			destroyEditor();
			loadPageData(currentId);
		}
	});

	// Save content with better error handling
	async function saveContent() {
		if (!pageData || !editor) return;

		saving = true;

		try {
			let content = editor.getHTML();
			content = content.replace(/[\u0000-\u001F\u007F-\u009F]/g, "");
			await api.post(`/pages/${pageData.id}/content`, {
				content: content,
			});
			console.log("Content saved");
		} catch (err: any) {
			console.error("Failed to save:", err);
		} finally {
			saving = false;
		}
	}

	// Save title
	async function saveTitle() {
		if (!pageData) return;
		if (title === pageData.title) return;

		try {
			await api.patch(`/pages/${pageData.id}`, { title });
			pageData = { ...pageData, title };
			console.log("Title saved");
		} catch (err) {
			console.error("Failed to save title:", err);
		}
	}

	// Insert table
	function insertTable() {
		if (!editor) return;
		editor
			.chain()
			.focus()
			.insertTable({ rows: 3, cols: 3, withHeaderRow: true })
			.run();
		closeAllDropdowns();
	}

	// Toolbar actions with focus feedback
	function toolbarBtn(action: () => void) {
		return (e: MouseEvent) => {
			e.preventDefault();
			action();
			editor?.commands.focus();
			setTimeout(() => updateToolbarStates(), 10);

			const target = e.currentTarget as HTMLElement;
			if (target) {
				target.classList.add("toolbar-btn-clicked");
				setTimeout(() => {
					target.classList.remove("toolbar-btn-clicked");
				}, 200);
			}
		};
	}

	// Handle click outside to close dropdowns
	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (!target.closest(".dropdown-container")) {
			closeAllDropdowns();
		}
	}

	onDestroy(() => {
		destroyEditor();
		document.removeEventListener("click", handleClickOutside);
	});
</script>

<svelte:window onclick={handleClickOutside} />

{#if loading}
	<div class="flex items-center justify-center h-full">
		<div class="text-gray-400 text-sm">Loading page...</div>
	</div>
{:else if error}
	<div class="flex flex-col items-center justify-center h-full gap-4">
		<div class="text-red-500 text-sm">{error}</div>
		<button
			onclick={() => (window.location.href = "/")}
			class="px-4 py-2 text-sm bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
		>
			Go to Dashboard
		</button>
	</div>
{:else}
	<div class="flex flex-col h-full bg-white">
		<!-- Title bar -->
		<div class="border-b border-gray-200 px-6 py-4 bg-white sticky top-0 z-10">
			<input
				type="text"
				bind:value={title}
				onblur={saveTitle}
				placeholder="Untitled"
				class="text-3xl font-bold text-gray-900 outline-none bg-transparent w-full focus:ring-0 px-0 border-0"
				spellcheck="false"
			/>
		</div>

		<!-- Toolbar -->
		<div
			class="border-b border-gray-200 px-4 py-2 flex items-center gap-0.5 flex-wrap bg-white sticky top-0 z-10 shadow-sm"
		>
			<button
				onmousedown={toolbarBtn(() => editor?.chain().focus().undo().run())}
				class="toolbar-btn"
				title="Undo (⌘Z)"
			>
				<Undo size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() => editor?.chain().focus().redo().run())}
				class="toolbar-btn"
				title="Redo (⌘⇧Z)"
			>
				<Redo size={16} />
			</button>

			<div class="toolbar-divider"></div>

			<!-- Format Dropdown -->
			<div class="relative dropdown-container">
				<button
					onclick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						toggleDropdown("format");
					}}
					class="toolbar-btn {activeMenus.format ? 'active' : ''}"
					title="Format"
				>
					<Heading1 size={16} />
				</button>
				{#if activeMenus.format}
					<div class="toolbar-dropdown" role="menu">
						<button
							onclick={() => {
								editor?.chain().focus().toggleHeading({ level: 1 }).run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isHeading1 ? 'active' : ''}"
							role="menuitem"
						>
							<Heading1 size={14} /> Heading 1
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().toggleHeading({ level: 2 }).run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isHeading2 ? 'active' : ''}"
							role="menuitem"
						>
							<Heading2 size={14} /> Heading 2
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().toggleHeading({ level: 3 }).run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isHeading3 ? 'active' : ''}"
							role="menuitem"
						>
							<Heading3 size={14} /> Heading 3
						</button>
						<div class="dropdown-divider"></div>
						<button
							onclick={() => {
								editor?.chain().focus().setParagraph().run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isParagraph ? 'active' : ''}"
							role="menuitem"
						>
							Paragraph
						</button>
					</div>
				{/if}
			</div>

			<div class="toolbar-divider"></div>

			<!-- Text Formatting Buttons -->
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleBold().run(),
				)}
				class="toolbar-btn {isBold ? 'active' : ''}"
				title="Bold (⌘B or **text**)"
			>
				<Bold size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleItalic().run(),
				)}
				class="toolbar-btn {isItalic ? 'active' : ''}"
				title="Italic (⌘I or *text*)"
			>
				<Italic size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleUnderline().run(),
				)}
				class="toolbar-btn {isUnderline ? 'active' : ''}"
				title="Underline (⌘U)"
			>
				<UnderlineIcon size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleStrike().run(),
				)}
				class="toolbar-btn {isStrike ? 'active' : ''}"
				title="Strikethrough (~~text~~)"
			>
				<Strikethrough size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleHighlight().run(),
				)}
				class="toolbar-btn {isHighlight ? 'active' : ''}"
				title="Highlight"
			>
				<Highlighter size={16} />
			</button>
			<button
				onmousedown={toolbarBtn(() =>
					editor?.chain().focus().toggleCode().run(),
				)}
				class="toolbar-btn {isCode ? 'active' : ''}"
				title="Inline Code (`code`)"
			>
				<Code size={16} />
			</button>

			<div class="toolbar-divider"></div>

			<!-- Alignment Dropdown -->
			<div class="relative dropdown-container">
				<button
					onclick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						toggleDropdown("align");
					}}
					class="toolbar-btn {activeMenus.align ? 'active' : ''}"
					title="Alignment"
				>
					<AlignLeft size={16} />
				</button>
				{#if activeMenus.align}
					<div class="toolbar-dropdown" role="menu">
						<button
							onclick={() => {
								editor?.chain().focus().setTextAlign("left").run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isAlignLeft ? 'active' : ''}"
							role="menuitem"
						>
							<AlignLeft size={14} /> Align Left
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().setTextAlign("center").run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isAlignCenter ? 'active' : ''}"
							role="menuitem"
						>
							<AlignCenter size={14} /> Align Center
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().setTextAlign("right").run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isAlignRight ? 'active' : ''}"
							role="menuitem"
						>
							<AlignRight size={14} /> Align Right
						</button>
					</div>
				{/if}
			</div>

			<div class="toolbar-divider"></div>

			<!-- Lists Dropdown -->
			<div class="relative dropdown-container">
				<button
					onclick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						toggleDropdown("list");
					}}
					class="toolbar-btn {activeMenus.list ? 'active' : ''}"
					title="Lists"
				>
					<List size={16} />
				</button>
				{#if activeMenus.list}
					<div class="toolbar-dropdown" role="menu">
						<button
							onclick={() => {
								editor?.chain().focus().toggleBulletList().run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isBulletList ? 'active' : ''}"
							role="menuitem"
						>
							<List size={14} /> Bullet List
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().toggleOrderedList().run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isOrderedList ? 'active' : ''}"
							role="menuitem"
						>
							<ListOrdered size={14} /> Numbered List
						</button>
						<div class="dropdown-divider"></div>
						<button
							onclick={() => {
								editor?.chain().focus().toggleBlockquote().run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isBlockquote ? 'active' : ''}"
							role="menuitem"
						>
							<Quote size={14} /> Quote
						</button>
						<button
							onclick={() => {
								editor?.chain().focus().toggleCodeBlock().run();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item {isCodeBlock ? 'active' : ''}"
							role="menuitem"
						>
							<Code2 size={14} /> Code Block
						</button>
					</div>
				{/if}
			</div>

			<div class="toolbar-divider"></div>

			<!-- Table Dropdown -->
			<div class="relative dropdown-container">
				<button
					onclick={(e) => {
						e.preventDefault();
						e.stopPropagation();
						toggleDropdown("table");
					}}
					class="toolbar-btn {activeMenus.table ? 'active' : ''}"
					title="Insert Table"
				>
					<TableIcon size={16} />
				</button>
				{#if activeMenus.table}
					<div class="toolbar-dropdown" role="menu">
						<button
							onclick={() => {
								insertTable();
								closeAllDropdowns();
							}}
							class="toolbar-dropdown-item"
							role="menuitem"
						>
							<TableIcon size={14} /> Insert 3x3 Table
						</button>
					</div>
				{/if}
			</div>

			<div class="ml-auto text-xs text-gray-400">
				{#if saving}
					<span class="animate-pulse">Saving...</span>
				{/if}
			</div>
		</div>

		<!-- Editor Content Area -->
		<div class="flex-1 overflow-y-auto bg-gray-50">
			<div
				class="max-w-4xl mx-auto bg-white shadow-sm min-h-full my-8 rounded-lg"
			>
				<div class="p-8">
					<div
						bind:this={editorEl}
						class="editor-container min-h-125 cursor-text"
						style="outline: none;"
					></div>
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(.toolbar-btn) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 6px;
		color: #6b7280;
		cursor: pointer;
		border: none;
		background: transparent;
		transition: all 0.15s ease;
		position: relative;
	}

	:global(.toolbar-btn:hover) {
		background: #f3f4f6;
		color: #111827;
	}

	:global(.toolbar-btn.active) {
		background: #111827;
		color: #ffffff;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
	}

	:global(.toolbar-btn-clicked) {
		transform: scale(0.95);
		background: #e5e7eb;
	}

	:global(.toolbar-divider) {
		width: 1px;
		height: 24px;
		background: #e5e7eb;
		margin: 0 4px;
	}

	:global(.toolbar-dropdown) {
		position: absolute;
		top: 100%;
		left: 0;
		margin-top: 4px;
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		box-shadow:
			0 4px 6px -1px rgba(0, 0, 0, 0.1),
			0 2px 4px -1px rgba(0, 0, 0, 0.06);
		min-width: 160px;
		z-index: 50;
		padding: 4px 0;
	}

	:global(.toolbar-dropdown-item) {
		display: flex;
		align-items: center;
		gap: 8px;
		width: 100%;
		padding: 8px 12px;
		font-size: 13px;
		color: #374151;
		background: transparent;
		border: none;
		cursor: pointer;
		text-align: left;
		transition: all 0.1s ease;
	}

	:global(.toolbar-dropdown-item:hover) {
		background: #f3f4f6;
		color: #111827;
	}

	:global(.toolbar-dropdown-item.active) {
		background: #111827;
		color: #ffffff;
	}

	:global(.dropdown-divider) {
		height: 1px;
		background: #e5e7eb;
		margin: 4px 0;
	}

	:global(.editor-container .ProseMirror) {
		outline: none;
		min-height: 500px;
	}

	:global(
			.editor-container .ProseMirror p.is-editor-empty:first-child::before
		) {
		content: attr(data-placeholder);
		color: #9ca3af;
		pointer-events: none;
		float: left;
		height: 0;
	}

	:global(.editor-container .ProseMirror:focus) {
		outline: none;
	}

	:global(.ProseMirror) {
		outline: none;
	}

	:global(.ProseMirror pre) {
		background: #1f2937;
		color: #f3f4f6;
		padding: 16px;
		border-radius: 8px;
		font-family: "Courier New", monospace;
		font-size: 14px;
		overflow-x: auto;
	}

	:global(.ProseMirror code) {
		background: #f3f4f6;
		padding: 2px 4px;
		border-radius: 4px;
		font-family: "Courier New", monospace;
		font-size: 0.9em;
		color: #ef4444;
	}

	:global(.ProseMirror pre code) {
		background: transparent;
		color: inherit;
		padding: 0;
	}

	:global(.ProseMirror blockquote) {
		border-left: 4px solid #e5e7eb;
		margin: 1rem 0;
		padding-left: 1rem;
		color: #6b7280;
		font-style: italic;
	}

	:global(.ProseMirror a) {
		color: #3b82f6;
		text-decoration: underline;
		cursor: pointer;
	}

	:global(.ProseMirror ul, .ProseMirror ol) {
		padding-left: 1.5rem;
	}

	:global(.ProseMirror h1) {
		font-size: 2.5rem;
		font-weight: bold;
		margin-top: 2rem;
		margin-bottom: 1rem;
		border-bottom: 2px solid #e5e7eb;
		padding-bottom: 0.5rem;
	}

	:global(.ProseMirror h2) {
		font-size: 2rem;
		font-weight: bold;
		margin-top: 1.75rem;
		margin-bottom: 0.75rem;
		border-bottom: 1px solid #e5e7eb;
		padding-bottom: 0.25rem;
	}

	:global(.ProseMirror h3) {
		font-size: 1.5rem;
		font-weight: bold;
		margin-top: 1.5rem;
		margin-bottom: 0.5rem;
	}

	:global(.ProseMirror p) {
		margin-bottom: 1rem;
		line-height: 1.6;
	}

	:global(.ProseMirror table) {
		border-collapse: collapse;
		width: 100%;
		margin: 1rem 0;
		display: table;
	}

	:global(.ProseMirror th) {
		background: #f3f4f6;
		font-weight: bold;
		padding: 10px 12px;
		border: 1px solid #e5e7eb;
		text-align: left;
	}

	:global(.ProseMirror td) {
		padding: 8px 12px;
		border: 1px solid #e5e7eb;
	}

	:global(.ProseMirror tr:nth-child(even)) {
		background: #f9fafb;
	}
</style>
