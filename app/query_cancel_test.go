package app

// Query-cancellation tests. Like the schema tests these need a real server, for a
// reason that is specific to cancellation: the difference between a cancel that
// works and one that only looks like it works is invisible on the client. Closing
// the client socket makes ExecuteQuery return immediately with a cancellation
// error while the postgres backend happily runs the query to completion, holding a
// connection and burning CPU. Only the server's own view — pg_stat_activity —
// distinguishes the two, so that is what these tests assert on.
//
// Same setup as table_schema_test.go; they skip unless DBMX_LIVE_PG names a server.

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"dbmx/model"
)

// backendsRunning counts the backends whose current query carries marker. The
// marker travels as a bind parameter, so this probe never matches itself.
func backendsRunning(t *testing.T, c *Connections, marker string) int {
	t.Helper()
	pool, err := c.poolForTab(liveTabID)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	var n int
	err = pool.QueryRow(context.Background(),
		`SELECT count(*) FROM pg_stat_activity
		  WHERE query LIKE $1 AND pid <> pg_backend_pid() AND state = 'active'`,
		"%"+marker+"%").Scan(&n)
	if err != nil {
		t.Fatalf("pg_stat_activity: %v", err)
	}
	return n
}

// waitForBackends polls until the marked backend count reaches want, so the tests
// never race the server's own bookkeeping.
func waitForBackends(t *testing.T, c *Connections, marker string, want int, within time.Duration) {
	t.Helper()
	deadline := time.Now().Add(within)
	var got int
	for time.Now().Before(deadline) {
		got = backendsRunning(t, c, marker)
		if got == want {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("wanted %d backend(s) running %q within %s, still %d", want, marker, within, got)
}

// cpuBoundQuery keeps a backend busy without ever touching the client socket,
// which is how a real slow query behaves. pg_sleep is no good here: it waits on
// the socket, so postgres tears the backend down the moment the client hangs up
// and a cancel that does nothing still looks like it worked.
const cpuBoundQuery = "SELECT count(*) FROM generate_series(1, 20000000000)"

func TestLiveCancelQueryStopsTheServerSideQuery(t *testing.T) {
	c, cleanup := liveConnections(t)
	defer cleanup()

	// Unique per run: a previous run that failed to clean up must not be able to
	// masquerade as this run's backend.
	marker := fmt.Sprintf("dbmx_cancel_probe_%d", time.Now().UnixNano())
	query := cpuBoundQuery + " /* " + marker + " */"

	// Never leave a runaway backend burning CPU behind a failure. Registered
	// after the pool teardown above so it runs before it: defers unwind LIFO and
	// terminating needs a live pool.
	defer terminateBackends(t, c, marker)

	done := make(chan model.QueryResult, 1)
	go func() { done <- c.ExecuteQuery(liveTabID, query, false) }()

	// Don't cancel until postgres is genuinely executing it.
	waitForBackends(t, c, marker, 1, 10*time.Second)

	if !c.CancelQuery(liveTabID) {
		t.Fatal("CancelQuery reported no running query while one was running")
	}

	select {
	case result := <-done:
		if result.OK {
			t.Errorf("cancelled query reported OK=true, message %q", result.Message)
		}
		if !strings.Contains(strings.ToLower(result.Message), "cancel") {
			t.Errorf("message %q does not describe a cancellation", result.Message)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("ExecuteQuery did not return within 10s of the cancel")
	}

	// The assertion that matters. Closing the client socket makes everything
	// above pass while the backend runs happily on, so the server's own view is
	// the only thing that tells a real cancel from a hangup.
	waitForBackends(t, c, marker, 0, 10*time.Second)
}

// terminateBackends kills any backend left running marker.
func terminateBackends(t *testing.T, c *Connections, marker string) {
	t.Helper()
	pool, err := c.poolForTab(liveTabID)
	if err != nil {
		return
	}
	_, _ = pool.Exec(context.Background(),
		`SELECT pg_terminate_backend(pid) FROM pg_stat_activity
		  WHERE query LIKE $1 AND pid <> pg_backend_pid()`, "%"+marker+"%")
}

func TestLiveCancelQueryWithNothingRunning(t *testing.T) {
	c, cleanup := liveConnections(t)
	defer cleanup()

	if c.CancelQuery(liveTabID) {
		t.Error("CancelQuery reported a cancellation with no query running")
	}
}

func TestLiveCancelQueryStopsATableViewLoad(t *testing.T) {
	c, cleanup := liveConnections(t)
	defer cleanup()

	exec(t, c, `DROP TABLE IF EXISTS cancel_probe`)
	exec(t, c, `CREATE TABLE cancel_probe (id int)`)
	// The table needs rows: postgres never evaluates the filter below against an
	// empty one, and the query the test depends on being slow finishes instantly.
	exec(t, c, `INSERT INTO cancel_probe SELECT generate_series(1, 50)`)
	defer exec(t, c, `DROP TABLE IF EXISTS cancel_probe`)

	marker := fmt.Sprintf("dbmx_cancel_table_%d", time.Now().UnixNano())
	// Correlated on purpose, so it is evaluated per row and the table view's own
	// COUNT(*) never returns.
	where := fmt.Sprintf(
		"id > (SELECT count(*) FROM generate_series(1, 20000000000) g WHERE g > cancel_probe.id) /* %s */",
		marker)

	defer terminateBackends(t, c, marker)

	done := make(chan model.QueryResult, 1)
	go func() {
		done <- c.GetTableData(liveTabID, "cancel_probe", "", "20", "0", where, "", "", false)
	}()

	waitForBackends(t, c, marker, 1, 10*time.Second)

	if !c.CancelQuery(liveTabID) {
		t.Fatal("CancelQuery reported no running query while a table load was running")
	}

	select {
	case result := <-done:
		if result.OK {
			t.Errorf("cancelled table load reported OK=true, message %q", result.Message)
		}
		if !strings.Contains(strings.ToLower(result.Message), "cancel") {
			t.Errorf("message %q does not describe a cancellation", result.Message)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("GetTableData did not return within 10s of the cancel")
	}

	waitForBackends(t, c, marker, 0, 10*time.Second)
}

// An editor query and a table-view load are independent: starting one must not
// be refused because the other is in flight on the same tab.
func TestLiveEditorAndTableViewDoNotBlockEachOther(t *testing.T) {
	c, cleanup := liveConnections(t)
	defer cleanup()

	exec(t, c, `DROP TABLE IF EXISTS cancel_probe_two`)
	exec(t, c, `CREATE TABLE cancel_probe_two (id int)`)
	defer exec(t, c, `DROP TABLE IF EXISTS cancel_probe_two`)

	marker := fmt.Sprintf("dbmx_cancel_both_%d", time.Now().UnixNano())
	defer terminateBackends(t, c, marker)

	editorDone := make(chan model.QueryResult, 1)
	go func() {
		editorDone <- c.ExecuteQuery(liveTabID, cpuBoundQuery+" /* "+marker+" */", false)
	}()
	waitForBackends(t, c, marker, 1, 10*time.Second)

	// The table load must run rather than be turned away by the editor query.
	result := c.GetTableData(liveTabID, "cancel_probe_two", "", "20", "0", "", "", "", false)
	if !result.OK {
		t.Errorf("table load refused while an editor query was running: %q", result.Message)
	}

	if !c.CancelQuery(liveTabID) {
		t.Fatal("CancelQuery found nothing to cancel")
	}
	<-editorDone
	waitForBackends(t, c, marker, 0, 10*time.Second)
}

// Table loads never used to exclude each other, and paging quickly through a
// table fires them back to back. Cancellation must not turn that into a refusal.
func TestLiveTableLoadsDoNotExcludeEachOther(t *testing.T) {
	c, cleanup := liveConnections(t)
	defer cleanup()

	exec(t, c, `DROP TABLE IF EXISTS cancel_probe_page`)
	exec(t, c, `CREATE TABLE cancel_probe_page (id int)`)
	exec(t, c, `INSERT INTO cancel_probe_page SELECT generate_series(1, 50)`)
	defer exec(t, c, `DROP TABLE IF EXISTS cancel_probe_page`)

	marker := fmt.Sprintf("dbmx_cancel_page_%d", time.Now().UnixNano())
	defer terminateBackends(t, c, marker)

	slowWhere := fmt.Sprintf(
		"id > (SELECT count(*) FROM generate_series(1, 20000000000) g WHERE g > cancel_probe_page.id) /* %s */",
		marker)

	slow := make(chan model.QueryResult, 1)
	go func() {
		slow <- c.GetTableData(liveTabID, "cancel_probe_page", "", "20", "0", slowWhere, "", "", false)
	}()
	waitForBackends(t, c, marker, 1, 10*time.Second)

	// A second page load while the first is still in flight must still run.
	result := c.GetTableData(liveTabID, "cancel_probe_page", "", "20", "20", "", "", "", true)
	if !result.OK {
		t.Errorf("second table load was refused: %q", result.Message)
	}

	c.CancelQuery(liveTabID)
	<-slow
	waitForBackends(t, c, marker, 0, 10*time.Second)
}
