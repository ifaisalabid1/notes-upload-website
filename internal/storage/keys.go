package storage

import "fmt"

func NoteKey(subjectID, noteID, ext string) string {
	return fmt.Sprintf("notes/%s/%s.%s", subjectID, noteID, ext)
}
