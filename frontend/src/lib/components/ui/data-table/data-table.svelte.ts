import {
	type RowData,
	type TableOptions,
	type TableOptionsResolved,
	type TableState,
	createTable,
} from "@tanstack/table-core";
import BinaryIcon from "lucide-svelte/icons/binary";
import BracesIcon from "lucide-svelte/icons/braces";
import BracketsIcon from "lucide-svelte/icons/brackets";
import CalendarIcon from "lucide-svelte/icons/calendar";
import CalendarClockIcon from "lucide-svelte/icons/calendar-clock";
import CircleHelpIcon from "lucide-svelte/icons/circle-help";
import ClockIcon from "lucide-svelte/icons/clock";
import FingerprintIcon from "lucide-svelte/icons/fingerprint";
import HashIcon from "lucide-svelte/icons/hash";
import ListIcon from "lucide-svelte/icons/list";
import NetworkIcon from "lucide-svelte/icons/network";
import RulerIcon from "lucide-svelte/icons/ruler";
import ToggleLeftIcon from "lucide-svelte/icons/toggle-left";
import TypeIcon from "lucide-svelte/icons/type";
import type { model } from "$lib/wailsjs/go/models";

/**
 * lucide-svelte 0.469 still ships its icons as legacy class components, so the
 * icon map is typed off one of them rather than off Svelte 5's `Component`.
 * Every icon in the package has the same props, so any of them will do.
 */
export type IconComponent = typeof CircleHelpIcon;

/**
 * Creates a reactive TanStack table object for Svelte.
 * @param options Table options to create the table with.
 * @returns A reactive table object.
 * @example
 * ```svelte
 * <script>
 *   const table = createSvelteTable({ ... })
 * </script>
 *
 * <table>
 *   <thead>
 *     {#each table.getHeaderGroups() as headerGroup}
 *       <tr>
 *         {#each headerGroup.headers as header}
 *           <th colspan={header.colSpan}>
 *         	   <FlexRender content={header.column.columnDef.header} context={header.getContext()} />
 *         	 </th>
 *         {/each}
 *       </tr>
 *     {/each}
 *   </thead>
 * 	 <!-- ... -->
 * </table>
 * ```
 */
export function createSvelteTable<TData extends RowData>(options: TableOptions<TData>) {
	const resolvedOptions: TableOptionsResolved<TData> = mergeObjects(
		{
			state: {},
			onStateChange() {},
			renderFallbackValue: null,
			mergeOptions: (
				defaultOptions: TableOptions<TData>,
				options: Partial<TableOptions<TData>>
			) => {
				return mergeObjects(defaultOptions, options);
			},
		},
		options
	);

	const table = createTable(resolvedOptions);
	let state = $state<Partial<TableState>>(table.initialState);

	function updateOptions() {
		table.setOptions((prev) => {
			return mergeObjects(prev, options, {
				state: mergeObjects(state, options.state || {}),

				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				onStateChange: (updater: any) => {
					if (updater instanceof Function) state = updater(state);
					else state = mergeObjects(state, updater);

					options.onStateChange?.(updater);
				},
			});
		});
	}

	updateOptions();

	$effect.pre(() => {
		updateOptions();
	});

	return table;
}

type MaybeThunk<T extends object> = T | (() => T | null | undefined);
type Intersection<T extends readonly unknown[]> = (T extends [infer H, ...infer R]
	? H & Intersection<R>
	: unknown) & {};

/**
 * Lazily merges several objects (or thunks) while preserving
 * getter semantics from every source.
 *
 * Proxy-based to avoid known WebKit recursion issue.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function mergeObjects<Sources extends readonly MaybeThunk<any>[]>(
	...sources: Sources
): Intersection<{ [K in keyof Sources]: Sources[K] }> {
	const resolve = <T extends object>(src: MaybeThunk<T>): T | undefined =>
		typeof src === "function" ? (src() ?? undefined) : src;

	const findSourceWithKey = (key: PropertyKey) => {
		for (let i = sources.length - 1; i >= 0; i--) {
			const obj = resolve(sources[i]);
			if (obj && key in obj) return obj;
		}
		return undefined;
	};

	return new Proxy(Object.create(null), {
		get(_, key) {
			const src = findSourceWithKey(key);

			return src?.[key as never];
		},

		has(_, key) {
			return !!findSourceWithKey(key);
		},

		ownKeys(): (string | symbol)[] {
			// eslint-disable-next-line svelte/prefer-svelte-reactivity
			const all = new Set<string | symbol>();
			for (const s of sources) {
				const obj = resolve(s);
				if (obj) {
					for (const k of Reflect.ownKeys(obj) as (string | symbol)[]) {
						all.add(k);
					}
				}
			}
			return [...all];
		},

		getOwnPropertyDescriptor(_, key) {
			const src = findSourceWithKey(key);
			if (!src) return undefined;
			return {
				configurable: true,
				enumerable: true,
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				value: (src as any)[key],
				writable: true,
			};
		},
	}) as Intersection<{ [K in keyof Sources]: Sources[K] }>;
}

/**
 * Column type metadata.
 *
 * A result-set column carries the postgres type it came from, which the grid
 * draws as an icon beside the column name. It rides on TanStack's `meta` rather
 * than on the header itself, because the header holds the column's name as a
 * plain string and both grids read it back from there to address a cell in an
 * UPDATE.
 */
declare module "@tanstack/table-core" {
	// eslint-disable-next-line @typescript-eslint/no-unused-vars
	interface ColumnMeta<TData extends RowData, TValue> {
		columnType?: model.ColumnType;
	}
}

/**
 * Builds the `meta` for one column of a query result. `columnTypes` is parallel
 * to the result's `columns`, but a result can arrive without it -- an error row,
 * the "Rows Affected" of a write, a tab restored from before types were saved --
 * so the column index is looked up defensively and the icon simply goes missing.
 */
export function columnTypeMeta(
	columnTypes: model.ColumnType[] | undefined,
	index: number
): { columnType?: model.ColumnType } {
	return { columnType: columnTypes?.[index] };
}

/**
 * Icons are chosen by type name first and by `pg_type.typcategory` second, so a
 * type this list has never heard of -- a domain, an enum, an extension's own type
 * -- still lands on the icon its family deserves instead of a question mark.
 */
const iconByTypeName: Record<string, IconComponent> = {
	boolean: ToggleLeftIcon,
	uuid: FingerprintIcon,
	json: BracesIcon,
	jsonb: BracesIcon,
	bytea: BinaryIcon,
	date: CalendarIcon,
	interval: ClockIcon,
	"time without time zone": ClockIcon,
	"time with time zone": ClockIcon,
};

const iconByCategory: Record<string, IconComponent> = {
	N: HashIcon, // numeric
	S: TypeIcon, // string
	D: CalendarClockIcon, // date/time
	B: ToggleLeftIcon, // boolean
	E: ListIcon, // enum
	A: BracketsIcon, // array
	I: NetworkIcon, // network address
	T: ClockIcon, // timespan
	V: BinaryIcon, // bit-string
	R: RulerIcon, // range
	C: BracesIcon, // composite
};

/** The icon for a column's type, or the fallback when the type is unknown. */
export function columnTypeIcon(columnType?: model.ColumnType): IconComponent {
	if (!columnType?.dataType) return CircleHelpIcon;
	return (
		iconByTypeName[columnType.dataType] ??
		iconByCategory[columnType.category] ??
		CircleHelpIcon
	);
}

/** The type as postgres prints it, for the header's tooltip. */
export function columnTypeLabel(columnType?: model.ColumnType): string {
	return columnType?.dataType || "unknown type";
}

/**
 * Whether a column holds JSON, which the grid cannot show usefully on the one
 * line a cell gives it. Those cells open the cell editor on double click
 * instead of the inline input, so the value can be read, formatted, copied and
 * edited at a readable size.
 *
 * Deliberately json and jsonb only. Every other type -- including long text and
 * arrays -- keeps the inline input it has always had, so double clicking a cell
 * you already edit in place does not suddenly open a dialog.
 */
export function isJsonColumn(columnType?: model.ColumnType): boolean {
	return columnType?.dataType === "json" || columnType?.dataType === "jsonb";
}
