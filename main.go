package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tomyl/pg-dump-upsert/pgdump"
)

func main() {
	dsn := flag.String("dsn", "", "Connection string. Example: postgres://user:password@host:5432/db")
	table := flag.String("table", "", "Table to dump.")
	insert := flag.String("insert-columns", "", "Comma-separated list of columns to include in INSERT statement. Defaults to all columns.")
	conflict := flag.String("conflict-column", "", "Append an ON CONFLICT clause for this column. All other columns will be included in a DO UPDATE SET list.")
	noconflict := flag.Bool("noconflict", false, "Append ON CONFLICT DO NOTHING.")
	tx := flag.Bool("tx", false, "Wrap INSERT statements in transaction.")
	query := flag.String("query", "", "Use custom SELECT query. By default fetches all rows. Note that column order must match -insert-columns. It is also valid to just specify a WHERE clause. It will be appended to the default query.")
	verbose := flag.Bool("verbose", false, "Log query statement to stderr.")
	flag.Parse()

	if *dsn == "" {
		log.Fatal("-dsn not supplied")
	}

	if *table == "" {
		log.Fatal("-table not supplied")
	}

	if *noconflict && *conflict != "" {
		log.Fatal("cannot combine -noconflict and -conflict-column")
	}

	var opts pgdump.Options
	opts.Query = strings.TrimSpace(*query)
	opts.ConflictColumn = strings.TrimSpace(*conflict)
	opts.NoConflict = *noconflict
	opts.Verbose = *verbose

	if *insert != "" {
		for _, col := range strings.Split(*insert, ",") {
			col = strings.TrimSpace(col)
			if col != "" {
				opts.InsertColumns = append(opts.InsertColumns, col)
			}
		}
	}

	db, err := sql.Open("postgres", *dsn)

	if err != nil {
		log.Fatal(err)
	}

	if *tx {
		fmt.Printf("BEGIN;\n")
	}

	if err := pgdump.DumpStream(os.Stdout, db, *table, &opts); err != nil {
		log.Fatal(err)
	}

	if *tx {
		fmt.Printf("COMMIT;\n")
	}
}
