package leases

type LeaseFile struct {
	ID           int    `json:"id"`
	LeaseID      int    `json:"lease_id"`
	Filename     string `json:"filename"`
	OriginalName string `json:"original_name"`
	FilePath     string `json:"file_path"`
	FileType     string `json:"file_type"`
	FileSize     int64  `json:"file_size"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

