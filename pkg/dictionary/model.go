package dictionary

type Library struct {
	Library []LibraryData `json:"library"`
}

type LibraryData struct {
	Key         string      `json:"key"`
	Translation translation `json:"translation"`
}

type translation struct {
	ID string `json:"id"`
	EN string `json:"en"`
}
