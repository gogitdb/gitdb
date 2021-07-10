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

const uploadDataset = "Bucket"

var validExtensions = map[string]bool{
	".html": true,
	".pdf":  true,
	".doc":  true,
	".xlsx": true,
	".csv":  true,
	".jpg":  true,
	".gif":  true,
	".png":  true,
	".svg":  true,
	".json": true,
	".yaml": true,
	".yml":  true,
	".md":   true,
}

//UploadModel represents a file upload
type UploadModel struct {
	Bucket     string
	File       string
	Path       string
	UploadedBy string
	TimeStampedModel
}

//GetSchema implements Model.GetSchema
func (u *UploadModel) GetSchema() *Schema {
	return newSchema(
		uploadDataset,
		// AutoBlock(u.db.dbDir(), name, BlockByCount, 1000),
		u.Bucket,
		u.File,
		make(map[string]interface{}),
	)
}

//Validate implements Model.Validate
func (u *UploadModel) Validate() error {
	//todo
	return nil
}

//Upload provides API for managing file uploads
type Upload struct {
	db    *gitdb
	model *UploadModel
}

//Get returns an upload by id
func (u *Upload) Get(id string, result *UploadModel) error {
	return u.db.Get(id, result)
}

//Delete an upload by id
func (u *Upload) Delete(id string) error {
	var data UploadModel
	if err := u.Get(id, &data); err != nil {
		return err
	}

	err := u.db.Delete(id)
	if err == nil {
		err = os.Remove(data.Path)
	}

	return err
}

//New uploads specified file into bucket
func (u *Upload) New(bucket, file string) error {
	err := u.db.Exists(u.id(bucket, file))
	if err == nil {
		return errors.New("file already exists")
	}

	return u.upload(bucket, file)
}

//Replace overrides a specified file in bucket
func (u *Upload) Replace(bucket, file string) error {
	err := u.db.Exists(u.id(bucket, file))
	if err != nil {
		return err
	}

	return u.upload(bucket, file)
}

func (u *Upload) id(bucket, file string) string {
	return uploadDataset + "/" + bucket + "/" + u.cleanFileName(file)
}

func (u *Upload) cleanFileName(filename string) string {
	//todo a better func to clean up filenames
	filename = strings.ReplaceAll(path.Clean(filename), "/", "-")
	filename = strings.ReplaceAll(filename, " ", "-")
	return filename
}

func (u *Upload) upload(bucket, file string) error {
	var src *os.File
	var dst *os.File
	var err error

	if src, err = os.Open(file); err != nil {
		return err
	}

	ext := filepath.Ext(file)
	if _, ok := validExtensions[ext]; !ok {
		return fmt.Errorf("%s files are not allowed", ext)
	}

	filename := u.cleanFileName(file)
	uploadPath := filepath.Join(u.db.dbDir(), uploadDataset, bucket, filename)
	fmt.Println(uploadPath)
	if err = os.MkdirAll(path.Dir(uploadPath), os.ModePerm); err != nil {
		return err
	}

	if dst, err = os.Create(uploadPath); err != nil {
		return err
	}
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	//neutralise file
	if err = dst.Chmod(0640); err != nil {
		return err
	}

	//Uploads/b0/bucket-file.jpg
	m := &UploadModel{
		Bucket:     bucket,
		File:       filename,
		Path:       uploadPath,
		UploadedBy: u.db.config.User.String(),
	}
	return u.db.Insert(m)
}

func (g *gitdb) Upload() *Upload {
	g.RegisterModel(uploadDataset, &UploadModel{})
	return &Upload{db: g}
}
