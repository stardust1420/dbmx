import { SvelteMap } from 'svelte/reactivity';
import { model } from '$lib/wailsjs/go/models';
import { writable } from 'svelte/store';

import type { ColumnDef, RowData } from '@tanstack/table-core';

// Columns and rows for the query output data table
export let columns = writable<ColumnDef<RowData, unknown>[]>([]);
export let rows = writable<RowData[]>([]);


export let postgresConnectionsMap = $state(new SvelteMap<number, model.Connection>());
export let connectionDatabasesMap = $state(new SvelteMap<number, string[]>());
export let databasesMap = $state(new SvelteMap<string, model.Database>());

export let loadingMap = $state<SvelteMap<number, boolean>>(new SvelteMap<number, boolean>());
export let dbLoadingMap = $state<SvelteMap<string, boolean>>(new SvelteMap<string, boolean>());

// Declare tabsMap as a reactive state variable
export let tabsMap = $state<SvelteMap<number, model.Tab>>(new SvelteMap<number, model.Tab>());

export let selectedDBDisplay = writable('Connect to a database');
export let currentColor = writable('');
export let activePoolID = writable('');


export let selectedQuery = writable('');

export let activeDBs = writable<Array<model.Database>>([]);

export let suggestions = writable<Set<string>>(new Set([
    'SELECT',
    'FROM',
    'WHERE',
    'INNER',
    'LEFT',
    'RIGHT',
    'JOIN',
    'AND',
    'OR',
    'NOT',
    'IN',
    'LIKE',
    'BETWEEN',
    'ORDER BY',
    'LIMIT',
    'IS',
    'true',
    'false',
    'NULL',
    'COUNT()',
    'SUM()',
    'AVG()',
    'MAX()',
    'MIN()',
    'UPDATE',
    'INSERT',
    'DELETE',
    'ALTER',
    'CREATE',
    'DROP',
    'TRUNCATE',
    'RENAME',
    'EXPLAIN',
    'EXPLAIN ANALYZE',
    'EXPLAIN QUERY PLAN',
    'EXPLAIN CACHE',
    'EXPLAIN CONFIG',
    'EXPLAIN DEBUG',
    'EXPLAIN TRIGGER',
    'SET'
]));
