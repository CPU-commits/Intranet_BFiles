package services

type FileNats struct {
	Location string `json:"location"`
	Filename string `json:"filename"`
	Mimetype string `json:"mime-type"`
	Key      string `json:"key"`
}

type FilePermission struct {
	Files  []string `json:"files"`
	IDUser string   `json:"id_user"`
}
