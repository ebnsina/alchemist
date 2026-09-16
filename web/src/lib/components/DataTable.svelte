<script lang="ts" generics="T extends RowData">
	import {
		createTable,
		FlexRender,
		tableFeatures,
		rowPaginationFeature,
		rowSortingFeature,
		type ColumnDef,
		type RowData,
		type SortingState
	} from '@tanstack/svelte-table';
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		Search01Icon,
		ArrowLeft01Icon,
		ArrowRight01Icon,
		ArrowUp01Icon,
		ArrowDown01Icon
	} from '@hugeicons/core-free-icons';

	// The server does the paging, searching and sorting; the table only asks for it.
	// Doing it here would page whatever twenty-five rows happened to arrive.
	const features = tableFeatures({ rowPaginationFeature, rowSortingFeature });

	let {
		columns,
		rows,
		total,
		loading = false,
		page = $bindable(0),
		size = $bindable(10),
		sorting = $bindable([]),
		q = $bindable(''),
		searchLabel = 'Search',
		toolbar,
		empty
	}: {
		// Loose in the feature set: every caller declares plain columns and the table
		// decides which features it runs.
		columns: ColumnDef<any, T>[];
		rows: T[];
		total: number;
		loading?: boolean;
		page?: number;
		size?: number;
		sorting?: SortingState;
		q?: string;
		searchLabel?: string;
		toolbar?: import('svelte').Snippet;
		empty?: import('svelte').Snippet;
	} = $props();

	const table = createTable({
		features,
		get columns() {
			return columns;
		},
		get data() {
			return rows;
		},
		get rowCount() {
			return total;
		},
		manualPagination: true,
		manualSorting: true,
		state: {
			get pagination() {
				return { pageIndex: page, pageSize: size };
			},
			get sorting() {
				return sorting;
			}
		},
		onPaginationChange: (next) => {
			const v = typeof next === 'function' ? next({ pageIndex: page, pageSize: size }) : next;
			page = v.pageIndex;
			size = v.pageSize;
		},
		onSortingChange: (next) => {
			sorting = typeof next === 'function' ? next(sorting) : next;
			// A different order is a different first page; staying put would show page
			// four of an ordering the customer has never seen.
			page = 0;
		}
	});

	const from = $derived(total === 0 ? 0 : page * size + 1);
	const to = $derived(Math.min(total, (page + 1) * size));
	const pages = $derived(Math.max(1, Math.ceil(total / size)));
	const num = (n: number) => new Intl.NumberFormat('en').format(n);

	let typed = $state(q);
	// Typed into, not typed: a request per keystroke would fire six for one word.
	$effect(() => {
		const v = typed;
		const t = setTimeout(() => {
			if (v !== q) {
				q = v;
				page = 0;
			}
		}, 250);
		return () => clearTimeout(t);
	});
</script>

<div class="flex flex-wrap items-center justify-between gap-3">
	<label class="dt-search">
		<HugeiconsIcon icon={Search01Icon} size={15} strokeWidth={1.8} class="flex-none text-faint" />
		<span class="vh">{searchLabel}</span>
		<input bind:value={typed} type="search" placeholder={searchLabel} />
	</label>
	{#if toolbar}
		<div class="flex flex-none flex-wrap items-center gap-2">{@render toolbar()}</div>
	{/if}
</div>

<div class="card mt-4 overflow-x-auto p-0">
	<table class="dt">
		<thead>
			{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
				<tr>
					{#each headerGroup.headers as header (header.id)}
						{@const dir = header.column.getIsSorted()}
						<th scope="col" aria-sort={dir === 'asc' ? 'ascending' : dir === 'desc' ? 'descending' : undefined}>
							{#if !header.isPlaceholder}
								{#if header.column.getCanSort()}
									<button type="button" class="dt-sort" onclick={header.column.getToggleSortingHandler()}>
										<FlexRender {header} />
										{#if dir}
											<HugeiconsIcon
												icon={dir === 'asc' ? ArrowUp01Icon : ArrowDown01Icon}
												size={13}
												strokeWidth={2.4}
											/>
										{/if}
									</button>
								{:else}
									<FlexRender {header} />
								{/if}
							{/if}
						</th>
					{/each}
				</tr>
			{/each}
		</thead>
		<tbody>
			{#if loading && rows.length === 0}
				{#each [0, 1, 2, 3, 4] as i (i)}
					<tr>
						{#each columns as _, c (c)}<td><div class="sk h-4 w-full"></div></td>{/each}
					</tr>
				{/each}
			{:else if rows.length === 0}
				<tr>
					<td colspan={columns.length} class="py-10 text-center">
						{#if empty}{@render empty()}{:else}<p class="sub">Nothing here yet.</p>{/if}
					</td>
				</tr>
			{:else}
				{#each table.getRowModel().rows as row (row.id)}
					<tr>
						{#each row.getAllCells() as cell (cell.id)}
							<td><FlexRender {cell} /></td>
						{/each}
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
</div>

<div class="mt-4 flex flex-wrap items-center justify-between gap-3">
	<label class="flex items-center gap-2">
		<span class="label">Rows</span>
		<select
			class="select w-20"
			value={size}
			onchange={(e) => {
				size = Number(e.currentTarget.value);
				// The row you were looking at is on a different page now; start at the top.
				page = 0;
			}}
		>
			{#each [10, 25, 50, 100] as n (n)}<option value={n}>{n}</option>{/each}
		</select>
	</label>

	<div class="flex flex-wrap items-center gap-3">
		<p class="label" aria-live="polite">
			{#if loading && total === 0}
				Loading
			{:else if total === 0}
				Nothing here
			{:else}
				{num(from)}–{num(to)} of {num(total)}
			{/if}
		</p>
		<div class="flex items-center gap-2">
			<button
				type="button"
				class="icon-btn"
				aria-label="Previous page"
				disabled={page === 0}
				onclick={() => (page -= 1)}
			>
				<HugeiconsIcon icon={ArrowLeft01Icon} size={15} strokeWidth={2.2} />
			</button>
			<button
				type="button"
				class="icon-btn"
				aria-label="Next page"
				disabled={page + 1 >= pages}
				onclick={() => (page += 1)}
			>
				<HugeiconsIcon icon={ArrowRight01Icon} size={15} strokeWidth={2.2} />
			</button>
		</div>
	</div>
</div>

<style>
	.dt-search {
		display: flex;
		align-items: center;
		gap: 8px;
		min-width: 220px;
		flex: 1 1 220px;
		max-width: 340px;
		padding: 9px 12px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-sunk);
		background: var(--color-card);
	}
	.dt-search input {
		flex: 1;
		min-width: 0;
		border: 0;
		background: transparent;
		color: var(--color-ink);
		font-family: var(--font-sans);
		font-size: 13px;
		outline: none;
	}
	.dt {
		width: 100%;
		border-collapse: collapse;
		font-size: 13px;
	}
	.dt :global(th) {
		text-align: left;
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-sunk);
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-dim);
		white-space: nowrap;
	}
	.dt :global(td) {
		padding: 12px 16px;
		border-bottom: 1px solid var(--color-sunk);
		vertical-align: middle;
	}
	.dt :global(tbody tr:last-child td) {
		border-bottom: 0;
	}
	.dt-sort {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		color: inherit;
		font: inherit;
		letter-spacing: inherit;
		text-transform: inherit;
		cursor: pointer;
	}
	.dt-sort:hover {
		color: var(--color-ink);
	}
</style>
