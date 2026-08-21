package reader

const (
	// jopTypeNote and jopTypeNotebook are items.jop_type values. 1 = note,
	// 2 = notebook/folder.
	jopTypeNote     = 1
	jopTypeNotebook = 2
)

// expectedColumns are the items columns reader depends on, and the
// Postgres data_type information_schema reports for each. Notably, title
// and deleted_time are NOT here: despite how the brief originally
// described the schema, both live only inside the JSON envelope in
// content, never as row-level columns.
var expectedColumns = map[string]string{
	"jop_id":                 "character varying",
	"jop_parent_id":          "character varying",
	"jop_type":               "integer",
	"jop_encryption_applied": "integer",
	"owner_id":               "character varying",
	"content":                "bytea",
}

const schemaCheckQuery = `
SELECT column_name, data_type
FROM information_schema.columns
WHERE table_name = 'items' AND column_name = ANY($1)
`

const ownerCountQuery = `SELECT COUNT(DISTINCT owner_id) FROM items`

// notebooksQuery finds every notebook (jop_type = 2), used to resolve
// "Session Meetings" by decoding each candidate's title.
const notebooksQuery = `SELECT jop_id, content FROM items WHERE jop_type = $1`

// childNotebooksQuery finds every notebook whose parent is the given
// jop_id, used to resolve "Archived" without matching its title globally.
const childNotebooksQuery = `SELECT jop_id, content FROM items WHERE jop_type = $1 AND jop_parent_id = $2`

// encryptionCheckQuery reports any note under the given notebooks that
// fails the jop_encryption_applied = 0 invariant.
const encryptionCheckQuery = `
SELECT jop_id
FROM items
WHERE jop_type = $1 AND jop_parent_id = ANY($2) AND jop_encryption_applied != 0
`

const listNotesQuery = `SELECT jop_id, content FROM items WHERE jop_type = $1 AND jop_parent_id = $2`

const getNoteQuery = `
SELECT content
FROM items
WHERE jop_id = $1 AND jop_type = $2 AND jop_parent_id = ANY($3)
`
