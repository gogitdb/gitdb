package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func (g *gitdb) insert(m Model) error {

	stamp(m)

	if _, err := os.Stat(g.fullPath(m)); err != nil {
		os.MkdirAll(g.fullPath(m), 0755)
	}

	if !m.Validate() {
		return errors.New("Model is not valid")
	}

	if g.getLock() {
		if err := g.flushQueue(); err != nil {
			log(err.Error())
		}
		err := g.write(m)
		g.releaseLock()
		return err
	}

	return g.queue(m)
}

func (g *gitdb) queue(m Model) error {

	if len(g.writeQueue) == 0 {
		g.writeQueue = map[string]Model{}
	}

	g.writeQueue[m.Id()] = m
	return g.updateId(m)
}

func (g *gitdb) flushQueue() error {

	for id, model := range g.writeQueue {
		log("Flushing: " + id)

		err := g.write(model)
		if err != nil {
			logError(err.Error())
			return err
		}

		delete(g.writeQueue, id)
	}

	return nil
}

func (g *gitdb) write(m Model) error {

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
	if err := g.writeBlock(blockFilePath, dataBlock); err != nil {
		return err
	}

	log(fmt.Sprintf("autoCommit: %v", g.autoCommit))

	logTest("sending write event to loop")
	g.events <- newWriteEvent(commitMsg, blockFilePath)
	g.updateIndexes(m.GetSchema().Name(), newRecord(m.Id(), newRecordStr))

	//what is the effect of this on InsertMany?
	return g.updateId(m)
}

func (g *gitdb) writeBlock(blockFile string, block *Block) error {

	model := g.getModelFromCache(block.dataset)

	//encrypt data if need be
	if model.ShouldEncrypt() {
		for k, record := range block.records {
			block.Add(k, encrypt(g.config.EncryptionKey, record.data))
		}
	}

	blockBytes, fmtErr := json.MarshalIndent(block.data(), "", "\t")
	if fmtErr != nil {
		return fmtErr
	}

	return ioutil.WriteFile(blockFile, blockBytes, 0744)
}

func (g *gitdb) delete(id string) error {
	return g.dodelete(id, false)
}

func (g *gitdb) deleteOrFail(id string) error {
	return g.dodelete(id, true)
}

func (g *gitdb) dodelete(id string, failNotFound bool) error {

	dataDir, _, _, err := ParseId(id)
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

func (g *gitdb) delById(id string, dataset string, blockFile string, failIfNotFound bool) error {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	dataBlock := newBlock(dataset)
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
