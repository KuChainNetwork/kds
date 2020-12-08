package dbservice

import (
	"kds/db"
	"kds/dbmodel"
	"kds/singleton"
	"testing"
)

func TestBlockData_AddAll(t *testing.T) {
	err := db.Initialize("dev", "dev", "localhost", "dev", 3306, 60)
	if nil != err {
		t.Error(err)
		return
	}
	if err = Initialize(); nil != err {
		t.Error(err)
		return
	}
	srv := NewBlockData()
	var list []*dbmodel.BlockData
	if list, err = srv.List(singleton.DB, 0, 32, false); nil != err {
		t.Error(err)
		return
	} else if nil == list || 0 >= len(list) {
		t.Log("Empty")
		return
	}
	err = srv.AddAll(singleton.DB, list)
	if nil != err {
		t.Error(err)
		return
	}
	t.Log("OK")
}
