package records

// type IdxRecord struct {
// 	File      string `json:"file"`
// 	Offset    uint64 `json:"offset"`
// 	Length    uint16 `json:"length"`
// 	Fuzziness uint8  `json:"fuzziness"`
// 	Data      []byte `json:"data"`
// }

func (r IdxRecord) OldJsonable() map[string]interface{} {
	var index = map[string]interface{}{
		"file":      r.File,
		"offset":    r.Offset,
		"length":    r.Length,
		"fuzziness": r.Fuzziness,
		"data":      r.Data,
	}

	return index
}

func (r IdxRecord) Jsonable() map[string]interface{} {
	var index = map[string]interface{}{
		"file":      r.File,
		"offset":    r.Offset,
		"length":    r.Length,
		"fuzziness": r.Fuzziness,
		"data":      r.Data,
	}

	return map[string]interface{}{
		"_index": index,
	}
}
