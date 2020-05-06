package gitdb

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//UploadModel represents a file upload
type UploadModel struct {
	Bucket     string
	File       string
	Path       string
	UploadedBy string
	TimeStampedModel
	db *gitdb
}

//GetSchema implements Model.GetSchema
func (u *UploadModel) GetSchema() *Schema {
	name := "Upload"
	return newSchema(
		name,
		// AutoBlock(u.db.dbDir(), name, BlockByCount, 1000),
		"b0",
		u.Bucket+"-"+u.File,
		make(map[string]interface{}),
	)
}

//Validate implements Model.Validate
func (u *UploadModel) Validate() error {
	//todo
	return nil
}

//IsLockable informs GitDb if a Model support locking
func (u *UploadModel) IsLockable() bool { return false }

//GetLockFileNames informs GitDb of files a Models using for locking
func (u *UploadModel) GetLockFileNames() []string { return nil }

//ShouldEncrypt informs GitDb if a Model support encryption
func (u *UploadModel) ShouldEncrypt() bool { return false }

const uploadDataset = "Bucket"

//Upload provides API for managing file uploads
type Upload struct {
	db    *gitdb
	model *UploadModel
}

//Get returns an upload by id
func (u *Upload) Get(id string) error {
	return nil
}

//Delete an upload by id
func (u *Upload) Delete(id string) error {
	return nil
}

//New uploads specified file into bucket
func (u *Upload) New(bucket, file string) error {
	id := "Uploads/b0/" + bucket + file
	err := u.db.Exists(id)
	if err == nil {
		return errors.New("file already exists")
	}

	return u.upload(bucket, file)
}

//Replace overrides a specified file in bucket
func (u *Upload) Replace(bucket, file string) error {
	id := "Uploads/b0/" + bucket + file
	err := u.db.Exists(id)
	if err != nil {
		return err
	}

	return u.upload(bucket, file)
}

func (u *Upload) upload(bucket, file string) error {
	src, err2 := os.Open(file)
	if err2 != nil {
		return err2
	}

	valid := map[string]bool{
		".pdf":  true,
		".doc":  true,
		".xlsx": true,
		".jpg":  true,
		".gif":  true,
		".png":  true,
		".json": true,
		".yaml": true,
		".yml":  true,
		".md":   true,
	}

	ext := filepath.Ext(file)
	if _, ok := valid[ext]; !ok {
		return fmt.Errorf("%s files are not allowed", ext)
	}

	//todo a better func to clean up filenames
	filename := strings.Replace(file, "/", "-", -1)
	filename = strings.Replace(filename, " ", "-", -1)

	uploadPath := filepath.Join(u.db.dbDir(), uploadDataset, bucket, filename)
	fmt.Println(uploadPath)
	os.MkdirAll(path.Dir(uploadPath), os.ModePerm)
	dst, err3 := os.Create(uploadPath)
	if err3 != nil {
		return err3
	}
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	//neutralise file
	dst.Chmod(0400)

	//Uploads/b0/bucket-file.jpg
	m := &UploadModel{Bucket: bucket, File: file, Path: uploadPath, db: u.db}
	return u.db.Insert(m)
}

func (g *gitdb) Upload() *Upload {
	return &Upload{db: g}
}
