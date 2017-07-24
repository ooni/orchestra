// Code generated by go-bindata.
// sources:
// proteus-events/data/migrations/1_jobs_create.sql
// proteus-events/data/migrations/1_tasks_create.sql
// proteus-events/data/migrations/2_add_jobs_state.sql
// proteus-events/data/migrations/3_add_job_type_tables.sql
// proteus-events/data/templates/home.tmpl
// DO NOT EDIT!

package events

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _dataMigrations1_jobs_createSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x92\x41\x8f\x9b\x30\x10\x85\xef\xfe\x15\xef\x10\x89\x44\xdd\x95\xba\xbd\xa2\x1e\x60\x71\x1a\x6f\xc1\x44\x60\xba\x4d\xab\xca\xf2\x86\x29\x65\x1b\x0c\x02\xa7\x6d\xfe\x7d\x05\x49\x94\x44\xda\x9b\x9f\xe7\xe3\x31\x6f\xc6\xf7\xf7\x78\xd7\xd4\x55\x6f\x1c\x21\x6a\xff\x5a\x16\x65\xe9\x1a\x2a\x08\x63\x0e\xb1\x04\xff\x2a\x72\x95\xe3\xb5\x7d\x19\x7c\xc6\xae\xe1\xa2\xbb\x91\xb9\x33\x8e\x1a\xb2\x2e\xa4\xaa\xb6\x2c\x4a\x31\x9b\x31\x00\x08\xf9\x27\x21\xa7\x93\x58\x42\xa6\xea\x6c\x39\xcf\x79\xcc\x1f\x15\x1e\xb0\xcc\xd2\x04\x5d\xa5\xdd\xa1\x23\x3c\xaf\x78\xc6\xe1\x0e\x9d\x35\x0d\xe1\x23\xbc\xd7\xf6\x45\x0f\xa3\xb9\xb7\x80\x5a\xf1\xa3\xd5\x63\xc6\x03\xc5\xa1\x36\x6b\x8e\xa7\x34\xd4\xb9\x1a\x65\x90\x83\xcb\x22\xc1\xdc\x33\x5b\x57\xff\x21\xef\x0e\x5e\x49\x3b\x72\x54\x4e\xc7\xd6\x92\xb7\xf0\x27\x03\x2e\x23\x88\xa5\xcf\xb8\x8c\x66\x33\x9f\xb1\xb3\xe1\x39\xf7\x55\xa3\x63\x76\x36\x9f\xbe\xaa\x4b\x14\x85\x88\xb0\xce\x44\x12\x64\x1b\x7c\xe6\x9b\x89\x94\x45\x1c\xdf\x4d\xc4\xb6\x6d\xc6\x21\xe0\x4b\x90\x3d\xae\x82\xec\x74\xd9\x93\x71\x75\x6b\xb5\xab\x1b\x82\x12\x09\xcf\x55\x90\xac\xf1\x2c\xd4\x6a\x92\xf8\x96\x4a\x7e\x64\x87\xed\x2f\x2a\xf7\x3b\xba\x75\x28\x69\x67\x0e\x10\x52\x1d\xa5\x33\x7d\x45\x4e\x6f\xdb\xbd\x75\x7d\x4d\xc3\x19\x9e\x7f\x58\xe0\xfb\x8f\x1b\xa6\xdb\x19\xf7\xb3\xed\x9b\x0b\xf3\xf0\xfe\x1a\x1a\x7e\x6b\x47\x83\xd3\xd3\xb8\x6f\xfe\x39\xd5\x4c\x5f\xed\xc7\x40\x03\x9e\xf2\x54\x86\xa7\x4a\xdd\xd0\xa0\xfb\xbd\xbd\x74\x64\xe9\x9f\x1b\x6f\xb4\x71\xc7\x44\x6f\x65\xab\x07\x3d\xee\x00\x61\x9a\xc6\x3c\x90\xa7\xc0\xe3\x6e\x2f\x4b\x64\x0b\xff\xed\x77\xc5\x6d\xc9\xfe\x07\x00\x00\xff\xff\x0d\xbe\x24\xa3\xad\x02\x00\x00")

func dataMigrations1_jobs_createSqlBytes() ([]byte, error) {
	return bindataRead(
		_dataMigrations1_jobs_createSql,
		"data/migrations/1_jobs_create.sql",
	)
}

func dataMigrations1_jobs_createSql() (*asset, error) {
	bytes, err := dataMigrations1_jobs_createSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/migrations/1_jobs_create.sql", size: 685, mode: os.FileMode(420), modTime: time.Unix(1498387809, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataMigrations1_tasks_createSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x92\x41\x8f\xd3\x30\x10\x85\xef\xfe\x15\xef\x50\x29\xad\xd8\x3d\x70\x8e\x38\xa4\x8d\x4b\x0d\xad\x53\xc5\x0e\x0b\x5c\x22\x6f\x32\x44\x59\x68\x12\xc5\xb3\x82\xfe\x7b\x14\xa7\x8b\xba\x12\x2b\xf5\xf6\xe6\xf9\xcd\x8c\xe6\x93\xef\xef\xf1\xee\xd4\x36\xa3\x63\x42\xda\xff\xee\x44\x9a\x67\x47\xd8\x64\xbd\x97\x50\x5b\xd0\x9f\xd6\xb3\x07\x3b\xff\xd3\xc7\x42\x5c\xa7\x8b\xe1\x55\x69\xd8\x31\x9d\xa8\xe3\x35\x35\x6d\x27\xd2\x0c\x8b\x85\x00\x80\xb5\xfc\xa8\x74\x50\x6a\x0b\x9d\x59\xc8\xaf\xca\x58\x83\xa5\x91\x7b\xb9\xb1\x78\x8f\x6d\x9e\x1d\x30\x34\x25\x9f\x07\xc2\xc3\x4e\xe6\x12\x7c\x1e\x3a\x77\x22\x7c\x40\x34\xed\x2e\xfd\x34\x3d\x5a\xc1\xee\xe4\x3c\x6b\x93\xcb\xc4\x4a\xd8\x6f\x47\x09\x9b\x98\xcf\xa5\xb1\x53\x9d\x18\x48\x5d\x1c\xb0\x8c\x46\x72\xf5\x39\xba\x43\xd4\xf5\xdc\xfe\x68\xa9\x9e\xb4\xab\x2a\x1a\x78\xd6\x23\x3d\x51\x75\xd1\x75\xdf\x51\xb4\x8a\xc3\x64\xa9\x53\xa8\x6d\x2c\xa4\x4e\x17\x8b\x58\x88\x97\x4d\x2f\x48\xae\x4e\x08\x58\xc4\x32\xb4\xb5\x35\x8a\x42\xa5\xe1\x59\x17\xfb\xfd\x5d\x70\x87\xb1\x7f\xa4\xf2\xf2\x36\x5b\x4f\xfd\xe3\x6b\x83\xc9\x73\x19\xae\xfd\x92\xe4\x9b\x5d\x92\xcf\xb6\x1b\x9b\xe7\x89\xa7\xc7\x27\x93\xe9\xf5\x6c\x06\x0e\x57\x07\xff\xdb\xd2\x8c\xe4\x3d\x94\xb6\xb3\x53\x8d\xe4\xb8\xed\xbb\x92\xdb\x13\xc1\xaa\x83\x34\x36\x39\x1c\xf1\xa0\xec\x2e\x94\xf8\x9e\xe9\x4b\xf7\x0c\xa8\xba\x39\x3f\x43\xbc\x25\x39\x61\xbd\x25\xf7\xcb\x79\x2e\x9f\x87\xda\x31\xd5\x6f\x46\xc5\x2a\xfe\xff\x87\x93\x5d\x2d\xfe\x06\x00\x00\xff\xff\x3a\x61\xc2\x6a\xc7\x02\x00\x00")

func dataMigrations1_tasks_createSqlBytes() ([]byte, error) {
	return bindataRead(
		_dataMigrations1_tasks_createSql,
		"data/migrations/1_tasks_create.sql",
	)
}

func dataMigrations1_tasks_createSql() (*asset, error) {
	bytes, err := dataMigrations1_tasks_createSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/migrations/1_tasks_create.sql", size: 711, mode: os.FileMode(420), modTime: time.Unix(1498387808, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataMigrations2_add_jobs_stateSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x91\x4f\x8f\x9b\x30\x10\xc5\xef\xfe\x14\xef\x80\xe4\x5d\xb5\x5b\xa9\x67\xd4\x03\x7f\x86\xc6\x15\x31\x11\x38\xda\xed\x09\xd8\x60\x45\xac\xc0\xa0\xe0\xb4\xcd\xb7\xaf\x70\x9b\x7f\x4d\xba\x9c\x06\xcf\xf8\xcd\xef\x3d\x3f\x3d\xe1\x43\xdf\x6e\x77\xb5\xd5\x88\x87\x9f\x86\x5d\x1e\x14\xb6\xb6\xba\xd7\xc6\x86\x7a\xdb\x1a\x16\xe7\xd9\x0a\xea\xfb\x8a\xf0\x2d\x0b\xcb\x42\x05\x8a\x20\x12\xd0\x8b\x28\x54\xe1\xb3\x20\x55\x94\x43\x05\x61\x4a\x78\x1b\x5e\x27\xb8\xf9\x28\x4b\xd7\x4b\x79\x9e\xc3\x34\x8b\xfa\xf7\xf7\x90\x69\xd8\x55\x67\x3d\xbe\x0b\x94\xc1\xf3\x18\x00\x84\xf4\x55\x48\x57\x89\x04\x32\x53\xc7\x65\x0f\x05\xa5\x14\x29\x7c\x46\x92\x67\x4b\x8c\xdb\xd2\x1e\x46\x8d\xe7\x05\xe5\x04\x7b\x18\x4d\xdd\x6b\x7c\x01\x7f\x1b\x5e\x4b\x07\xc6\x1f\xa1\x16\xf4\x47\x2a\xca\x69\xb6\xf8\x8f\xe3\xa0\x00\xc9\xf5\x12\x0f\xbc\xde\xd8\xf6\x87\xe6\x1f\xc1\x1b\xdd\x69\xab\x1b\x57\x0e\x46\xf3\x47\xdf\x09\x90\x8c\x21\x12\x9f\x91\x8c\x3d\xcf\x67\x77\x79\x6f\xff\xe6\xef\x26\xcb\x20\x8e\x8f\x51\x3a\xce\x33\x90\x7f\xba\x48\x2f\x11\xad\x94\xc8\xae\xa5\x9e\x17\x24\xd1\xec\xc7\xae\xdd\xd4\x56\x97\x9b\xa1\xdb\xf7\xc6\x99\x44\x1e\x88\x82\xe6\xb8\x44\x44\xe0\x7f\x3b\x95\xd3\xaf\x50\x77\x3b\x5d\x37\x07\xe8\x5f\xed\x64\x27\xb4\x06\xd5\x4c\x52\x7d\xe2\x17\x1b\x65\x7c\x72\xea\x33\xcf\xfb\xff\xab\xfe\x0e\x00\x00\xff\xff\x64\x65\x02\xda\x68\x02\x00\x00")

func dataMigrations2_add_jobs_stateSqlBytes() ([]byte, error) {
	return bindataRead(
		_dataMigrations2_add_jobs_stateSql,
		"data/migrations/2_add_jobs_state.sql",
	)
}

func dataMigrations2_add_jobs_stateSql() (*asset, error) {
	bytes, err := dataMigrations2_add_jobs_stateSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/migrations/2_add_jobs_state.sql", size: 616, mode: os.FileMode(420), modTime: time.Unix(1500482133, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataMigrations3_add_job_type_tablesSql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x94\x4f\x73\xda\x3c\x10\xc6\xef\xfe\x14\xba\x91\xcc\xfb\xa6\x1f\x20\x9c\x0c\xa8\x6d\x5a\x0a\xd4\x98\xce\x70\xd2\xac\xed\xc5\x28\xc8\x92\x23\xad\x21\xf9\xf6\x1d\xdb\xc8\x31\x34\x10\x8e\x7e\xf6\xd1\xea\xa7\xfd\xe3\x87\x07\xf6\x5f\x21\x73\x0b\x84\x6c\x62\x0e\x3a\xe8\x0b\x4b\x02\xc2\x02\x35\x8d\x30\x97\x3a\x98\x44\xf3\x05\x8b\xc3\xd1\x94\xb3\x67\x93\x08\x02\xb7\x73\xc3\x73\x15\x14\x5a\x72\xc3\x20\x08\xa7\x31\x8f\x8e\x81\xb2\x4a\x94\x4c\xbf\x3c\x9b\xc4\xb1\xc6\x5f\x9f\x15\xda\x0c\xaf\xbb\x9a\x5c\x8d\xed\xa2\x2f\x9c\x4c\xda\x64\x84\x8e\x84\x86\x02\xd9\x9f\x30\x1a\x7f\x0f\x23\x36\x5b\x4d\xa7\x97\x2f\xe8\x0e\x82\xcd\xab\xfa\x8d\x8e\xfd\x58\xce\x67\xa3\x61\x70\x7c\xd1\x7a\xc1\xd9\x7c\x1e\xf3\x65\x7c\x7c\xe3\x92\xff\x5e\xf1\xd9\x98\x7b\x78\xe1\xf0\xe5\x3c\xe4\x89\xdb\xd8\x87\xb5\xe4\x3a\x0b\x4e\x22\xab\xf2\x5a\xd1\x53\x8b\xb5\x4a\x6f\x25\x1e\x71\x18\x38\x86\xba\x2a\xd8\x5d\xc0\x18\x63\x83\x03\x26\x22\x35\x5a\x63\x4a\x72\x2f\xe9\x6d\xf0\x7f\xab\x6f\x89\x4a\x61\xf1\xa5\x42\x47\xce\x8b\x99\x76\xb5\xd9\x49\x47\xa8\xd3\x53\xaf\xd4\x7b\x50\x32\xf3\x67\x84\x92\x1a\xbd\x21\xb1\x32\xcb\x51\x58\x84\x74\x0b\x89\x54\xbd\x7b\x28\x2d\xfd\xfd\x27\xe9\xb6\x08\x19\x5a\xb1\x91\xa8\x32\x51\x80\x96\x65\xa5\x80\xa4\xd1\xa7\x2e\xe3\xba\x63\x45\xa5\x48\x8a\xd2\x1a\x32\xa9\x51\x82\x2c\xa4\x68\x4d\x45\x1d\x45\x81\xb8\x13\x1b\x6b\x34\x61\x87\xe9\x9a\xd6\x7b\xc7\x61\x0b\xe4\xa0\x2c\xfd\xf7\x1e\xb4\x54\x0a\x04\x19\xeb\xa5\x0d\xa4\x98\x18\xb3\x13\x05\x3a\x87\x3a\xc7\x2e\xa2\x33\x1a\x04\xf7\xc3\x20\x18\x47\x3c\x8c\xf9\x85\x8e\xfb\xe8\xd9\x26\x34\xed\x38\x1a\xd9\xd3\x2c\xe6\xdf\x78\xc4\x26\xfc\x6b\xb8\x9a\xc6\x4c\xe3\x2b\xed\x41\xdd\x0d\x7a\x99\x06\x8f\x8f\x16\xf3\x54\x81\x73\xf7\x6c\x11\x3d\xfd\x0a\xa3\x35\xfb\xc9\xd7\x6c\x36\x8f\x9b\xe1\xad\xa9\xde\xc7\xba\x6d\x7e\xad\x9d\x4d\xec\x87\xc4\xa7\x83\xf8\x2f\x72\xbb\xa6\x0d\xb3\xb7\x5e\x86\xee\x27\xbb\x89\xba\xae\x2c\xe4\xdd\x2a\xd6\x12\xbe\x92\x85\x1e\xf0\xa7\x7b\xd9\x02\xdd\xb0\xc5\x7d\xfe\x4f\xec\xef\xff\x9e\xae\xae\xb7\x98\xbb\x82\x5f\x59\xe9\xbf\x01\x00\x00\xff\xff\x93\xf2\x5e\x9a\x49\x05\x00\x00")

func dataMigrations3_add_job_type_tablesSqlBytes() ([]byte, error) {
	return bindataRead(
		_dataMigrations3_add_job_type_tablesSql,
		"data/migrations/3_add_job_type_tables.sql",
	)
}

func dataMigrations3_add_job_type_tablesSql() (*asset, error) {
	bytes, err := dataMigrations3_add_job_type_tablesSqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/migrations/3_add_job_type_tables.sql", size: 1353, mode: os.FileMode(420), modTime: time.Unix(1498242860, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _dataTemplatesHomeTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x54\x41\x73\xdb\x36\x13\x3d\x93\xbf\x62\x3f\xe4\xf6\x8d\x68\x4a\x69\xd3\xda\x34\xc9\x43\xec\x66\x92\x43\xed\x4c\x9d\x1c\x7a\x04\xc1\x25\x89\x06\xc4\x72\x80\x95\x2c\xc5\xa3\xff\xde\x01\x28\x31\xaa\x3b\xd3\x93\x76\x1f\x76\xdf\x7b\x5a\x60\x59\xfe\xef\xfe\xf1\xee\xcb\x9f\x9f\x7f\x83\x81\x47\x53\xa7\x65\xf8\x01\x23\x6d\x5f\x09\xb4\xa2\x4e\x01\xca\x01\x65\x1b\x83\x11\x59\x82\x1a\xa4\xf3\xc8\x95\xd8\x72\x97\x5d\x8b\x1f\x07\x56\x8e\x58\x89\x9d\xc6\xe7\x89\x1c\x0b\x50\x64\x19\x2d\x57\xe2\x59\xb7\x3c\x54\x2d\xee\xb4\xc2\x2c\x26\x2b\xd0\x56\xb3\x96\x26\xf3\x4a\x1a\xac\x36\xa2\x4e\x03\x0f\x6b\x36\x58\xbf\xbc\xc0\x55\x8c\xe0\x78\x2c\xf3\x19\x4b\x93\xd2\xf3\xc1\x20\xf0\x61\xc2\x4a\x30\xee\x39\x57\xde\x8b\x3a\x4d\xfe\x0f\x2f\x69\x92\x8c\xd2\xf5\xda\x16\xb0\xbe\x4d\x93\x64\x92\x6d\xab\x6d\x7f\xca\x42\x71\xe6\xd0\xb6\xe8\x22\xd8\x23\x8d\xc8\x4e\xab\xcf\x0e\x95\xf6\x9a\x6c\xa8\x6a\x68\x9f\x79\xfd\x3d\x56\x34\xe4\x5a\x74\x59\x43\xfb\xdb\x34\x39\xa6\x49\x43\xed\x61\x15\x27\x14\xb5\x3a\xb2\x9c\x75\x72\xd4\xe6\x50\x40\x26\xa7\xc9\x60\xe6\x0f\x9e\x71\x5c\xc1\x7b\xa3\xed\xb7\xdf\xa5\x7a\x8a\xf9\x07\xb2\xbc\x02\xf1\x84\x3d\x21\x7c\xfd\x24\x56\x20\xfe\xa0\x86\x98\x42\xf4\xb8\x3f\xf4\x68\x43\xf4\xb5\xd9\x5a\xde\x86\xe8\x4e\x5a\x96\x0e\x8d\x09\xc9\x07\xed\x24\x3c\x49\xeb\x43\x72\xef\x48\xb7\x4b\xf6\x11\xcd\x0e\x59\x2b\x09\x0f\xb8\x45\xb1\x02\x2f\xad\xcf\x3c\x3a\xdd\x45\xcb\x00\x00\xc1\x35\xbc\xc4\x10\xa0\x91\xea\x5b\xef\x68\x6b\xdb\x02\xde\x74\x5d\x77\x7b\xc2\x97\x51\xfd\xb4\x9e\xf6\x33\x38\x77\x8f\x52\xdb\xa5\x7b\x94\xfb\xf9\xe6\x0a\xb8\x79\xfb\xaa\xf0\x6a\x40\x63\xe8\xa2\x34\x5c\x44\xd6\x10\x33\x8d\x97\xb4\x00\x71\x6e\x5e\x7f\xc7\x02\x36\xd7\xaf\xe0\x67\xd4\xfd\xc0\x05\xbc\x5d\xaf\xcf\xb8\xd1\x16\xb3\xe1\x84\xbf\xb6\xb7\xe8\x0e\x9b\x45\xfa\x82\xff\x5f\xb2\x67\xfe\x77\x3f\xf8\x4f\x4e\x99\xa6\xf8\x50\x66\x50\x91\x21\x57\xc0\x9b\xf5\xcf\xbf\xfc\x7a\x73\x73\xa9\x28\x17\x9d\xa5\xe6\xdd\xf5\xf5\xdd\xfb\x73\x67\x7c\x66\x2d\x2a\x72\x92\x35\xd9\x02\x2c\x59\x3c\x13\x24\x65\x1e\xdf\x6f\x5c\x97\x7c\xd9\xa8\x70\x45\x75\x2c\x29\xc3\xbc\xeb\x13\x55\xd9\xea\x1d\x28\x23\xbd\xaf\x44\xfc\x97\xe2\x7c\x12\xd6\x71\x53\x7f\x0c\xd8\x0a\x78\xd0\x1e\xb4\x87\xb0\x30\x8a\xc6\x89\x2c\x5a\x7e\x90\xe3\xbc\x38\xc3\xe6\xa2\x69\xaa\x27\xe9\x18\xa8\x03\x1e\x10\x88\xac\x9e\x1c\x35\x08\xe4\xd4\x80\x9e\x67\xcb\x30\x3f\xe2\x32\x9f\x16\x23\x79\xab\x77\x4b\xb2\xc0\xff\x10\xbc\x47\xaf\x9c\x9e\x22\xc1\xf1\xb8\x34\x4e\x17\x6d\x5f\x08\x0c\x4a\x67\x61\x24\x87\x20\x1b\xda\x32\x3c\x3e\x3e\x7c\x82\x9d\xf6\x9a\x0b\x28\x25\x0c\x0e\xbb\x4a\x0c\xcc\x93\x2f\xf2\x3c\x18\xbc\x62\x72\x93\xa3\xbf\x50\xf1\x15\xb9\x3e\x17\xf5\x7f\x9d\x96\xb9\xac\x17\xd1\x32\x3f\x4f\xb3\xcc\xe7\x11\x97\xf9\xfc\x7d\xfb\x3b\x00\x00\xff\xff\xec\xc7\xc0\x2a\xf0\x04\x00\x00")

func dataTemplatesHomeTmplBytes() ([]byte, error) {
	return bindataRead(
		_dataTemplatesHomeTmpl,
		"data/templates/home.tmpl",
	)
}

func dataTemplatesHomeTmpl() (*asset, error) {
	bytes, err := dataTemplatesHomeTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "data/templates/home.tmpl", size: 1264, mode: os.FileMode(420), modTime: time.Unix(1495831185, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"data/migrations/1_jobs_create.sql": dataMigrations1_jobs_createSql,
	"data/migrations/1_tasks_create.sql": dataMigrations1_tasks_createSql,
	"data/migrations/2_add_jobs_state.sql": dataMigrations2_add_jobs_stateSql,
	"data/migrations/3_add_job_type_tables.sql": dataMigrations3_add_job_type_tablesSql,
	"data/templates/home.tmpl": dataTemplatesHomeTmpl,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"data": &bintree{nil, map[string]*bintree{
		"migrations": &bintree{nil, map[string]*bintree{
			"1_jobs_create.sql": &bintree{dataMigrations1_jobs_createSql, map[string]*bintree{}},
			"1_tasks_create.sql": &bintree{dataMigrations1_tasks_createSql, map[string]*bintree{}},
			"2_add_jobs_state.sql": &bintree{dataMigrations2_add_jobs_stateSql, map[string]*bintree{}},
			"3_add_job_type_tables.sql": &bintree{dataMigrations3_add_job_type_tablesSql, map[string]*bintree{}},
		}},
		"templates": &bintree{nil, map[string]*bintree{
			"home.tmpl": &bintree{dataTemplatesHomeTmpl, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

