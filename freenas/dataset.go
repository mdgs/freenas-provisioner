package freenas

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/golang/glog"
)

var (
	_ FreenasResource = &Dataset{}
)

// Dataset is a representation of freenas dataset
type Dataset struct {
	Avail      int64  `json:"avail,omitempty"`
	Mountpoint string `json:"mountpoint,omitempty"`
	Name       string `json:"name"`
	Pool       string `json:"pool"`
	Refer      int64  `json:"refer,omitempty"`
	Used       int64  `json:"used,omitempty"`
	Comments   string `json:"comments,omitempty"`
}

func (d *Dataset) String() string {
	return filepath.Join(d.Pool, d.Name)
}

// CopyFrom duplicate freenas dataset resource
func (d *Dataset) CopyFrom(source FreenasResource) error {
	src, ok := source.(*Dataset)
	if ok {
		d.Avail = src.Avail
		d.Mountpoint = src.Mountpoint
		d.Name = src.Name
		d.Pool = src.Pool
		d.Refer = src.Refer
		d.Used = src.Used
	}

	return errors.New("Cannot copy, src is not a Dataset")
}

// Get list all the freenas dataset
func (d *Dataset) Get(server *FreenasServer) error {
	// Getting all datasets (default is 20)
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/datasets/?limit=1000", d.Pool)
	var datasets []Dataset
	resp, err := server.getSlingConnection().Get(endpoint).ReceiveSuccess(&datasets)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	for _, ds := range datasets {
		if ds.Pool == d.Pool && ds.Name == d.Name {
			d.CopyFrom(&ds)
			return nil
		}
	}

	// Nothing found
	return errors.New("No dataset has been found")
}

func (d *Dataset) Create(server *FreenasServer) error {
	parent, dsName := filepath.Split(d.Name)
	endpoint := filepath.Join("/api/v1.0/storage/dataset/", d.Pool, parent) + "/"

	d.Name = dsName

	//glog.Infof("POST endpoint %s, %+v", endpoint, *d)

	resp, err := server.getSlingConnection().Post(endpoint).BodyJSON(d).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error creating dataset - message: %v, status: %d", body, resp.StatusCode))
	}

	return nil
}

func (d *Dataset) Delete(server *FreenasServer) error {
	endpoint := fmt.Sprintf("/api/v1.0/storage/volume/%s/datasets/%s/", d.Pool, d.Name)
	resp, err := server.getSlingConnection().Delete(endpoint).Receive(nil, nil)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := ioutil.ReadAll(resp.Body)
		return errors.New(fmt.Sprintf("Error deleting dataset %+v - %v", *d, body))
	}

	return nil
}
