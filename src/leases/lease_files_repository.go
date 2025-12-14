package leases

import (
	"database/sql"
	"fmt"
)

type LeaseFileRepository interface {
	Create(leaseFile LeaseFile) (LeaseFile, error)
	GetByLeaseID(leaseID int) ([]LeaseFile, error)
	GetByID(id int) (LeaseFile, error)
	Delete(id int) error
}

type leaseFileRepo struct {
	db *sql.DB
}

func NewLeaseFileRepository(db *sql.DB) LeaseFileRepository {
	return &leaseFileRepo{db: db}
}

func (r *leaseFileRepo) Create(leaseFile LeaseFile) (LeaseFile, error) {
	result, err := r.db.Exec(
		"INSERT INTO lease_files (lease_id, filename, original_name, file_path, file_type, file_size) VALUES (?, ?, ?, ?, ?, ?)",
		leaseFile.LeaseID, leaseFile.Filename, leaseFile.OriginalName, leaseFile.FilePath, leaseFile.FileType, leaseFile.FileSize,
	)

	if err != nil {
		return leaseFile, err
	}

	id, _ := result.LastInsertId()
	leaseFile.ID = int(id)

	err = r.db.QueryRow(
		"SELECT id, lease_id, filename, original_name, file_path, file_type, file_size, created_at, updated_at FROM lease_files WHERE id = ?",
		leaseFile.ID,
	).Scan(
		&leaseFile.ID, &leaseFile.LeaseID, &leaseFile.Filename, &leaseFile.OriginalName, &leaseFile.FilePath, &leaseFile.FileType, &leaseFile.FileSize, &leaseFile.CreatedAt, &leaseFile.UpdatedAt,
	)

	return leaseFile, err
}

func (r *leaseFileRepo) GetByLeaseID(leaseID int) ([]LeaseFile, error) {
	rows, err := r.db.Query(
		"SELECT id, lease_id, filename, original_name, file_path, file_type, file_size, created_at, updated_at FROM lease_files WHERE lease_id = ? ORDER BY created_at DESC",
		leaseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leaseFiles := []LeaseFile{}
	for rows.Next() {
		var leaseFile LeaseFile
		err := rows.Scan(
			&leaseFile.ID, &leaseFile.LeaseID, &leaseFile.Filename, &leaseFile.OriginalName, &leaseFile.FilePath, &leaseFile.FileType, &leaseFile.FileSize, &leaseFile.CreatedAt, &leaseFile.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		leaseFiles = append(leaseFiles, leaseFile)
	}

	return leaseFiles, nil
}

func (r *leaseFileRepo) GetByID(id int) (LeaseFile, error) {
	var leaseFile LeaseFile
	err := r.db.QueryRow(
		"SELECT id, lease_id, filename, original_name, file_path, file_type, file_size, created_at, updated_at FROM lease_files WHERE id = ?",
		id,
	).Scan(
		&leaseFile.ID, &leaseFile.LeaseID, &leaseFile.Filename, &leaseFile.OriginalName, &leaseFile.FilePath, &leaseFile.FileType, &leaseFile.FileSize, &leaseFile.CreatedAt, &leaseFile.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return leaseFile, fmt.Errorf("lease file not found")
	}

	return leaseFile, err
}

func (r *leaseFileRepo) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM lease_files WHERE id = ?", id)
	return err
}

