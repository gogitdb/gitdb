package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/mgo.v2/bson"
)

func (g *Gitdb) Insert(m Model) error {

	stamp(m)

	if _, err := os.Stat(g.fullPath(m)); err != nil {
		os.MkdirAll(g.fullPath(m), 0755)
	}

	if !m.Validate() {
		return errors.New("Model is not valid")
	}

	if g.getLock() {
		if err := g.flushQueue(m); err != nil {
			log(err.Error())
		}
		err := g.write(m)
		g.releaseLock()
		return err
	} else {
		return g.queue(m)
	}
}

func (g *Gitdb) InsertMany(models []Model) error {
	//todo polish this up later
	if len(models) > 100 {
		return errors.New("max number of models InsertMany supports is 100")
	}

	tx := g.NewTransaction("InsertMany")
	for _, m := range models {
		tx.AddOperation(func() error { return g.Insert(m) })
	}
	return tx.Commit()
}

func (g *Gitdb) queue(m Model) error {

	dataBlock, err := g.loadBlock(g.queueFilePath(m), m.GetSchema().Name())
	if err != nil {
		return err
	}

	writeErr := g.writeBlock(g.queueFilePath(m), dataBlock)
	if writeErr != nil {
		return writeErr
	}

	return g.updateId(m)
}

func (g *Gitdb) flushQueue(m Model) error {

	if _, err := os.Stat(g.queueFilePath(m)); err == nil {

		log("flushing queue")
		dataBlock := NewBlock(m.GetSchema().Name())
		err := g.readBlock(g.queueFilePath(m), dataBlock)
		if err != nil {
			return err
		}

		//todo optimize: this will open and close block file to delete each record it flushes
		model := g.config.Factory(m.GetSchema().Name())
		for recordId, record := range dataBlock.records {
			log("Flushing: " + recordId)

			record.Hydrate(model)

			err = g.write(model)
			if err != nil {
				println(err.Error())
				return err
			}
			err = g.delById(recordId, m.GetSchema().Name(), g.queueFilePath(m), false)
			if err != nil {
				return err
			}
		}

		return os.Remove(g.queueFilePath(m))
	}

	log("empty queue :)")

	return nil
}

func (g *Gitdb) flushDb() error {
	return nil
}

func (g *Gitdb) write(m Model) error {

	blockFilePath := g.blockFilePath(m)
	commitMsg := "Inserting " + m.Id() + " into " + blockFilePath

	dataBlock, err := g.loadBlock(blockFilePath, m.GetSchema().Name())
	if err != nil {
		return err
	}

	//...append new record to block
	newRecordBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	if _, err := dataBlock.Get(m.Id()); err == nil {
		commitMsg = "Updating " + m.Id() + " in " + blockFilePath
	}

	logTest(commitMsg)

	newRecordStr := string(newRecordBytes)
	dataBlock.Add(m.GetSchema().RecordId(), newRecordStr)

	g.events <- newWriteBeforeEvent("...", blockFilePath)
	writeErr := g.writeBlock(blockFilePath, dataBlock)
	if writeErr != nil {
		return writeErr
	}

	log(fmt.Sprintf("autoCommit: %v", g.autoCommit))

	logTest("sending write event to loop")
	g.events <- newWriteEvent(commitMsg, blockFilePath)
	g.updateIndexes(m.GetSchema().Name(), newRecord(m.Id(), newRecordStr))

	//what is the effect of this on InsertMany?
	return g.updateId(m)
}

func (g *Gitdb) writeBlock(blockFile string, block *Block) error {

	model := g.getModelFromCache(block.dataset)

	//encrypt data if need be
	if model.ShouldEncrypt() {
		for k, record := range block.records {
			block.Add(k, encrypt(g.config.EncryptionKey, record.data))
		}
	}

	//determine which format we need to write data in
	var blockBytes []byte
	var fmtErr error
	switch model.GetDataFormat() {
	case JSON:
		blockBytes, fmtErr = json.MarshalIndent(block.data(), "", "\t")
		break
	case BSON:
		blockBytes, fmtErr = bson.Marshal(block.data())
		break
	}

	if fmtErr != nil {
		return fmtErr
	}

	return ioutil.WriteFile(blockFile, blockBytes, 0744)
}

func (g *Gitdb) Delete(id string) error {
	return g.delete(id, false)
}

func (g *Gitdb) DeleteOrFail(id string) error {
	return g.delete(id, true)
}

func (g *Gitdb) delete(id string, failNotFound bool) error {

	dataDir, _, _, err := g.ParseId(id)
	if err != nil {
		return err
	}

	model := g.getModelFromCache(dataDir)

	blockFilePath := g.blockFilePath(model)
	err = g.delById(id, dataDir, blockFilePath, failNotFound)

	if err == nil {
		logTest("sending delete event to loop")
		g.events <- newDeleteEvent("Deleting "+id+" in "+blockFilePath, blockFilePath)
	}

	return err
}

func (g *Gitdb) delById(id string, dataset string, blockFile string, failIfNotFound bool) error {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	dataBlock := NewBlock(dataset)
	err := g.readBlock(blockFile, dataBlock)
	if err != nil {
		return err
	}

	if err := dataBlock.Delete(id); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	//write undeleted records back to block file
	return g.writeBlock(blockFile, dataBlock)
}
