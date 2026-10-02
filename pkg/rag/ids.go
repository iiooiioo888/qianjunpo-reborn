package rag

// DocumentIDs returns retrievable chunk ids in Top-K order.
func DocumentIDs(docs []Document) []string {
	if len(docs) == 0 {
		return nil
	}
	out := make([]string, len(docs))
	for i, d := range docs {
		out[i] = d.ID
	}
	return out
}
